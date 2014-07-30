package types

import (
	"encoding/json"
	"fmt"
	"sync"
)

// normal hall
type NormalHall struct {
	*Tables
	currId, maxId int
	mu            sync.RWMutex
}

func NewNormalHall() *NormalHall {
	return &NormalHall{
		Tables: NewTables(),
		maxId:  1 << 9,
		currId: 1,
	}
}

// for hprose
func (h NormalHall) Wrap() []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Tables.Wrap()
}

func (h NormalHall) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"tables": h.Tables,
	})
}

// find the next table id of normal hall
func (h *NormalHall) NextTableId() int {
	var i int
	for h.IsTableExist(h.currId) {
		h.mu.Lock()
		if i == h.maxId {
			h.maxId += 1
			h.currId = h.maxId
			h.mu.Unlock()
			break
		}
		i++
		h.currId = (h.currId + 1) % (h.maxId + 1)
		h.mu.Unlock()
	}
	return h.currId
}

const (
	TournamentStatWaiting = iota
	TournamentStatInGame
	TournamentStatPending
	TournamentStatEnd
)

// tournament hall
type TournamentHall struct {
	*Tables
	winners                        map[int][]string
	losers                         map[int][]string
	numCandidate, currentCandidate int
	currentTableId                 int
	awardGold, awardSilver         int
	goldGetter, silverGetter       string
	round                          int
	stat                           int
	tBase                          int
	host                           string
	idleTables                     map[int]int // someone apply and then quit, tables would be idle
	mu                             sync.RWMutex
}

func NewTournamentHall(numCandidate, awardGold, awardSilver int, host string) *TournamentHall {
	return &TournamentHall{
		Tables:       NewTables(),
		winners:      make(map[int][]string),
		losers:       make(map[int][]string),
		numCandidate: numCandidate,
		awardGold:    awardGold,
		awardSilver:  awardSilver,
		round:        0,
		tBase:        0,
		stat:         TournamentStatWaiting,
		host:         host,
		idleTables:   make(map[int]int),
	}
}

func (th *TournamentHall) GetStat() (stat string) {
	th.mu.RLock()
	defer th.mu.RUnlock()
	switch th.stat {
	case TournamentStatWaiting:
		stat = "接受报名中..."
	case TournamentStatPending:
		stat = fmt.Sprintf("等待 Round %d 开始", th.round)
	case TournamentStatInGame:
		stat = fmt.Sprintf("Round %d 进行中", th.round)
	default:
		stat = "争霸赛已结束"
	}
	return stat
}

func (th *TournamentHall) getTitle() string {
	return fmt.Sprintf("Round %d", th.round+1)
}

// for hprose
func (th *TournamentHall) Wrap() map[string]interface{} {
	stat := th.GetStat()
	th.mu.RLock()
	defer th.mu.RUnlock()
	return map[string]interface{}{
		"numCandidate":     th.numCandidate,
		"currNumCandidate": th.currentCandidate,
		"awardGold":        fmt.Sprintf("%d mBTC", th.awardGold),
		"awardSilver":      fmt.Sprintf("%d mBTC", th.awardSilver),
		"status":           stat,
		"host":             th.host,
		"tables":           th.Tables.Wrap(),
	}
}

func (th *TournamentHall) MarshalJSON() ([]byte, error) {
	stat := th.GetStat()
	th.mu.RLock()
	defer th.mu.RUnlock()
	return json.Marshal(map[string]interface{}{
		"numCandidate":       th.numCandidate,
		"currNumOfCandidate": th.currentCandidate,
		"awardGold":          fmt.Sprintf("%d mBTC", th.awardGold),
		"awardSilver":        fmt.Sprintf("%d mBTC", th.awardSilver),
		"status":             stat,
		"host":               th.host,
		"tables":             th.Tables,
	})
}

var errCantCreateNewTable = fmt.Errorf("can not create new table because of the status is not waiting or in game")
var errIncorrectTableId = fmt.Errorf("the table id is incorrect, should not be negative")

// create a new table
func (th *TournamentHall) newTable(id int) error {
	if id <= 0 {
		return errIncorrectTableId
	}
	if th.stat != TournamentStatWaiting && th.stat != TournamentStatInGame {
		return errCantCreateNewTable
	}
	return th.Tables.NewTable(id, th.getTitle(), th.host, 0)
}

var errCantAcceptMoreApplication = fmt.Errorf("不好意思, 报名人数已满, 请参加下期的争霸赛~")

