package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gogames/go_tetris/tetris"
	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
)

func initSocketServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+gameServerSockPort)
	if err != nil {
		log.Critical("can not resolve tcp address: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Critical("can not resolve tcp address: %v", err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}

	go func() {
		log.Info("successfully initialization, socket server accepting connection...")
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.Warn("do not accept tcp connection: %v", err)
				continue
			}
			if !isServerActive() {
				log.Info("the server is currently inactive, do not accept new connections...")
				closeConn(conn)
				continue
			}
			go serveTcpConn(conn)
		}
	}()
}

const (
	opRotate = "rotate"
	opLeft   = "left"
	opRight  = "right"
	opDown   = "down"
	opDrop   = "drop"
	opHold   = "hold"
)

// request command
const (
	cmdAuth    = "auth"
	cmdChat    = "chat"
	cmdOperate = "operate"
	cmdReady   = "switchState"
	cmdQuit    = "quit"
)

// response description
const (
	descError                      = "error"
	descChatMsg                    = "chat"
	descRefreshNormalTableInfo     = "refreshNormal"
	descRefreshTournamentTableInfo = "refreshTournament"
	descSysMsg                     = "sysMsg"
	descStart                      = "start"
	desc1p                         = "1p"
	desc2p                         = "2p"
	descTimer                      = "timer"
	descGameWin                    = "win"
	descGameLose                   = "lose"
	descGameResult                 = "result"
)

