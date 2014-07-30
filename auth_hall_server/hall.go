package main

import (
	"time"

	"github.com/gogames/go_tetris/types"
)

var normalHall = types.NewNormalHall()
var tournamentHall *types.TournamentHall

func initHall() {
	go releaseExpires()
}

func releaseExpires() {
	for {
		for tid, _ := range normalHall.GetExpireTables() {
			tt := normalHall.GetTableById(tid)
			if tt == nil {
				continue
			}
			// inform game server
			ip := tt.GetIp()
			if err := clients.GetStub(ip).Delete(tid); err != nil {
				log.Warn("can not inform game server %v to delete table %v: %v", ip, tid, err)
			}
			// release the busy users in cache, including the observers and players
			users.SetFree(tt.GetAllUsers()...)
			// release the expire table also
			normalHall.ReleaseExpireTable(tid)
		}
		time.Sleep(5 * time.Second)
	}
}
