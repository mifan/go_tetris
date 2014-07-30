package types

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/timer"
)

var (
	ErrExisted  = fmt.Errorf("the table is already exist")
	ErrNotExist = fmt.Errorf("找不到该桌子.")
	ErrRoomFull = fmt.Errorf("桌子已满, 无法加入游戏.")
)

type Tables struct {
	Tables  map[int]*Table
	mu      sync.RWMutex
	expires map[int]*Table
}

func NewTables() *Tables {
	ts := &Tables{
		Tables:  make(map[int]*Table),
		expires: make(map[int]*Table),
	}
	return ts.init()
}

func (ts *Tables) init() *Tables {
	go findExpires(ts)
	return ts
}

func findExpires(ts *Tables) {
	for {
		func() {
			ts.mu.Lock()
			defer ts.mu.Unlock()
			for tid, t := range ts.Tables {
				if t.Expire() {
					ts.expires[tid] = t
				}
			}
		}()
		time.Sleep(30 * time.Second)
	}
}

func (ts *Tables) GetExpireTables() map[int]*Table {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.expires
}

func (ts *Tables) ReleaseExpireTable(tid int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.Tables, tid)
	delete(ts.expires, tid)
}

// for hprose
func (ts Tables) Wrap() []map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	res := make([]map[string]interface{}, 0)
	for _, t := range ts.Tables {
		res = append(res, t.WrapTable())
	}
	return res
}

func (ts Tables) MarshalJSON() ([]byte, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	var res = make(map[string]*Table)
	for id, t := range ts.Tables {
		res[fmt.Sprintf("%d", id)] = t
	}
	return json.Marshal(res)
}

// get all connections in all tables
func (ts *Tables) GetAllConnsInAllTables() []*net.TCPConn {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	conns := make([]*net.TCPConn, 0)
	for _, table := range ts.Tables {
		for _, c := range table.GetAllConns() {
			if c != nil {
				conns = append(conns, c)
			}
		}
	}
	return conns
}

// check if the Table exist
func (ts *Tables) IsTableExist(id int) bool {
	return ts.GetTableById(id) != nil
}

// get a Table
func (ts *Tables) GetTableById(id int) *Table {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.Tables[id]
}

// delete the Table
func (ts *Tables) DelTable(id int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.Tables, id)
}

// create a new Table
func (ts *Tables) NewTable(id int, title, host string, bet int) error {
	if ts.IsTableExist(id) {
		return ErrExisted
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Tables[id] = newTable(id, title, host, bet)
	return nil
}

// join a Table
func (ts *Tables) JoinTable(id int, u *User, isOb bool) error {
	t := ts.GetTableById(id)
	if t == nil {
		return ErrNotExist
	}
	if isOb {
		t.JoinOB(u)
		return nil
	}
	return t.Join(u)
}

// number of tables
func (ts *Tables) Length() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.Tables)
}

const (
	statWaiting = "等待"
	statInGame  = "已开始"
)

// basic table info
type tableInfo struct {
	TId    int    `json:"table_id"`
	TTitle string `json:"table_title"`
	TStat  string `json:"table_status"`
	TBet   int    `json:"table_bet"`
	THost  string `json:"table_host"`
}

func (ti tableInfo) IsStart() bool {
	return ti.TStat == statInGame
}

func (ti tableInfo) GetHost() string {
	return ti.THost
}

func (ti tableInfo) GetIp() string {
	return strings.Split(ti.THost, ":")[0]
}

func (ti tableInfo) IsGamble() bool {
	return ti.TBet > 0
}

const (
	zoneHeight            = 20
	zoneWidth             = 10
	defaultNumOfNextPiece = 5
	defaultInterval       = 1000
)

const (
	GameoverNormal = iota
	Gameover1pQuit
	Gameover2pQuit
)

// table
type Table struct {
	mu sync.Mutex
	// basic table information
	tableInfo
	// observers
	obs *obs
	// player 1p, 2p
	_1p, _2p *User
	// game 1p, 2p
	g1p, g2p *tetris.Game
	// 1p 2p ready ?
	ready1p, ready2p bool
	startTime        int64
	// timer
	timer               *timer.Timer
	remainedSeconds     int
	RemainedSecondsChan chan int
	// game over
	GameoverChan chan int
}

func newTable(id int, title, host string, bet int) *Table {
	return &Table{
		tableInfo: tableInfo{
			TId:    id,
			TTitle: title,
			TStat:  statWaiting,
			TBet:   bet,
			THost:  host,
		},
		obs:                 NewObs(),
		startTime:           time.Now().Unix(),
		remainedSeconds:     120,
		timer:               timer.NewTimer(1000),
		RemainedSecondsChan: make(chan int, 1<<3),
		GameoverChan:        make(chan int, 1<<3),
	}
}

func (t *Table) UpdateTimer() {
	for {
		t.timer.Wait()
		if b := func() bool {
			t.mu.Lock()
			defer t.mu.Unlock()
			t.remainedSeconds--
			if t.remainedSeconds <= 0 {
				t.GameoverChan <- GameoverNormal
				return true
			}
			t.RemainedSecondsChan <- t.remainedSeconds
			return false
		}(); b {
			return
		}
	}
}