// apply
func (th *TournamentHall) Apply(u *User) (int, error) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.currentCandidate >= th.numCandidate {
		return -1, errCantAcceptMoreApplication
	}
	// first check the idles
	for tid, v := range th.idleTables {
		if v > 0 {
			v--
			if v == 0 {
				delete(th.idleTables, tid)
			} else {
				th.idleTables[tid] = v
			}
			return tid, th.GetTableById(tid).Join(u)
		}
	}
	// then check current table id
	t := th.GetTableById(th.currentTableId)
	if t != nil && !t.IsFull() {
		return th.currentTableId, t.Join(u)
	}
	nid := th.nextTableId()
	if err := th.newTable(nid); err != nil {
		return nid, err
	}
	if err := th.GetTableById(nid).Join(u); err != nil {
		return nid, err
	}
	th.currentTableId = nid
	th.currentCandidate++
	return nid, nil
}

// allocate
func (th *TournamentHall) Allocate(u *User) (int, error) {
	th.mu.Lock()
	defer th.mu.Unlock()
	// first check the idles
	for tid, v := range th.idleTables {
		if v > 0 {
			v--
			if v == 0 {
				delete(th.idleTables, tid)
			} else {
				th.idleTables[tid] = v
			}
			return tid, th.GetTableById(tid).Join(u)
		}
	}
	// then check current table id
	t := th.GetTableById(th.currentTableId)
	if t != nil && !t.IsFull() {
		return th.currentTableId, t.Join(u)
	}
	nid := th.nextTableId()
	if err := th.newTable(nid); err != nil {
		return nid, err
	}
	if err := th.GetTableById(nid).Join(u); err != nil {
		return nid, err
	}
	th.currentTableId = nid
	return nid, nil
}

// quit a user
func (th *TournamentHall) Quit(tid, uid int) {
	th.mu.Lock()
	defer th.mu.Unlock()
	switch table := th.GetTableById(tid); uid {
	case table._1p.GetUid():
		th.currentCandidate--
		th.idleTables[tid]++
	case table._2p.GetUid():
		th.currentCandidate--
		th.idleTables[tid]++
	default:
		table.obs.Quit(uid)
	}
}

// find the next table id of tournament hall
func (th *TournamentHall) nextTableId() int {
	var nid int
	switch th.stat {
	case TournamentStatInGame, TournamentStatWaiting:
		th.tBase++
		nid = (th.round+1)*1e5 + th.tBase
		th.currentTableId = nid
	default:
		nid = -1
	}
	return nid
}

func (th *TournamentHall) SetStatPending() {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.stat != TournamentStatPending {
		th.stat = TournamentStatPending
		th.round++
		th.tBase = 0
	}
}

func (th *TournamentHall) SetStatInGame() {
	th.mu.Lock()
	defer th.mu.Unlock()
	if th.stat != TournamentStatInGame {
		th.stat = TournamentStatInGame
		th.tBase = 0
	}
}

func (th *TournamentHall) SetStatEnd() {
	th.mu.Lock()
	defer th.mu.Unlock()
	th.stat = TournamentStatEnd
}

func (th *TournamentHall) SetWinnerLoser(tableId, uidWin int) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if uidWin < 0 {
		return
	}
	t := th.GetTableById(tableId)
	if t == nil {
		return
	}
	win, lose := "", ""
	switch uidWin {
	case t._1p.GetUid():
		win = t._1p.Nickname
		lose = t._2p.Nickname
	case t._2p.GetUid():
		win = t._2p.Nickname
		lose = t._1p.Nickname
	default:
		return
	}
	if th.winners[th.round] == nil {
		th.winners[th.round] = make([]string, 0)
	}
	th.winners[th.round] = append(th.winners[th.round], win)
	if th.losers[th.round] == nil {
		th.losers[th.round] = make([]string, 0)
	}
	th.losers[th.round] = append(th.losers[th.round], lose)
	th.DelTable(tableId)
}

func (th *TournamentHall) SetGold(gold string) {
	th.mu.Lock()
	defer th.mu.Unlock()
	th.goldGetter = gold
}

func (th *TournamentHall) SetSilver(silver string) {
	th.mu.Lock()
	defer th.mu.Unlock()
	th.silverGetter = silver
}

func (th *TournamentHall) ShouldEnd() bool {
	th.mu.Lock()
	defer th.mu.Unlock()
	fmt.Printf("round: %d, number of candiate: %d\n", th.round, th.numCandidate)
	return 1<<uint(th.round) == th.numCandidate
}

// check if it is full
// if yes, send the game to pending mode
// start a timer in game server(count 30 seconds)
// then send the game to inGame mode
func (th *TournamentHall) IsFull() bool {
	th.mu.RLock()
	defer th.mu.RUnlock()
	numTables := th.numCandidate >> uint(th.round+1)
	if len(th.Tables.Tables) == numTables {
		for _, t := range th.Tables.Tables {
			if !t.IsFull() {
				return false
			}
		}
		return true
	}
	return false
}