func serveTcpConn(conn *net.TCPConn) {
	// auth -> check the connection
	data, err := recv(conn)
	if err != nil {
		log.Info("can not read from the tcp connection: %v", err)
		closeConn(conn)
		return
	}
	if data.Cmd != cmdAuth {
		log.Debug("the first command is not auth, the data is %v", data)
		send(conn, descError, "the first command should be auth")
		closeConn(conn)
		return
	}
	// parse the token, see what to do next
	uid, nickname, isApply, isOb, isTournament, tid, err := utils.ParseToken(data.Data)
	if err != nil {
		log.Debug("can not parse the token: %v", err)
		send(conn, descError, fmt.Sprintf("can not parse the token %s, are you hacker?", data.Data))
		closeConn(conn)
		return
	}
	// create a new user, add it into tables
	u := types.NewUser(uid, "", "", nickname, "")
	u.SetConn(conn)
	switch {
	case isApply:
		// apply for tournament
		tid, err := authServerStub.Apply(uid)
		if err != nil {
			log.Warn("can not apply for tournament, auth server error: %v", err)
			send(conn, descError, fmt.Sprintf("报名失败, 错误: %v", err))
			closeConn(conn)
			return
		}
		if !tables.IsTableExist(tid) {
			tables.NewTable(tid, "", "", 0)
		}
		if err := tables.JoinTable(tid, u, false); err != nil {
			log.Debug("can not join the table, game server error: %v", err)
			send(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConn(conn)
			return
		}
		refreshTable(tid, true)
		sendAll(descSysMsg, fmt.Sprintf("参赛者 %s 加入", nickname), tables.GetTableById(tid).GetAllConns()...)
	case isOb:
		// inform the auth server that some one is going to observe a game
		if err := obGame(tid, uid, isTournament); err != nil {
			log.Warn("can not ob a game, auth server error: %v", err)
			send(conn, descError, fmt.Sprintf("无法观战, 错误: %v", err))
			closeConn(conn)
			return
		}
		if err := tables.JoinTable(tid, u, true); err != nil {
			log.Critical("can not ob a game, game server error: %v", err)
			send(conn, descError, fmt.Sprintf("无法观战, 错误: %v", err))
			closeConn(conn)
			return
		}
		refreshTable(tid, isTournament)
		sendAll(descSysMsg, fmt.Sprintf("用户 %s 进入观战", nickname), tables.GetTableById(tid).GetAllConns()...)
	default:
		// normal hall
		if err := authServerStub.Join(tid, uid, false); err != nil {
			log.Warn("can not join a game, auth server error: %v", err)
			send(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConn(conn)
			return
		}
		if err := tables.JoinTable(tid, u, isOb); err != nil {
			log.Critical("can not join a game, game server error: %v", err)
			send(conn, descError, fmt.Sprintf("无法加入桌子, 错误: %v", err))
			closeConn(conn)
			return
		}
		refreshTable(tid, false)
		sendAll(descSysMsg, fmt.Sprintf("玩家 %s 加入游戏", nickname), tables.GetTableById(tid).GetAllConns()...)
	}
	go handleConn(conn, uid, tid, nickname, isOb, tables.GetTableById(tid).Is1p(uid), isTournament)
}

func handleConn(conn *net.TCPConn, uid, tid int, nickname string, isOb, is1p, isTournament bool) {
forLoop:
	for {
		table := tables.GetTableById(tid)
		// receive data from client
		data, err := recv(conn)
		if err != nil {
			log.Debug("can not receive request from table %d, user %s: %v", tid, nickname, err)
			quit(tid, uid, nickname, is1p, isTournament)
			closeConn(conn)
			refreshTable(tid, isTournament)
			return
		}
		switch data.Cmd {
		case cmdChat:
			msg := fmt.Sprintf("%s: %s", nickname, data.Data)
			sendAll(descChatMsg, msg, table.GetObConns()...)
			if !isOb {
				send(table.Get1pConn(), descChatMsg, msg)
				send(table.Get2pConn(), descChatMsg, msg)
			}
		case cmdReady:
			if table.IsStart() {
				log.Debug("receive a ready command after game is start: %v", data)
				send(conn, descError,
					"game is already start, how come you send a ready switch command to me! Are you hacker?")
				continue forLoop
			}
			if isOb {
				send(conn, descError, "观战者无法准备或者取消准备")
				continue forLoop
			}
			if err := authServerStub.SwitchReady(tid, uid); err != nil {
				log.Warn("can not switch user's ready state: %v", err)
				continue forLoop
			}
			refreshTable(tid, isTournament)
		case cmdQuit:
			// quit a game
			quit(tid, uid, nickname, is1p, isTournament)
			closeConn(conn)
			refreshTable(tid, isTournament)
			return
		case cmdOperate:
			if !table.IsStart() {
				continue forLoop
			}
			var g *tetris.Game
			if is1p {
				g = table.GetGame1p()
			} else {
				g = table.GetGame2p()
			}
			switch data.Data {
			case opDown:
				g.MoveDown()
			case opDrop:
				g.DropDown()
			case opLeft:
				g.MoveLeft()
			case opRight:
				g.MoveRight()
			case opRotate:
				g.Rotate()
			case opHold:
				g.Hold()
			default:
				send(conn, descError, fmt.Sprintf("operation can only be %s, %s, %s, %s, %s, %s",
					opDown, opDrop, opLeft, opRight, opHold, opRotate))
			}
		default:
			send(conn, descError, fmt.Sprintf("the command %s does not exist, are you hacker?", data.Cmd))
		}
	}
}

// quit a game
func quit(tid, uid int, nickname string, is1p, isTournament bool) {
	table := tables.GetTableById(tid)
	table.Quit(uid)
	if table.IsStart() {
		if is1p {
			table.GameoverChan <- types.Gameover1pQuit
		} else {
			table.GameoverChan <- types.Gameover2pQuit
		}
	}
	if err := authServerStub.Quit(tid, uid, isTournament); err != nil {
		log.Warn("can not quit user %s from table %d: %v", nickname, tid, err)
	}
}

// send to all
func sendAll(desc string, val interface{}, conns ...*net.TCPConn) {
	for _, c := range conns {
		send(c, desc, val)
	}
}

// inform the client side to refresh the table information
func refreshTable(tid int, isTournament bool) {
	table := tables.GetTableById(tid)
	if isTournament {
		sendAll(descRefreshTournamentTableInfo, tid, table.GetAllConns()...)
	} else {
		sendAll(descRefreshNormalTableInfo, tid, table.GetAllConns()...)
	}
}

// inform the auth server, some one is going to ob a game
func obGame(tid, uid int, isTournament bool) error {
	if isTournament {
		return authServerStub.ObTournament(tid, uid)
	}
	return authServerStub.Join(tid, uid, true)
}
