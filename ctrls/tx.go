package ctrls

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/conf"
	"github.com/qjpcpu/ethereum/etherscan"
	"github.com/qjpcpu/log"
	"net/http"
	"strings"
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

type RawTxInfo struct {
	Nonce    string `json:"nonce"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasLimit string `json:"gas"`
	GasPrice string `json:"gas_price"`
	Data     string `json:"data"` // 0x...
	Sign     string `json:"sign"` // 0x...
}

func CalcTxHash(c *gin.Context) {
	var err error
	var tx *types.Transaction
	var hash common.Hash
	for loop := true; loop; loop = false {
		var rawTx RawTxInfo
		if err = c.ShouldBindJSON(&rawTx); err != nil {
			break
		}
		log.Infof("get raw tx:%+v", rawTx)
		if tx, err = rawTx.ToTransaction(); err != nil {
			log.Errorf("parse tx fail:%v", err)
			break
		}
		hash = types.HomesteadSigner{}.Hash(tx)
	}
	if err != nil {
		log.Errorf("calc tx fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "hash": hash.Hex()})
	}
}

func SendRawTx(c *gin.Context) {
	var err error
	var tx *types.Transaction
	var hash common.Hash
	for loop := true; loop; loop = false {
		var rawTx RawTxInfo
		if err = c.ShouldBindJSON(&rawTx); err != nil {
			break
		}
		log.Infof("get raw tx:%+v", rawTx)
		if tx, err = rawTx.ToSignedTransaction(); err != nil {
			log.Errorf("parse tx fail:%v", err)
			break
		}
		if err = conf.EthConn().SendTransaction(context.Background(), tx); err != nil {
			log.Errorf("send raw tx fail:%v", err)
			break
		}
		hash = tx.Hash()
	}
	if err != nil {
		log.Errorf("calc tx fail:%v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "hash": hash.Hex()})
	}
}

func (rt RawTxInfo) ToTransaction() (*types.Transaction, error) {
	if rt.Nonce == "" {
		return nil, errors.New("no nonce")
	}
	nonce, err := hexutil.DecodeUint64(rt.Nonce)
	if err != nil {
		return nil, err
	}
	if rt.To == "" {
		return nil, errors.New("no to ")
	}
	gas, err := hexutil.DecodeUint64(rt.GasLimit)
	if err != nil || gas <= 0 {
		return nil, err
	}
	rt.Data = strings.TrimPrefix(rt.Data, "0x")
	value, err := hexutil.DecodeBig(rt.Value)
	if err != nil {
		return nil, err
	}
	if rt.GasPrice == "" {
		return nil, errors.New("no gas price")
	}
	gp, err := hexutil.DecodeBig(rt.GasPrice)
	if err != nil {
		return nil, err
	}
	return types.NewTransaction(
		nonce,
		common.HexToAddress(rt.To),
		value,
		gas,
		gp,
		common.Hex2Bytes(rt.Data),
	), nil
}

func (rt RawTxInfo) ToSignedTransaction() (*types.Transaction, error) {
	tx, err := rt.ToTransaction()
	if err != nil {
		return nil, err
	}
	if rt.Sign == "" {
		return nil, errors.New("no sign")
	}
	rt.Sign = strings.TrimPrefix(rt.Sign, "0x")
	sign := common.Hex2Bytes(rt.Sign)
	if len(sign) != 65 {
		return nil, errors.New("bad sign")
	}
	if sign[64] != 27 && sign[64] != 28 {
		return nil, errors.New("invalid Ethereum signature (V is not 27 or 28)")
	}
	sign[64] -= 27
	signer := types.HomesteadSigner{}
	signedTx, err := tx.WithSignature(signer, sign)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}
