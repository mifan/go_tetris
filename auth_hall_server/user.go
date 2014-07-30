package main

import "github.com/gogames/go_tetris/types"

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
}
