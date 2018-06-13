package worker

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/qjpcpu/chainflow/conf"
	"github.com/qjpcpu/chainflow/cursor"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/ethereum/events"
	"github.com/qjpcpu/log"
	"math/big"
	"strings"
	"time"
)

const (
	erc20Cursor = "ERC20"
	TransferABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
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
	rep, err := builder.SetClient(conf.EthConn()).
		SetContract(common.Address{}, TransferABI, "Transfer").
		SetBlockMargin(2).
		SetFrom(from).
		SetStep(10000).
		SetGracefullExit(true).
		SetProgressChan(progressCh).
		SetInterval(time.Second*20).
		SetDataChan(dataCh, errCh).
		BuildAndRun()
	if err != nil {
		log.Errorf("fail to start fetcher:%v", err)
		return err
	}
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
			getTransferData(data).Save(o)
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

func (td TransferData) Save(o orm.Ormer) error {
	return db.ExecTransaction(o, func() error {
		if _, err := o.Insert(&db.TokenTransfer{
			Contract: strings.ToLower(td.Contract),
			From:     strings.ToLower(td.From),
			To:       strings.ToLower(td.To),
			Value:    td.Value.Uint64(),
			Tx:       strings.ToLower(td.Tx),
			Block:    td.Block,
		}); err != nil {
			return err
		}
		contract_addr := strings.ToLower(td.Contract)
		from := strings.ToLower(td.From)
		to := strings.ToLower(td.To)
		sql := fmt.Sprintf(`insert ignore into %s(contract,user) values(?,?)`, new(db.TokenBalance).TableName())
		o.Raw(sql, contract_addr, from).Exec()
		o.Raw(sql, contract_addr, to).Exec()
		if _, err := o.QueryTable(new(db.TokenBalance)).
			Filter("contract", contract_addr).
			Filter("user", to).Update(orm.Params{
			"amount": orm.ColValue(orm.ColAdd, td.Value.Uint64()),
			"block":  td.Block,
		}); err != nil {
			return err
		}
		if _, err := o.QueryTable(new(db.TokenBalance)).
			Filter("contract", contract_addr).
			Filter("user", from).Update(orm.Params{
			"amount": orm.ColValue(orm.ColMinus, td.Value.Uint64()),
			"block":  td.Block,
		}); err != nil {
			return err
		}
		return nil
	})
}
