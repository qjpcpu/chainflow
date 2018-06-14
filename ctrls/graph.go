package ctrls

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/chainflow/network"
	"github.com/qjpcpu/log"
	"net/http"
	"strings"
)

func GetTopBalance(c *gin.Context) {
	var err error
	var res []struct {
		User   string `json:"user"`
		Amount int64  `json:"amount"`
	}
	for loop := true; loop; loop = false {
		contract := c.Query("contract")
		if contract == "" {
			err = errors.New("no contract")
			break
		}
		o := db.GetOrm()
		top := 10
		if _, err = o.Raw("select user,amount from token_balance where contract=? order by amount desc limit ?", contract, top).QueryRows(&res); err != nil {
			break
		}
	}
	if err != nil {
		log.Errorf("get top balance fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "top": res})
	}
}

func QueryNetwork(c *gin.Context) {
	var err error
	var txs []db.TokenTransfer
	for loop := true; loop; loop = false {
		contract := c.Query("contract")
		if contract == "" {
			err = errors.New("no contract")
			break
		}
		address := c.Query("address")
		if address == "" {
			err = errors.New("no address")
			break
		}
		var direction network.Direction
		ns, nerr := network.GetGraphOfToken(contract)
		if nerr != nil {
			err = nerr
			break
		}
		defer ns.Close()
		switch c.Query("direction") {
		case "in":
			direction = network.In
		case "out":
			direction = network.Out
		default:
			direction = network.Out
		}
		paths := ns.NetworkOf(strings.ToLower(address), network.Pred_Transfer, 5, direction)
		txs = make([]db.TokenTransfer, len(paths))
		o := db.GetOrm()
		for i, p := range paths {
			switch direction {
			case network.In:
				o.QueryTable(new(db.TokenTransfer)).
					Filter("from", p.To).
					Filter("to", p.From).
					OrderBy("-block").
					Limit(1).
					One(&txs[i])
			default:
				o.QueryTable(new(db.TokenTransfer)).
					Filter("from", p.From).
					Filter("to", p.To).
					OrderBy("-block").
					Limit(1).
					One(&txs[i])
			}
		}
	}
	if err != nil {
		log.Errorf("get network fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "txs": txs})
	}
}
