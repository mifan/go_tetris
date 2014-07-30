package types

import (
	"encoding/json"
	"net"
	"sync"
)

// observers
type obs struct {
	users map[int]*User
	mu    sync.RWMutex
}

func NewObs() *obs {
	return &obs{
		users: make(map[int]*User),
	}
}

func (this *obs) Wrap() []*User {
	this.mu.RLock()
	defer this.mu.Unlock()
	us := make([]*User, 0)
	for _, u := range this.users {
		us = append(us, u)
	}
	return us
}

// join a new observer
func (this *obs) Join(u *User) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.users[u.Uid] = u
}

// quit an observer
func (this *obs) Quit(uId int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.users, uId)
}

// quit all users
func (this *obs) QuitAll() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for uid, u := range this.users {
		u.Close()
		delete(this.users, uid)
	}
}

// get all observers uid
func (this *obs) GetAll() []int {
	us := make([]int, 0)
	for i, _ := range this.users {
		us = append(us, i)
	}
	return us
}

// get all observers' connection
func (this *obs) GetConns() []*net.TCPConn {
	this.mu.RLock()
	defer this.mu.RUnlock()
	conns := make([]*net.TCPConn, 0)
	for _, u := range this.users {
		if c := u.GetConn(); c != nil {
			conns = append(conns, c)
		}
	}
	return conns
}

func (this *obs) MarshalJSON() ([]byte, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	res := make([]*User, 0)
	for _, v := range this.users {
		res = append(res, v)
	}
	return json.Marshal(res)
}
