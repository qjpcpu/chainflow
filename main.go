package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qjpcpu/chainflow/conf"
	"github.com/qjpcpu/chainflow/ctrls"
	"github.com/qjpcpu/chainflow/db"
	"github.com/qjpcpu/chainflow/network"
	"github.com/qjpcpu/chainflow/worker"
	"github.com/qjpcpu/log"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

func main() {
	app := cli.NewApp()
	app.Name = "chainflow"
	app.Usage = "Chain Flow"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "JasonQu",
			Email: "qjpcpu@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "config file",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "http",
			Usage:  "start web service",
			Before: initCmd,
			Action: func(c *cli.Context) error {
				startServer()
				return nil
			},
		},
		{
			Name:   "sync",
			Usage:  "sync erc20 transfer",
			Before: initCmd,
			Action: func(c *cli.Context) error {
				return worker.FetchERC20CoinTransfer()
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("no command to run")
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func startServer() {
	cfg := conf.Get()
	router := gin.Default()

	router.Use(ctrls.AllowCORS)
	txRouter := router.Group("/tx")
	{
		txRouter.GET("/pending", ctrls.GetPendingTx)
		txRouter.POST("/calc_hash", ctrls.CalcTxHash)
		txRouter.POST("/send_raw_tx", ctrls.SendRawTx)
	}

	tokenRouter := router.Group("/token")
	{
		tokenRouter.GET("/topbalance", ctrls.GetTopBalance)
		tokenRouter.GET("/network", ctrls.QueryNetwork)
	}

	router.Run(cfg.Listen)
}

func initCmd(c *cli.Context) error {
	config_file := c.GlobalString("c")
	if config_file == "" {
		return errors.New("no config file")
	}
	cfg := conf.Get()
	if err := conf.LoadJson(config_file, cfg); err != nil {
		fmt.Printf("load config err:%v\n", err)
		return err
	}

	// init log
	log.InitLog(log.LogOption{
		LogFile: filepath.Join(cfg.LogDir, c.Command.Name+".log"),
		Level:   log.ParseLogLevel(cfg.LogLevel),
		Format:  "%{level}: [%{time:2006-01-02 15:04:05.000}][%{shortfile}][%{message}]",
	})
	if err := cfg.InitEthClients(); err != nil {
		return err
	}
	db.InitRedis(cfg.Redisconn, cfg.RedisDb, cfg.RedisPwd)
	if err := db.InitOrm(cfg.MysqlConn); err != nil {
		return err
	}
	db.SetDBLog(filepath.Join(cfg.LogDir, "mysql.log"))
	network.SetGraphDir(cfg.GraphDir)
	return nil
}
