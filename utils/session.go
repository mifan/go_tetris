package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	cookieSessId = "sessId"
)

// session store
type sessionStore struct {
	sess           map[string]*session // sessionId -> *session
	expireInMinute int64               // minimum 30 minutes
	mu             sync.RWMutex
}

func NewSessionStore() *sessionStore {
	ss := &sessionStore{
		sess:           make(map[string]*session),
		expireInMinute: 120,
	}
	return ss.start()
}

// session store start
func (ss *sessionStore) start() *sessionStore {
	go ss.delExpireSessions()
	return ss
}

// delete expire sessions
func (ss *sessionStore) delExpireSessions() {
	getExpire := func() []string {
		ss.mu.RLock()
		defer ss.mu.RUnlock()
		tNow := time.Now().Unix()
		sss := make([]string, 0)
		for sessId, v := range ss.sess {
			if (tNow-v.updated)/60 > ss.expireInMinute {
				sss = append(sss, sessId)
			}
		}
		return sss
	}
	for {
		time.Sleep(1 * time.Minute)
		ss.delSession(getExpire()...)
	}
}

// online users
func (ss *sessionStore) NumOfOnlineUsers() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.sess)
}

// generate unique session id
func (ss *sessionStore) generateSessionId(ctx interface{}) string {
	return getRand()
}

// get session id from context
func (ss *sessionStore) getSessionId(ctx interface{}) string {
	sess, err := GetCookie(cookieSessId, ctx)
	fmt.Printf("get session id %v, err is %v\n", sess, err)
	if err != nil {
		return ""
	}
	return sess
}

// check if session id is already exist
func (ss *sessionStore) isSessIdExist(sessId string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sess[sessId] != nil
}

// delete session from session store
func (ss *sessionStore) delSession(sessionIds ...string) {
	fmt.Println("delete session by ids: ", sessionIds...)
	ss.mu.Lock()
	defer ss.mu.Unlock()
	for _, sessionId := range sessionIds {
		delete(ss.sess, sessionId)
	}
}

// create a session or refresh it
func (ss *sessionStore) CreateSession(ctx interface{}) {
	sessId := ss.getSessionId(ctx)
	if !ss.isSessIdExist(sessId) {
		sessId = ss.generateSessionId(ctx)
	}
	if err := AddCookie(cookieSessId, sessId, ctx); err != nil {
		fmt.Println("add cookie error: ", err)
		return
	}
	if err := SetCookie(cookieSessId, sessId, ctx); err != nil {
		fmt.Println("set cookie error: ", err)
		return
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sess[sessId] = newSession()
}

// store data in session
func (ss *sessionStore) SetSession(key string, val interface{}, ctx interface{}) {
	sessId := ss.getSessionId(ctx)
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	fmt.Printf("set session key %s, value %v\n", key, val)
	ss.sess[sessId].set(key, val)
}

// delete data from session
func (ss *sessionStore) DeleteKey(key string, ctx interface{}) {
	sessId := ss.getSessionId(ctx)
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	fmt.Printf("delete session key %s\n", key)
	ss.sess[sessId].del(key)
}

// del the session id
func (ss *sessionStore) DelSession(ctx interface{}) {
	ss.delSession(ss.getSessionId(ctx))
}

// get data from session
func (ss *sessionStore) GetSession(key string, ctx interface{}) interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	sessId := ss.getSessionId(ctx)
	fmt.Printf("get session by sessionId %s and key %s from session %v\n", sessId, key, ss.String())
	if sess, ok := ss.sess[sessId]; ok {
		return sess.get(key)
	}
	return nil
}

// String for testing
func (ss *sessionStore) String() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	str := ""
	for sessId, sess := range ss.sess {
		str += sessId + " --> \n"
		for name, val := range sess.vals {
			str += fmt.Sprintf("\t%v -> %v\n", name, val)
		}
	}
	return str
}

// session
type session struct {
	updated int64
	vals    map[string]interface{}
	mu      sync.RWMutex
}

func newSession() *session {
	return &session{
		updated: time.Now().Unix(),
		vals:    make(map[string]interface{}),
	}
}

func (s *session) set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals[key] = val
	s.updated = time.Now().Unix()
}

func (s *session) get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vals[key]
}

func (s *session) del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vals, key)
}
