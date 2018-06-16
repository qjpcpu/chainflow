package db

import (
	"crypto/md5"
	"fmt"
	"math/big"
	"strings"
)

func init() {
	RegisterModel(
		new(TokenTransfer),
		new(TokenBalance),
	)
}

type TokenTransfer struct {
	Id       uint64 `orm:"pk;auto" json:"-"`
	Contract string `json:"contract"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Digits   int    `json:"digits"`
	Tx       string `json:"tx"`
	Block    uint64 `json:"block"`
	FromTo   string `json:"-"`
}

func (t *TokenTransfer) TableName() string {
	return "token_transfer"
}

type TokenBalance struct {
	Id       uint64 `orm:"pk;auto"`
	Contract string
	User     string
	Amount   string
	Digits   int
	Block    uint64
}

func (t *TokenBalance) TableName() string {
	return "token_balance"
}

func FromTo(from, to string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.ToLower(from)+strings.ToLower(to))))
}

func (tt *TokenTransfer) BigIntValue() *big.Int {
	s, _ := new(big.Int).SetString(tt.Value, 10)
	return s
}
