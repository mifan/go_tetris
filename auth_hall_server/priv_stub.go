package main

import (
	"fmt"

	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

var errInsufficientEnergy = fmt.Errorf("能量不足, 每局游戏需消耗 1 能量")

// register a game server
func (privStub) Register(maxConn int, ctx interface{}) {
	clients.NewGameServer(utils.GetIp(ctx), maxConn)
}

// deactivate a game server
func (privStub) Deactivate(ctx interface{}) {
	clients.Deactivate(utils.GetIp(ctx))
}

// unregister a game server
func (privStub) Unregister(ctx interface{}) {
	clients.Delete(utils.GetIp(ctx))
}

// join a game
func (privStub) Join(tid, uid int, isOb bool) error {
	u := users.GetById(uid)
	if u == nil {
		return fmt.Errorf(errUserNotExist, uid)
	}
	if u.GetEnergy() <= 0 {
		return errInsufficientEnergy
	}
	if err := normalHall.JoinTable(tid, u, isOb); err != nil {
		return err
	}
	users.SetBusy(uid)
	return nil
}

// observe a tournament
func (privStub) ObTournament(tid, uid int) error {
	u := users.GetById(uid)
	if u == nil {
		return fmt.Errorf(errUserNotExist, uid)
	}
	if err := tournamentHall.JoinTable(tid, u, true); err != nil {
		return err
	}
	users.SetBusy(uid)
	return nil
}

// set normal game result
func (privStub) SetNormalGameResult(tid, winner, loser int, ctx interface{}) {
	t := normalHall.GetTableById(tid)
	// update winner info
	func() {
		w := users.GetById(winner)
		upts := make([]types.UpdateInterface, 0)
		upts = append(upts, types.NewUpdateInt(types.UF_Balance, w.Balance+w.Freezed*2))
		upts = append(upts, types.NewUpdateInt(types.UF_Freezed, 0))
		upts = append(upts, types.NewUpdateInt(types.UF_Win, w.Win+1))
		if w.Win > (w.Level * w.Level) {
			upts = append(upts, types.NewUpdateInt(types.UF_Level, w.Level+1))
		}
		if err := w.Update(upts...); err != nil {
			log.Critical("normal hall -> can not update winner %v: %v", w.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(w) })
	}()

	// update loser info
	func() {
		l := users.GetById(loser)
		upts := make([]types.UpdateInterface, 0)
		upts = append(upts, types.NewUpdateInt(types.UF_Freezed, 0))
		upts = append(upts, types.NewUpdateInt(types.UF_Lose, l.Lose+1))
		if err := l.Update(upts...); err != nil {
			log.Critical("normal hall -> can not update loser %v: %v", l.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(l) })
	}()

	// update busy timestamp
	users.SetBusy(t.GetAllUsers()...)

	if err := clients.GetStub(utils.GetIp(ctx)).SetNormalGameResult(tid, winner, t.GetBet()); err != nil {
		log.Warn("can not inform game server to set the game result: %v", err)
	}
}

// set tournament game result
func (privStub) SetTournamentResult(tid, winner, loser int) (int, error) {
	t := tournamentHall.GetTableById(tid)
	// update winner info
	w := users.GetById(winner)
	func() {
		upts := make([]types.UpdateInterface, 0)
		upts = append(upts, types.NewUpdateInt(types.UF_Win, w.Win+1))
		if w.Win > (w.Level * w.Level) {
			upts = append(upts, types.NewUpdateInt(types.UF_Level, w.Level+1))
		}
		if err := w.Update(upts...); err != nil {
			log.Critical("tournament hall -> can not update winner %v: %v", w.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(w) })
	}()

	// update loser info
	func() {
		l := users.GetById(loser)
		if err := l.Update(types.NewUpdateInt(types.UF_Lose, l.Lose+1)); err != nil {
			log.Critical("tournament hall -> can not update loser %v: %v", l.Nickname, err)
		}
		pushFunc(func() { insertOrUpdateUser(l) })
	}()

	// update tournament hall
	tournamentHall.SetWinnerLoser(tid, winner)
	nid, err := tournamentHall.Allocate(w)
	if err != nil {
		log.Critical("tournament hall -> can not allocate user %v to next table: %v", w.Nickname, err)
		return -1, err
	}

	// winner continue to player, update busy timestamp
	users.SetBusy(winner)

	// loser and observers quit, set free
	users.SetFree(loser)
	users.SetFree(t.GetObservers()...)
	return nid, nil
}

// apply for tournament
func (privStub) Apply(uid int) (int, error) {
	tid, err := tournamentHall.Apply(users.GetById(uid))
	if err != nil {
		return -1, err
	}
	users.SetBusy(uid)
	return tid, nil
}

// allocate for tournament
func (privStub) Allocate(uid int) (int, error) {
	tid, err := tournamentHall.Allocate(users.GetById(uid))
	if err != nil {
		return -1, err
	}
	users.SetBusy(uid)
	return tid, nil
}

// quit a user
func (privStub) Quit(tid, uid int, isTournament bool) {
	if isTournament {
		tournamentHall.GetTableById(tid).Quit(uid)
	} else {
		normalHall.GetTableById(tid).Quit(uid)
	}
	users.SetFree(uid)
}

var errTournamentDefaultReady = fmt.Errorf("tournament default is ready and can not set to not ready, what is wrong?")

// switch ready state
func (privStub) SwitchReady(tid, uid int) error {
	if isTournament(tid) {
		return errTournamentDefaultReady
	}
	normalHall.GetTableById(tid).SwitchReady(uid)
	return nil
}

func isTournament(tid int) bool {
	return tid >= 1e5
}
