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

/*
CREATE TABLE `token_transfer` (
  `id` bigint(11) unsigned NOT NULL AUTO_INCREMENT,
  `contract` char(42) DEFAULT '' COMMENT 'ERC20兼容合约地址',
  `from` char(42) DEFAULT '' COMMENT 'token发送者',
  `to` char(42) DEFAULT '' COMMENT 'token接收者',
  `value` varchar(512) DEFAULT '0' COMMENT 'token数量',
  `digits` bigint(20) DEFAULT '0',
  `tx` char(66) DEFAULT '' COMMENT '交易',
  `block` bigint(20) DEFAULT '0' COMMENT '区块',
  `from_to` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `contract` (`contract`),
  KEY `from_to` (`from_to`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
*/
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

/*
CREATE TABLE `token_balance` (
  `id` bigint(11) unsigned NOT NULL AUTO_INCREMENT,
  `contract` char(42) DEFAULT '' COMMENT 'ERC20兼容合约地址',
  `user` char(42) DEFAULT '' COMMENT '用户地址',
  `amount` varchar(512) DEFAULT '0' COMMENT '当前余额',
  `digits` bigint(20) DEFAULT '0' COMMENT '十进制位数',
  `block` bigint(11) DEFAULT '0' COMMENT '当前区块',
  PRIMARY KEY (`id`),
  UNIQUE KEY `contract` (`contract`,`user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
*/
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
