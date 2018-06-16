package worker

import (
	"github.com/astaxie/beego/orm"
	"github.com/qjpcpu/chainflow/cursor"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/common/redo"
	"github.com/qjpcpu/log"
	"math/big"
)

func ConsumeTransferRecord(ctx *redo.RedoCtx) {
	redisConn := db.RedisPool().Get()
	from, _ := cursor.Get(redisConn, consumeCursor)
	defer redisConn.Close()
	o := db.GetOrm()
	var records []db.TokenTransfer
	if _, err := o.QueryTable(new(db.TokenTransfer)).Filter("id__gte", from).OrderBy("id").Limit(100).All(&records); err != nil {
		log.Errorf("get record fail:%v", err)
		return
	}
	if len(records) == 0 {
		return
	}
	for _, record := range records {
		// update from
		if record.From != "0x0000000000000000000000000000000000000000" {
			var balance db.TokenBalance
			if err := o.QueryTable(new(db.TokenBalance)).
				Filter("contract", record.Contract).
				Filter("user", record.From).
				One(&balance); err != nil {
				o.Insert(&db.TokenBalance{
					Contract: record.Contract,
					User:     record.From,
					Amount:   "-" + record.Value,
					Digits:   len(record.Value),
					Block:    record.Block,
				})
			} else {
				amount, _ := new(big.Int).SetString(balance.Amount, 10)
				amount.Sub(amount, record.BigIntValue())
				o.QueryTable(new(db.TokenBalance)).
					Filter("contract", record.Contract).
					Filter("user", record.From).Update(orm.Params{
					"amount": amount.String(),
					"digits": len(amount.String()),
					"block":  record.Block,
				})
			}
		}

		// update to
		if record.To != "0x0000000000000000000000000000000000000000" {
			var balance db.TokenBalance
			if err := o.QueryTable(new(db.TokenBalance)).
				Filter("contract", record.Contract).
				Filter("user", record.To).
				One(&balance); err != nil {
				o.Insert(&db.TokenBalance{
					Contract: record.Contract,
					User:     record.To,
					Amount:   record.Value,
					Digits:   len(record.Value),
					Block:    record.Block,
				})
			} else {
				amount, _ := new(big.Int).SetString(balance.Amount, 10)
				amount.Add(amount, record.BigIntValue())
				o.QueryTable(new(db.TokenBalance)).
					Filter("contract", record.Contract).
					Filter("user", record.To).Update(orm.Params{
					"amount": amount.String(),
					"digits": len(amount.String()),
					"block":  record.Block,
				})
			}
		}

		// update graph
		TransferData{
			From: record.From,
			To:   record.To,
		}.UpdateGraph()
	}
	cursor.Set(redisConn, consumeCursor, records[len(records)-1].Id+1)
	ctx.StartNextRightNow()
}
