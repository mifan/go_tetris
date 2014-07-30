package main

import "time"

// the status indicate if a new connection should be accepted or not
var serverStatus = statusChecking

const (
	statusActive = iota
	statusInactive
	statusChecking
)

// activate the server
func activateServer() {
	log.Info("activating the server...")
	for {
		err := authServerStub.Register(maxConn)
		if err == nil {
			log.Info("the server is activated...")
			serverStatus = statusActive
			break
		}
		log.Info("trying to activate the server but failed: %v", err)
		time.Sleep(5 * time.Second)
	}
}

// deactivate the server
func deactivateServer() {
	// set server status to inactive
	// wait for all games done
	log.Info("deactivating the server...")
	serverStatus = statusInactive
	for {
		l := tables.Length()
		if l <= 0 {
			break
		}
		log.Info("the server currently has %d tables, still handling...", l)
		// release the expire tables
		for tid, t := range tables.Tables {
			if !t.IsStart() {
				isTournament := tid >= 1e5
				// inform auth server to quit the users
				if uid1p := t.Get1pUid(); uid1p != -1 {
					authServerStub.Quit(tid, uid1p, isTournament)
				}
				if uid2p := t.Get2pUid(); uid2p != -1 {
					authServerStub.Quit(tid, uid2p, isTournament)
				}
				for _, uid := range t.GetObservers() {
					authServerStub.Quit(tid, uid, isTournament)
				}
				closeConn(t.GetAllConns()...)
				tables.ReleaseExpireTable(tid)
			}
		}
		time.Sleep(5 * time.Second)
	}
	// inform the auth server unregister it
	if err := authServerStub.Unregister(); err != nil {
		log.Warn("can not inform the auth server that it is already unregister: %v", err)
	}
	log.Info("the server is deactivated...\nenter checking status")
	// set the status to checking
	serverStatus = statusChecking
}

// if the server is active
func isServerActive() bool {
	return serverStatus == statusActive
}

func initServerStatus() {
	go func() {
		for {
			if serverStatus == statusChecking {
				activateServer()
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
