package main

func init() {
	initFlags()
	initConf()
	initLogger()
	initRpcClient()
	initServerStatus()
	initSocketServer()
	initRpcServer()
	initGraceful()
}

func main() {
	c := make(chan bool)
	<-c
}
