package main

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/logs"
)

var log = logs.NewLogger(10000)

func initLogger() {
	if err := log.SetLogger("file", fmt.Sprintf(`{"filename":"%s"}`, logPath)); err != nil {
		os.Exit(1)
	}
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(2)
	log.Info("initialize logger...")
}
