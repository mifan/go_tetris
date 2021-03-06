package main

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

var httpPubServer = hprose.NewHttpService()

var pubServerEnable = true

type (
	pubStub struct{}
	pubSe   struct{}
)

func (pubSe) OnBeforeInvoke(fName string, params []reflect.Value, isSimple bool, ctx interface{}) {
	log.Info("%v", time.Now())
	log.Info(utils.HproseLog(fName, params, ctx))
	if !pubServerEnable {
		panic("we are closing the server, not accept request at the moment")
	}

	session.CreateSession(ctx)
}

func (pubSe) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {
	log.Info("%v", time.Now())
}

func (pubSe) OnSendError(error, interface{}) {}

func initPubServer() {
	httpPubServer.DebugEnabled = true
	httpPubServer.AddMethods(pubStub{})
	httpPubServer.ServiceEvent = pubSe{}
	httpPubServer.CrossDomainEnabled = true
	// httpServer.AddAccessControlAllowOrigin()
	go servePubHttp()
}

func servePubHttp() {
	if err := http.ListenAndServe(":"+pubRpcPort, httpPubServer); err != nil {
		panic(err)
	}
}
