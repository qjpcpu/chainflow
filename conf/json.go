package conf

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/qjpcpu/ethereum/etherscan"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
)

var (
	config       Configuration
	conn         *ethclient.Client
	EtherscanEnv etherscan.Env
)

type Configuration struct {
	MysqlConn   string
	Redisconn   string
	RedisDb     string
	RedisPwd    string
	LogDir      string
	GraphDir    string
	LogLevel    string
	EthNodePath string
	Listen      string
}

func Get() *Configuration {
	return &config
}

// obj must be pointer
func LoadJson(config_file string, obj interface{}) error {
	file, err := os.Open(config_file)
	if err != nil {
		return err
	}
	defer file.Close()

	config_str, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(config_str, obj)
}

func (cfg *Configuration) InitEthClients() error {
	var err error
	conn, err = ethclient.Dial(cfg.EthNodePath)
	if err != nil {
		return err
	}
	if strings.Contains(cfg.EthNodePath, "ropsten") {
		EtherscanEnv = etherscan.Ropsten
	} else {
		EtherscanEnv = etherscan.Online
	}
	return nil
}

func EthConn() *ethclient.Client {
	return conn
}

func AsEth(num *big.Int) float64 {
	one_eth := big.NewFloat(1000000000000000000)
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(num), one_eth).Float64()
	return f
}
