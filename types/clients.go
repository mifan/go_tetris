package types

import (
	"fmt"
	"sync"

	"github.com/hprose/hprose-go/hprose"
)

// stub
type gameServerStub struct {
	Create              func(tid int, title string, bet int) error
	Delete              func(tid int) error
	Start               func(tid int) error
	SetNormalGameResult func(tid, winnerUid, bet int) error
	SetTournamentResult func(tid, winnerUid int) error
	SysText             func(text string) error
	Deactivate          func() error
}

func newGameServerStub() *gameServerStub { return new(gameServerStub) }

// num of conns
// max should be larger than 1
type numConn struct{ maxConn, currConn uint }

func newNumConn(max int) *numConn { return &numConn{maxConn: uint(max), currConn: 0} }

func (nc *numConn) Add() { nc.currConn++ }

func (nc *numConn) Release() { nc.currConn-- }

func (nc numConn) Load() float64 { return float64(nc.currConn) / float64(nc.maxConn) }

func (nc numConn) CurrConns() int { return int(nc.currConn) }

// rpc
type GameServersRpc struct {
	gameServerClient   map[string]hprose.Client
	gscMu              sync.RWMutex
	gameServerStubs    map[string]*gameServerStub
	gssMu              sync.RWMutex
	gameServerNumConns map[string]*numConn
	gsncMu             sync.RWMutex
	port               string
}

func NewGameServerRpc(port string) *GameServersRpc {
	return &GameServersRpc{
		gameServerClient:   make(map[string]hprose.Client),
		gameServerStubs:    make(map[string]*gameServerStub),
		gameServerNumConns: make(map[string]*numConn),
		port:               port,
	}
}

func (gsr *GameServersRpc) NumOfGS() int {
	gsr.gscMu.RLock()
	defer gsr.gscMu.RUnlock()
	return len(gsr.gameServerClient)
}

func (gsr *GameServersRpc) TotalConnections() int {
	gsr.gsncMu.RLock()
	defer gsr.gsncMu.RUnlock()
	total := 0
	for _, v := range gsr.gameServerNumConns {
		total += v.CurrConns()
	}
	return total
}

func (gsr *GameServersRpc) IsServerExist(ip string) bool {
	gsr.gscMu.RLock()
	defer gsr.gscMu.RUnlock()
	_, ok := gsr.gameServerClient[ip]
	return ok
}

// add a new game server
func (gsr *GameServersRpc) NewGameServer(ip string, maxConn int) {
	addStub := func() {
		gsr.gssMu.Lock()
		defer gsr.gssMu.Unlock()
		gsr.gameServerStubs[ip] = newGameServerStub()
	}
	addClient := func() {
		gsr.gscMu.Lock()
		defer gsr.gscMu.Unlock()
		gsr.gameServerClient[ip] = hprose.NewHttpClient(fmt.Sprintf("http://%s:%s/", ip, gsr.port))
		gsr.gameServerClient[ip].UseService(gsr.gameServerStubs[ip])
		gsr.gameServerClient[ip].SetKeepAlive(true)
	}
	addNc := func() {
		gsr.gsncMu.Lock()
		defer gsr.gsncMu.Unlock()
		gsr.gameServerNumConns[ip] = newNumConn(maxConn)
	}

	addStub()
	addClient()
	addNc()
}

// delete a game server
func (gsr *GameServersRpc) Delete(ip string) {
	delStub := func() {
		gsr.gssMu.Lock()
		defer gsr.gssMu.Unlock()
		delete(gsr.gameServerStubs, ip)
	}
	delClient := func() {
		gsr.gscMu.Lock()
		defer gsr.gscMu.Unlock()
		delete(gsr.gameServerClient, ip)
	}
	delNc := func() {
		gsr.gsncMu.Lock()
		defer gsr.gsncMu.Unlock()
		delete(gsr.gameServerNumConns, ip)
	}

	delStub()
	delClient()
	delNc()

}

// check if the ip exist before invoke this
func (gsr *GameServersRpc) AddConn(ip string) {
	gsr.gsncMu.RLock()
	defer gsr.gsncMu.RUnlock()
	gsr.gameServerNumConns[ip].Add()
}

// check if the ip exist before invoke this
func (gsr *GameServersRpc) ReleaseConn(ip string) {
	gsr.gsncMu.RLock()
	defer gsr.gsncMu.RUnlock()
	gsr.gameServerNumConns[ip].Release()
}

// best game server
func (gsr *GameServersRpc) BestServer() (ip string) {
	gsr.gsncMu.RLock()
	defer gsr.gsncMu.RUnlock()
	var usage float64 = 2
	for i, v := range gsr.gameServerNumConns {
		if l := v.Load(); l < usage {
			usage = l
			ip = i
		}
	}
	return
}

// get stub
func (gsr *GameServersRpc) GetStub(ip string) *gameServerStub {
	gsr.gssMu.RLock()
	defer gsr.gssMu.RUnlock()
	return gsr.gameServerStubs[ip]
}

// deactivate all
func (gsr *GameServersRpc) DeactivateAll() error {
	gsr.gssMu.RLock()
	defer gsr.gssMu.RUnlock()
	for _, stub := range gsr.gameServerStubs {
		if err := stub.Deactivate(); err != nil {
			return err
		}
	}
	return nil
}
