package ctrls

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/conf"
	"github.com/qjpcpu/ethereum/etherscan"
	"github.com/qjpcpu/log"
	"net/http"
)

func AllowCORS(c *gin.Context) {
	from := c.Request.Header.Get("Origin")
	if from != "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", from)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT,PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func GetPendingTx(c *gin.Context) {
	var err error
	var txDetail etherscan.PendingTx
	owner := c.Query("address")
	for loop := true; loop; loop = false {
		if owner == "" {
			err = errors.New("no address")
			break
		}
		nonce, nerr := conf.EthConn().PendingNonceAt(context.Background(), common.HexToAddress(owner))
		if nerr != nil {
			err = nerr
			break
		}
		log.Infof("get pending nonce:%v", nonce)
		txDetail, err = etherscan.GetBlockedPendingTx(conf.EtherscanEnv, owner, nonce)
	}
	if err != nil {
		log.Errorf("get pending tx fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "tx": txDetail})
	}
}
