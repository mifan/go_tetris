package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initGraceful() { go notify() }

func notify() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigs:
	}
	deactivateServer()
	log.Info("gracefully exit the program...")
	time.Sleep(time.Second)
	os.Exit(1)
}
