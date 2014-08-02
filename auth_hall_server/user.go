package main

import (
	"time"

	"github.com/gogames/go_tetris/types"
)

var users = types.NewUsers()

func initUsers() {
	us := queryUsers()
	users.Add(us...)
	id := 0
	for _, v := range us {
		if v.Uid >= id {
			id = v.Uid + 1
		}
	}
	users.SetNextId(id)

	log.Info("initialize users in the cache...")
	go energyGiveout()
}

var nextGiveoutTime time.Time

// give out energy to all users every 00:00:00 on tiemzone utc +8
func energyGiveout() {
	setNextGiveoutTime()
	for {
		if time.Now().Sub(nextGiveoutTime).Seconds() >= 0 {
			setNextGiveoutTime()
			users.EnergyGiveout()
			insertOrUpdateUser(users.GetAllUsers()...)
		}
	}
}

func setNextGiveoutTime() {
	tN := time.Now()
	nextGiveoutTime = time.Date(tN.Year(), tN.Month(), tN.Day()+1, 0, 0, 0, 0, time.Local)
}
