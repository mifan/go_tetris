package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gogames/go_tetris/types"
)

func main() {
	// 8 candidates
	var candidates = []*types.User{
		types.NewUser(1, "1@email.com", "pass1", "user1", "addr1"),
		types.NewUser(2, "2@email.com", "pass2", "user2", "addr2"),
		types.NewUser(3, "3@email.com", "pass3", "user3", "addr3"),
		types.NewUser(4, "4@email.com", "pass4", "user4", "addr4"),
		types.NewUser(5, "5@email.com", "pass5", "user5", "addr5"),
		types.NewUser(6, "6@email.com", "pass6", "user6", "addr6"),
		types.NewUser(7, "7@email.com", "pass7", "user7", "addr7"),
		types.NewUser(8, "8@email.com", "pass8", "user8", "addr8"),
	}

	// model a whole game process
	h := types.NewTournamentHall(1<<3, 5, 3)

	// apply
	log.Println("accepting applications...")
	for _, u := range candidates {
		if err := h.Apply(u); err != nil {
			log.Panicln(err)
		}
		if h.IsFull() {
			h.SetStatPending()
			log.Println("waiting 5 seconds for game start...")
			time.Sleep(5 * time.Second)
			break
		}
	}

	// game start
	log.Println("1st round game start...")
	h.SetStatInGame()

	// 1st round
	h.SetWinnerLoser(100001, 1) // should always get winner, loser, table id from game server
	h.SetWinnerLoser(100002, 4)
	h.SetWinnerLoser(100003, 6)
	h.SetWinnerLoser(100004, 7)
	log.Printf("1st round winners: %d %d %d %d\n\n", 1, 4, 6, 7)

	// 2nd allocate
	log.Println("allocating for 2nd round...")
	if err := h.Allocate(candidates[0]); err != nil {
		log.Panicln(err)
	}
	if err := h.Allocate(candidates[3]); err != nil {
		log.Panicln(err)
	}
	if err := h.Allocate(candidates[5]); err != nil {
		log.Panicln(err)
	}
	if err := h.Allocate(candidates[6]); err != nil {
		log.Panicln(err)
	}
	log.Println("done allocation")
	h.SetStatPending()
	log.Println("waiting 5 seconds for 2nd round game start...")
	time.Sleep(5 * time.Second)

	// 2nd start
	log.Println("2nd round game start")
	h.SetStatInGame()

	// 2nd round
	h.SetWinnerLoser(200001, 1)
	h.SetWinnerLoser(200002, 6)
	log.Printf("2nd round winners: %d %d\n\n", 1, 6)

	// 3rd round allocate
	log.Println("allocating for 3rd round...")
	if err := h.Allocate(candidates[0]); err != nil {
		log.Panicln(err)
	}
	if err := h.Allocate(candidates[5]); err != nil {
		log.Panicln(err)
	}
	log.Println("done allocation")
	h.SetStatPending()
	log.Println("waiting 5 seconds for 3rd round game start...")
	time.Sleep(5 * time.Second)

	// 3rd start
	log.Println("3rd round game start")
	h.SetStatInGame()

	// 3rd round result
	h.SetWinnerLoser(300001, 1)
	log.Printf("3rd round winners: %d\n\n", 1)

	// set gold, silver
	h.SetGold("user1")
	h.SetSilver("user6")

	for tid, t := range h.Tables.Tables {
		log.Printf("table id %d, 1p: %v, 2p: %v", tid, t.Get1p2p())
	}

	if h.ShouldEnd() {
		h.SetStatEnd()
		log.Println("the tournament is end")
	}

	b, err := json.Marshal(h)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))
}
