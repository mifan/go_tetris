package main

import (
	"fmt"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/timer"
	"github.com/gogames/go_tetris/types"
)

// auth server inform game server to become inactive
// do not accept any new connections
// handle the current connections
func (stub) Deactivate() {
	go deactivateServer()
}

// count down after a game start
func countDown(table *types.Table) {
	t := timer.NewTimer(1000)
	for i := 3; i > 0; i-- {
		sendAll(descStart, i, table.GetAllConns()...)
		t.Wait()
	}
	t.Stop()
	sendAll(descStart, 0, table.GetAllConns()...)
}

// auth server inform game server to start a table
func (stub) Start(tid int) {
	go func() {
		table := tables.GetTableById(tid)
		countDown(table)
		table.StartGame()
		go table.UpdateTimer()
		serveGame(tid)
	}()
}

// auth server inform game server the game result
func (stub) SetNormalGameResult(tid, winnerUid, bet int) {
	construct := func(win bool, bet int) (str string) {
		if win {
			str = "你很厉害哦!!"
			if bet > 0 {
				str += fmt.Sprintf(" 本局游戏你赢得了 %d mBTC", bet)
			}
			return str
		}
		if bet > 0 {
			str = fmt.Sprintf("本局游戏你输掉了 %d mBTC! ", bet)
		}
		str += "再接再厉!"
		return str
	}
	table := tables.GetTableById(tid)
	switch winnerUid {
	case table.Get1pUid():
		send(table.Get1pConn(), descGameWin, construct(true, bet))
		send(table.Get2pConn(), descGameLose, construct(false, bet))
		sendAll(descGameResult, "1P 赢得本局游戏", table.GetObConns()...)
	case table.Get2pUid():
		send(table.Get2pConn(), descGameWin, construct(true, bet))
		send(table.Get1pConn(), descGameLose, construct(false, bet))
		sendAll(descGameResult, "2P 赢得本局游戏", table.GetObConns()...)
	default:
	}
	table.ResetTable()
	refreshTable(tid, false)
}

// TODO: not confirmed yet
func (stub) SetTournamentResult(tid, winnerUid int, isFinalRound bool) {
	construct := func(win, isFinalRound bool) (str string) {
		if win {
			if isFinalRound {
				str = "恭喜你获得冠军!"
			} else {
				str = "恭喜你获得进入下一轮游戏的资格!"
			}
		} else if isFinalRound {
			str = "恭喜你获得亚军!"
		} else {
			str = "不要气馁, 再接再厉!"
		}
		return
	}
	table := tables.GetTableById(tid)
	switch winnerUid {
	case table.Get1pUid():
		send(table.Get1pConn(), descGameWin, construct(true, isFinalRound))
		send(table.Get2pConn(), descGameLose, construct(false, isFinalRound))
		closeConn(table.Get2pConn())
		sendAll(descGameResult, "1P 赢得本局游戏", table.GetObConns()...)
	case table.Get2pUid():
		send(table.Get2pConn(), descGameWin, construct(true, isFinalRound))
		send(table.Get1pConn(), descGameLose, construct(false, isFinalRound))
		closeConn(table.Get1pConn())
		sendAll(descGameResult, "2P 赢得本局游戏", table.GetObConns()...)
	default:
	}
	table.ResetTable()
	table.QuitAllObs()
}

