/*
	rpc server for auth server to control it
*/
package main

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

var httpServer = hprose.NewHttpService()

type stub struct{}

type se struct{}

func (se) OnBeforeInvoke(funcName string, params []reflect.Value, isSimple bool, ctx interface{}) {
	if !isServerActive() {
		panic("game server is closing, do not accept any request...")
	}

	log.Info(utils.HproseLog(funcName, params, ctx))

	if ip := utils.GetIp(ctx); ip != authServerIp {
		panic(fmt.Errorf("do not accept request from this client: %v", ip))
	}
}

func (se) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {
}

func (se) OnSendError(error, interface{}) {
}

func initRpcServer() {
	httpServer.AddMethods(stub{})
	httpServer.ServiceEvent = se{}
	go serveHttp()
}

func serveHttp() {
	if err := http.ListenAndServe(fmt.Sprintf(":%s", gameServerRpcPort), httpServer); err != nil {
		panic(err)
	}
}