func (t *Table) WrapTable() map[string]interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	return map[string]interface{}{
		"table_bet":      t.TBet,
		"table_id":       t.TId,
		"table_host":     t.THost,
		"table_status":   t.TStat,
		"table_title":    t.TTitle,
		"table_1p":       t._1p,
		"table_2p":       t._2p,
		"table_1p_ready": t.ready1p,
		"table_2p_ready": t.ready2p,
		"table_obs":      t.obs.Wrap(),
	}
}

// table json
func (t *Table) MarshalJSON() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return json.Marshal(map[string]interface{}{
		"info":      t.tableInfo,
		"observers": t.obs,
		"1p":        t._1p,
		"2p":        t._2p,
		"1p_ready":  t.ready1p,
		"2p_ready":  t.ready2p,
	})
}

// start the game, only used on game server
func (t *Table) StartGame() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.g1p, _ = tetris.NewGame(zoneHeight, zoneWidth, defaultNumOfNextPiece, defaultInterval)
	t.g2p, _ = tetris.NewGame(zoneHeight, zoneWidth, defaultNumOfNextPiece, defaultInterval)
	t.timer.Start()
	t.g1p.Start()
	t.g2p.Start()
	t.startTime = time.Now().Unix()
}

// stop the game, only used on game server
func (t *Table) StopGame() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timer.Pause()
	t.timer.Reset()
	t.g1p.Stop()
	t.g2p.Stop()
	t.TStat = statWaiting
	t.startTime = time.Now().Unix()
}

// reset the table
func (t *Table) ResetTable() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.g1p = nil
	t.g2p = nil
	t.ready1p = false
	t.ready2p = false
	t.remainedSeconds = 120
	t.TStat = statWaiting
}

// set ready
func (t *Table) SwitchReady(uid int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if uid < 0 {
		return
	}
	switch uid {
	case t._1p.GetUid():
		t.ready1p = !t.ready1p
	case t._2p.GetUid():
		t.ready2p = !t.ready2p
	}
}

// check if the table should expire
func (t *Table) Expire() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	tNow := time.Now().Unix()
	// if the game is start and have been played for longer than 300 seconds
	// or if the game is not start for 3600 seconds -> 1 hour
	// there should be some network errors occur
	// so we have to manually release the table otherwise the users are not able to join game any more
	if t.IsStart() {
		return (tNow - t.startTime) > 300
	}
	return (tNow - t.startTime) > 3600
}

// start the game in the table
func (t *Table) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startTime = time.Now().Unix()
	t.TStat = statInGame
}

// stop the game
func (t *Table) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TStat = statWaiting
	t.startTime = time.Now().Unix()
}

// ob join the table
func (t *Table) JoinOB(u *User) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.obs.Join(u)
}

// check if the table is full
func (t *Table) IsFull() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._1p != nil && t._2p != nil
}

// player join the Table
func (t *Table) Join(u *User) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch {
	case t._1p == nil:
		t._1p = u
	case t._2p == nil:
		t._2p = u
	default:
		err = ErrRoomFull
	}
	return
}

// quit a user
func (t *Table) Quit(uid int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if uid < 0 {
		return
	}
	switch uid {
	case t._1p.GetUid():
		t._1p = nil
		t.ready1p = false
	case t._2p.GetUid():
		t._2p = nil
		t.ready2p = false
	default:
		t.obs.Quit(uid)
	}
}

// get bet
func (t *Table) GetBet() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TBet
}

// get all users
func (t *Table) GetAllUsers() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	us := t.obs.GetAll()
	us = append(us, t._1p.GetUid())
	us = append(us, t._2p.GetUid())
	return us
}

// get all observers
func (t *Table) GetObservers() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.obs.GetAll()
}

// get all observers' connections
func (t *Table) GetObConns() []*net.TCPConn {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.obs.GetConns()
}

// get 1p conn
func (t *Table) Get1pConn() *net.TCPConn {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._1p.GetConn()
}

// get 2p conn
func (t *Table) Get2pConn() *net.TCPConn {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._2p.GetConn()
}

// get 1p uid
func (t *Table) Get1pUid() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._1p.GetUid()
}

// get 2p uid
func (t *Table) Get2pUid() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._2p.GetUid()
}

// get all conns
func (t *Table) GetAllConns() []*net.TCPConn {
	conns := t.GetObConns()
	if c := t.Get1pConn(); c != nil {
		conns = append(conns, c)
	}
	if c := t.Get2pConn(); c != nil {
		conns = append(conns, c)
	}
	return conns
}

// close all ob connections, for game server used
func (t *Table) QuitAllObs() {
	t.obs.QuitAll()
}

// get all players
func (t *Table) GetPlayers() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return []int{t._1p.GetUid(), t._2p.GetUid()}
}

// get opponent
func (t *Table) GetOpponent(uid int) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t._1p.GetUid() == uid {
		return t._2p.GetUid()
	}
	return t._1p.GetUid()
}

// check if the player is 1p or 2p
func (t *Table) Is1p(uid int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t._1p.GetUid() == uid
}

// get 1p game
func (t *Table) GetGame1p() *tetris.Game {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.g1p
}

// get 2p game
func (t *Table) GetGame2p() *tetris.Game {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.g2p
}
