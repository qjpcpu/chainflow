package ctrls

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/chainflow/network"
	"github.com/qjpcpu/log"
	"net/http"
	"strings"
	"time"
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

type NetworkNode struct {
	Id      string `json:"id"`
	Cluster string `json:"cluster"`
	Title   string `json:"title"`
	Root    bool   `json:"root"`
}

type NetworkEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Amount string `json:"relatedness"`
	Tx     string `json:"tx"`
}

type NetworkGraph struct {
	Nodes []NetworkNode `json:"nodes"`
	Edges []NetworkEdge `json:"edges"`
}

func QueryNetwork(c *gin.Context) {
	var err error
	var ng NetworkGraph
	for loop := true; loop; loop = false {
		contract := c.Query("contract")
		if contract == "" {
			err = errors.New("no contract")
			break
		}
		address := strings.ToLower(c.Query("address"))
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
		nodes := make(map[string]NetworkNode)
		var tmStart, tmEnd time.Time
		tmStart = time.Now()
		paths := ns.NetworkOf(strings.ToLower(address), network.Pred_Transfer, 2, direction)
		tmEnd = time.Now()
		log.Infof("query cayley cost:%v,get %v path", tmEnd.Sub(tmStart), len(paths))
		if len(paths) == 0 {
			break
		}
		var txs []db.TokenTransfer
		o := db.GetOrm()
		tmStart = time.Now()
		var fromto []interface{}
		for _, p := range paths {
			nodes[p.From] = NetworkNode{
				Id:      p.From,
				Cluster: "1",
				Title:   p.From,
				Root:    p.From == address,
			}
			nodes[p.To] = NetworkNode{
				Id:      p.To,
				Cluster: "1",
				Title:   p.To,
				Root:    p.To == address,
			}
			switch direction {
			case network.In:
				fromto = append(fromto, db.FromTo(p.To, p.From))
			default:
				fromto = append(fromto, db.FromTo(p.From, p.To))
			}
		}
		offset := 0
		step := 100
	LOOP:
		for {
			end := offset + step
			if end > len(fromto) {
				end = len(fromto)
			}
			sql := fmt.Sprintf("select * from token_transfer where id in (select max(id) from token_transfer where contract=? and from_to in (%s) group by from_to)", strings.TrimSuffix(strings.Repeat("?,", end-offset), ","))
			args := append([]interface{}{strings.ToLower(contract)}, fromto[offset:end]...)
			var ntxs []db.TokenTransfer
			if _, err = o.Raw(sql, args...).QueryRows(&ntxs); err != nil {
				break LOOP
			}
			txs = append(txs, ntxs...)
			if end >= len(fromto) {
				break LOOP
			}
			offset += step
		}
		if err != nil {
			break
		}
		tmEnd = time.Now()
		log.Infof("query db cost:%v", tmEnd.Sub(tmStart))
		for _, node := range nodes {
			ng.Nodes = append(ng.Nodes, node)
		}
		for _, tx := range txs {
			ng.Edges = append(ng.Edges, NetworkEdge{
				Source: tx.From,
				Target: tx.To,
				Tx:     tx.Tx,
				Amount: tx.Value,
			})
		}
	}
	if err != nil {
		log.Errorf("get network fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "network": ng})
	}
}
