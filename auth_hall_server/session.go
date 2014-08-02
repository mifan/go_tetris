package main

import "github.com/gogames/go_tetris/utils"

// session store
var session = utils.NewSessionStore()

func initSession() {
	session.Init(querySessions())
}
