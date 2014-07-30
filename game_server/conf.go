package main

import (
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego/config"
	"github.com/gogames/go_tetris/utils"
)

var (
	conf config.ConfigContainer

	tokenEncryptKey    string
	logPath            string
	authServerIp       string
	authServerRpcPort  string
	gameServerRpcPort  string
	gameServerSockPort string
	maxConn            int
	privKey            []byte
)

func initConf() {
	var err error
	conf, err = config.NewConfig("json", *confPath)
	if err != nil {
		fmt.Printf("can not read configuration: %v", err)
		os.Exit(1)
	}

	tokenEncryptKey = conf.String("tokenEncryptKey")
	logPath = conf.String("log")
	authServerIp = conf.String("authServerIp")
	authServerRpcPort = conf.String("authServerRpcPort")
	gameServerRpcPort = conf.String("gameServerRpcPort")
	gameServerSockPort = conf.String("gameServerSockPort")
	privKeyString := conf.String("privKey")
	maxConn, err = conf.Int("maxConn")
	if err != nil {
		log.Critical("can not get maxConn: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}

	utils.CheckEmptyConf(tokenEncryptKey, logPath, authServerIp,
		authServerRpcPort, gameServerRpcPort, gameServerSockPort, maxConn)

	utils.SetTokenKey([]byte(tokenEncryptKey))
	privKey = []byte(privKeyString)
}
