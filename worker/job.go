package worker

import (
	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/qjpcpu/chainflow/conf"
	"github.com/qjpcpu/chainflow/cursor"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/chainflow/network"
	"github.com/qjpcpu/common/redo"
	"github.com/qjpcpu/ethereum/events"
	"github.com/qjpcpu/log"
	"math/big"
	"strings"
	"time"
)

const (
	erc20Cursor   = "ERC20"
	consumeCursor = "Consumer"
	TransferABI   = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
)

type TransferData struct {
	Tx       string
	Block    uint64
	Contract string

	From  string   `json:"from"`
	To    string   `json:"to"`
	Value *big.Int `json:"value"`
}

func FetchERC20CoinTransfer() error {
	log.Infof("监听ERC20 Transfer日志")
	redisConn := db.RedisPool().Get()
	from, err := cursor.Get(redisConn, erc20Cursor)
	redisConn.Close()
	if err != nil {
		return err
	}
	if from == 0 {
		from = 1
	}

	dataCh, errCh := make(chan events.Event, 1000), make(chan error, 1)
	progressCh := make(chan events.Progress, 1)
	builder := events.NewScanBuilder()
	rep1, err := builder.SetClient(conf.EthConn()).
		SetContract(common.Address{}, TransferABI, "Transfer").
		SetBlockMargin(2).
		SetFrom(from).
		SetGracefullExit(true).
		SetProgressChan(progressCh).
		SetInterval(time.Second*20).
		SetDataChan(dataCh, errCh).
		BuildAndRun()
	if err != nil {
		log.Errorf("fail to start fetcher:%v", err)
		return err
	}

	rep2 := redo.PerformSafe(ConsumeTransferRecord, time.Second*5)
	rep := rep1.Concat(rep2)

	done := rep.WaitChan()
	updateCursorFunc := func(bn uint64) {
		redis_conn := db.RedisPool().Get()
		defer redis_conn.Close()
		cursor.Set(redis_conn, erc20Cursor, bn+1)
	}
	o := db.GetOrm()
	for {
		select {
		case data := <-dataCh:
			td := getTransferData(data)
			td.Save(o)
			//		td.UpdateGraph()
		case err1 := <-errCh:
			log.Errorf("fetcher receive err:%v", err1)
		case progress := <-progressCh:
			log.Debugf("sync to block:%v", progress.To)
			updateCursorFunc(progress.To)
		case <-done:
			log.Info("exit fetch")
			return nil
		}
	}
}

func getTransferData(data events.Event) TransferData {
	td := TransferData{}
	data.Data.Unmarshal(&td)
	td.Contract = data.Address.Hex()
	td.Block = data.BlockNumber
	td.Tx = data.TxHash.Hex()
	return td
}

func (td TransferData) UpdateGraph() error {
	from := strings.ToLower(td.From)
	to := strings.ToLower(td.To)
	if from == "0x0000000000000000000000000000000000000000" || to == "0x0000000000000000000000000000000000000000" {
		return nil
	}
	graph, err := network.GetGraphOfToken(td.Contract)
	if err != nil {
		return err
	}
	defer graph.Close()
	graph.AddQuadString(
		from,
		network.Pred_Transfer,
		to,
	)
	return nil
}

func (td TransferData) Save(o orm.Ormer) error {
	contract_addr := strings.ToLower(td.Contract)
	from := strings.ToLower(td.From)
	to := strings.ToLower(td.To)
	if _, err := o.Insert(&db.TokenTransfer{
		Contract: contract_addr,
		From:     from,
		To:       to,
		FromTo:   db.FromTo(td.From, td.To),
		Value:    td.Value.String(),
		Digits:   len(td.Value.String()),
		Tx:       strings.ToLower(td.Tx),
		Block:    td.Block,
	}); err != nil {
		return err
	}
	// update from
	// if from != "0x0000000000000000000000000000000000000000" {
	// 	var balance db.TokenBalance
	// 	if err := o.QueryTable(new(db.TokenBalance)).
	// 		Filter("contract", contract_addr).
	// 		Filter("user", from).
	// 		One(&balance); err != nil {
	// 		o.Insert(&db.TokenBalance{
	// 			Contract: contract_addr,
	// 			User:     from,
	// 			Amount:   "-" + td.Value.String(),
	// 			Digits:   len(td.Value.String()),
	// 			Block:    td.Block,
	// 		})
	// 	} else {
	// 		amount, _ := new(big.Int).SetString(balance.Amount, 10)
	// 		amount.Sub(amount, td.Value)
	// 		o.QueryTable(new(db.TokenBalance)).
	// 			Filter("contract", contract_addr).
	// 			Filter("user", from).Update(orm.Params{
	// 			"amount": amount.String(),
	// 			"digits": len(amount.String()),
	// 			"block":  td.Block,
	// 		})
	// 	}
	// }

	// update to
	// if to != "0x0000000000000000000000000000000000000000" {
	// 	var balance db.TokenBalance
	// 	if err := o.QueryTable(new(db.TokenBalance)).
	// 		Filter("contract", contract_addr).
	// 		Filter("user", to).
	// 		One(&balance); err != nil {
	// 		o.Insert(&db.TokenBalance{
	// 			Contract: contract_addr,
	// 			User:     to,
	// 			Amount:   td.Value.String(),
	// 			Digits:   len(td.Value.String()),
	// 			Block:    td.Block,
	// 		})
	// 	} else {
	// 		amount, _ := new(big.Int).SetString(balance.Amount, 10)
	// 		amount.Add(amount, td.Value)
	// 		o.QueryTable(new(db.TokenBalance)).
	// 			Filter("contract", contract_addr).
	// 			Filter("user", from).Update(orm.Params{
	// 			"amount": amount.String(),
	// 			"digits": len(amount.String()),
	// 			"block":  td.Block,
	// 		})
	// 	}
	// }
	return nil
}
