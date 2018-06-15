package ctrls

import (
	"errors"
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
	Type    string `json:"type"`
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
	var ng NetworkGraph = NetworkGraph{
		Nodes: make([]NetworkNode, 0),
		Edges: make([]NetworkEdge, 0),
	}
	contract := strings.ToLower(c.Query("contract"))
	address := strings.ToLower(c.Query("address"))
	for loop := true; loop; loop = false {
		if contract == "" {
			err = errors.New("no contract")
			break
		}
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

		rootType := func(isRoot bool) string {
			if isRoot {
				return "philosopher"
			} else {
				return ""
			}
		}
		shortName := func(longname string) string {
			if len(longname) > 5 {
				return string(longname[:5]) + "..."
			} else {
				return longname
			}
		}
		for i, p := range paths {
			nodes[p.From] = NetworkNode{
				Id:      p.From,
				Cluster: "1",
				Title:   shortName(p.From),
				Root:    p.From == address,
				Type:    rootType(p.From == address),
			}
			nodes[p.To] = NetworkNode{
				Id:      p.To,
				Cluster: "1",
				Title:   shortName(p.To),
				Root:    p.To == address,
				Type:    rootType(p.To == address),
			}
			switch direction {
			case network.In:
				ng.Edges = append(ng.Edges, NetworkEdge{
					Source: p.To,
					Target: p.From,
				})
			default:
				ng.Edges = append(ng.Edges, NetworkEdge{
					Source: p.From,
					Target: p.To,
				})
			}
			if i > 50 {
				break
			}
		}

		for _, node := range nodes {
			ng.Nodes = append(ng.Nodes, node)
		}
	}
	if err != nil {
		log.Errorf("get network fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "network": ng, "contract": contract})
	}
}

func QueryHotTx(c *gin.Context) {
	contract := strings.ToLower(c.Query("contract"))
	from := strings.ToLower(c.Query("from"))
	to := strings.ToLower(c.Query("to"))
	var err error
	var max, hot db.TokenTransfer
	for loop := true; loop; loop = false {
		if contract == "" || from == "" || to == "" {
			err = errors.New("params err")
			break
		}
		o := db.GetOrm()
		if err = o.QueryTable(new(db.TokenTransfer)).
			Filter("contract", contract).
			Filter("from", from).
			Filter("to", to).
			OrderBy("-block").
			Limit(1).
			One(&hot); err != nil {
			break
		}
		if err = o.QueryTable(new(db.TokenTransfer)).
			Filter("contract", contract).
			Filter("from", from).
			Filter("to", to).
			OrderBy("-digits", "-value").
			Limit(1).
			One(&max); err != nil {
			break
		}
	}
	if err != nil {
		log.Errorf("get hot tx fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "hot": hot, "max": max})
	}
}
