package ctrls

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/log"
	"net/http"
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
