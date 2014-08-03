package main

import (
	"time"

	"github.com/gogames/go_tetris/types"
)

// user cache never expire
var users = types.NewUsers()

func initUsers() {
	us := queryUsers()
	users.Add(us...)

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
		time.Sleep(time.Minute)
	}
}

func setNextGiveoutTime() {
	tN := time.Now()
	nextGiveoutTime = time.Date(tN.Year(), tN.Month(), tN.Day()+1, 0, 0, 0, 0, time.Local)
}

func getUserById(uid int) *types.User {
	return users.GetById(uid)
	// if u := users.GetById(uid); u != nil {
	// 	return u
	// }
	// if u := queryUser("Uid", uid); u != nil {
	// 	u.Update()
	// 	insertOrUpdateUser(u)
	// 	users.Add(u)
	// 	return u
	// }
	// return nil
}

func getUserByEmail(email string) *types.User {
	return users.GetByEmail(email)
	// if u := users.GetByEmail(email); u != nil {
	// 	return u
	// }
	// if u := queryUser("Email", email); u != nil {
	// 	u.Update()
	// 	insertOrUpdateUser(u)
	// 	users.Add(u)
	// 	return u
	// }
	// return nil
}

func getUserByNickname(nickname string) *types.User {
	return users.GetByNickname(nickname)
	// if u := users.GetByNickname(nickname); u != nil {
	// 	return u
	// }
	// if u := queryUser("Nickname", nickname); u != nil {
	// 	u.Update()
	// 	insertOrUpdateUser(u)
	// 	users.Add(u)
	// 	return u
	// }
	// return nil
}