// game server serve the game
func serveGame(tid int) {
	table := tables.GetTableById(tid)
	for {
		select {

		// table timer
		case remain := <-table.RemainedSecondsChan:
			sendAll(descTimer, remain, table.GetAllConns()...)

		// game over
		case gameover := <-table.GameoverChan:
			switch gameover {
			case types.GameoverNormal:
				// normal game over
				gameOver(tid)
			case types.Gameover1pQuit:
				// 1p quit, game over, 2p winner
				gameOver(tid, false)
			case types.Gameover2pQuit:
				// 2p quit, game over, 1p winner
				gameOver(tid, true)
			}
			return

		// 1p
		case msg := <-table.GetGame1p().MsgChan:
			switch msg.Description {
			// ko, audio only send to the player himself
			case tetris.DescAudio, tetris.DescKo:
				sendAll(desc1p, msg, table.Get1pConn())
			// clear, combo, attack only sends to the player and obs
			case tetris.DescClear, tetris.DescCombo, tetris.DescAttack:
				sendAll(desc1p, msg, table.Get1pConn())
				sendAll(desc1p, msg, table.GetObConns()...)
			// the others send to all
			default:
				sendAll(desc1p, msg, table.GetAllConns()...)
			}

		case beingKo := <-table.GetGame1p().BeingKOChan:
			if beingKo {
				table.GetGame2p().KoOpponent()
				sendAll(desc1p, tetris.NewMessage(tetris.DescBeingKo, table.GetGame2p().GetKo()), table.Get1pConn())
				sendAll(desc1p, tetris.NewMessage(tetris.DescBeingKo, table.GetGame2p().GetKo()), table.GetObConns()...)
				if table.GetGame2p().GetKo() >= 5 {
					table.GetGame1p().GameoverChan <- true
				}
			}

		// attack 2p
		case attack := <-table.GetGame1p().AttackChan:
			table.GetGame2p().BeingAttacked(attack)

		// 1p game over, 2p win
		case gameover := <-table.GetGame1p().GameoverChan:
			if gameover {
				gameOver(tid, false)
				return
			}

		// 2p
		case msg := <-table.GetGame2p().MsgChan:
			// ko, audio only send to the player himself
			switch msg.Description {
			case tetris.DescAudio, tetris.DescKo:
				sendAll(desc2p, msg, table.Get2pConn())
			case tetris.DescClear, tetris.DescCombo, tetris.DescAttack:
				sendAll(desc2p, msg, table.Get2pConn())
				sendAll(desc2p, msg, table.GetObConns()...)
			default:
				sendAll(desc2p, msg, table.GetAllConns()...)
			}

		case attack := <-table.GetGame2p().AttackChan:
			table.GetGame2p().BeingAttacked(attack)

		// 2p game over, 1p win
		case gameover := <-table.GetGame2p().GameoverChan:
			if gameover {
				gameOver(tid, true)
				return
			}

		case beingKo := <-table.GetGame2p().BeingKOChan:
			if beingKo {
				table.GetGame1p().KoOpponent()
				sendAll(desc2p, tetris.NewMessage(tetris.DescBeingKo, table.GetGame1p().GetKo()), table.Get2pConn())
				sendAll(desc2p, tetris.NewMessage(tetris.DescBeingKo, table.GetGame1p().GetKo()), table.GetObConns()...)
				if table.GetGame1p().GetKo() >= 5 {
					table.GetGame2p().GameoverChan <- true
				}
			}
		}
	}
}

// stop the game
// inform the auth server that the game is over
func gameOver(tid int, is1pWin ...bool) {
	table := tables.GetTableById(tid)
	table.StopGame()
	var is1pWinner = false
	var winner, loser int
	var err error

	// normal checker
	if len(is1pWin) == 0 {
		g1p, g2p := table.GetGame1p(), table.GetGame2p()
		switch {
		case g1p.GetKo() > g2p.GetKo():
			is1pWinner = true
		case g1p.GetKo() < g2p.GetKo():
		default:
			if g1p.GetScore() >= g2p.GetScore() {
				is1pWinner = true
			}
		}
	} else {
		is1pWinner = is1pWin[0]
	}

	// inform the auth server
	if is1pWinner {
		winner, loser = table.Get1pUid(), table.Get2pUid()
	} else {
		winner, loser = table.Get2pUid(), table.Get1pUid()
	}

	// 1e5 magic number
	if tid >= 1e5 {
		err = authServerStub.SetTournamentResult(tid, winner, loser)
	} else {
		err = authServerStub.SetNormalGameResult(tid, winner, loser)
	}
	if err != nil {
		log.Warn("can not set game result for table %d: %v", tid, err)
	}
}
