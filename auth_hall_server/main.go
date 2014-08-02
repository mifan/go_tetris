package main

func init() {
	initFlags()
	initConf()
	initLogger()
	initClient()
	initDatabase()
	initSession()
	initPubServer()
	initPrivServer()
	initUsers()
	initBitcoin()
	initQueue()
	initHall()
	initGraceful()
}

func main() {
	c := make(chan bool, 0)
	<-c
}
