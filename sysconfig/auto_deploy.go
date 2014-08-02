package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	bufferLength = 1 << 10
	pathToAuth   = "/usr/local/bin/auth_hall_server"
	pathToGame   = "/usr/local/bin/game_server"
	gopathToAuth = "/go/bin/auth_hall_server"
	gopathToGame = "/go/bin/game_server"
	gogetAuth    = "go get -u github.com/gogames/go_tetris/auth_hall_server"
	gogetGame    = "go get -u github.com/gogames/go_tetris/game_server"
)

var (
	hashAuth string
	hashGame string
	log      *logs.BeeLogger
)

func init() {
	log = logs.NewLogger(100000)
	fmt.Println(log.SetLogger("console", ""))
	checkHash(pathToAuth, &hashAuth)
	checkHash(pathToGame, &hashGame)

	go auth()
	go game()
}

func goget(path string) {
	log.Info("go get %s", path)
	if err := exec.Command("sh -c", path).Run(); err != nil {
		log.Info("can not go get %s: %v", path, err)
	}
}

func checkHash(path string, hash *string) (changed bool) {
	log.Info("checking hash of %s", path)
	f, err := os.Open(path)
	if err != nil {
		log.Info("can not open file %s: %v", path, err)
		return
	}
	defer f.Close()
	bytes := make([]byte, 0)
	reader := bufio.NewReader(f)
	for {
		buffer := make([]byte, bufferLength)
		n, err := reader.Read(buffer)
		if err != nil {
			log.Info("can not read byte: %v", err)
			break
		}
		bytes = append(bytes, buffer[:n]...)
	}
	h := md5.New()
	_, err = h.Write(bytes)
	if err != nil {
		log.Info("can not write hash: %v", err)
		return
	}
	newHash := hex.EncodeToString(h.Sum(nil))
	*hash = newHash
	return newHash == *hash
}

func restart(gopath, path, restartCMD string) {
	log.Info("moving %s to %s and restart", gopath, path)
	if err := exec.Command("sh -c", fmt.Sprintf("mv %s %s", gopath, path)).Run(); err != nil {
		log.Critical("can not move %s to %s: %v", gopath, path, err)
		return
	}
	if err := exec.Command("sh -c", restartCMD).Run(); err != nil {
		log.Critical("can not restart the %s: %v", path, err)
		return
	}
}

func auth() {
	for {
		goget(gogetAuth)
		if checkHash(gopathToAuth, &hashAuth) {
			restart(gopathToAuth, pathToAuth, "service auth_server stop; sleep 60; service auth_server start;")
		}
		time.Sleep(time.Minute)
	}
}

func game() {
	for {
		goget(gogetGame)
		if checkHash(gopathToAuth, &hashGame) {
			restart(gopathToGame, pathToGame, "service game_server stop; sleep 60; service game_server start;")
		}
		time.Sleep(time.Minute)
	}
}

func main() {
	c := make(chan bool)
	<-c
}
