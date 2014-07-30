// private server is rpc server provided for ease of game server implementation
package main

import (
	"net/http"
	"reflect"

	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

var (
	httpPrivServer = hprose.NewHttpService()
)

type (
	privStub   struct{}
	privSe     struct{}
	privFilter struct{}
)

func (privSe) OnBeforeInvoke(funcN string, params []reflect.Value, isSimple bool, ctx interface{}) {
	log.Info("invoker from ip %v, invoke function %v, params %v", utils.GetIp(ctx), funcN, params)
	if funcN != "Register" && !clients.IsServerExist(utils.GetIp(ctx)) {
		panic("game server should first register, otherwise it is not allowed to invoke functions in auth server")
	}
	switch funcN {
	case "Quit":
		clients.ReleaseConn(utils.GetIp(ctx))
	case "Apply", "Join", "ObTournament":
		clients.AddConn(utils.GetIp(ctx))
	}
}

func (privSe) OnAfterInvoke(string, []reflect.Value, bool, []reflect.Value, interface{}) {}

func (privSe) OnSendError(error, interface{}) {}

func (privFilter) InputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Decrypt(data, privKey)
}

func (privFilter) OutputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Encrypt(data, privKey)
}

func initPrivServer() {
	httpPrivServer.AddMethods(privStub{})
	httpPrivServer.ServiceEvent = privSe{}
	httpPrivServer.SetFilter(privFilter{})
	go servePrivHttp()
}

func servePrivHttp() {
	if err := http.ListenAndServe(":"+privRpcPort, httpPrivServer); err != nil {
		panic(err)
	}
}
