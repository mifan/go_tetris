package main

import (
	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

type authServer struct {
	Register            func(maxConn int) error
	Unregister          func() error
	Join                func(tid, uid int, isOb bool) error
	ObTournament        func(tid, uid int) error
	SwitchReady         func(tid, uid int) error
	Quit                func(tid, uid int, isTournament bool) error
	SetNormalGameResult func(tid, winner, loser int) error
	SetTournamentResult func(tid, winner, loser int) error
	Apply               func(uid int) (int, error)
	Allocate            func(uid int) (int, error)
}

type authFilter struct{}

func (authFilter) InputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Decrypt(data, privKey)
}

func (authFilter) OutputFilter(data []byte, ctx interface{}) []byte {
	return xxtea.Encrypt(data, privKey)
}

var (
	client         hprose.Client
	authServerStub = new(authServer)
)

func initRpcClient() {
	uri := "http://" + authServerIp + ":" + authServerRpcPort + "/"
	client = hprose.NewHttpClient(uri)
	client.UseService(&authServerStub)
	client.SetKeepAlive(true)
	client.SetFilter(authFilter{})
}
