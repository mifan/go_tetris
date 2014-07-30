package main

import "github.com/gogames/go_tetris/types"

var (
	clients *types.GameServersRpc
)

func initClient() {
	clients = types.NewGameServerRpc(gameServerRpcPort)
}
