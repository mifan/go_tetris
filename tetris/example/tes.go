package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gogames/go_tetris/tetris"
)

var g *tetris.Game

func init() {
	var err error
	g, err = tetris.NewGame(20, 10, 5, 500)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	g.Start()
}

func main() {
	go handleInput()
	go attack()
	for {
		d := g.GetData()
		switch d.Description {
		case "zone":
			renderScreen(d.Val.(tetris.ZoneData))
		default:
			b, _ := json.Marshal(d.Val)
			log.Printf("%s: %v", d.Description, string(b))
		}
	}
}

func attack() {
	for {
		time.Sleep(5 * time.Second)
		g.BeingAttacked(1)
	}
}

func renderScreen(z tetris.ZoneData) {
	b, _ := json.Marshal(map[string]interface{}{
		"zone": z,
	})
	fmt.Println(len(b))

	// do something
	for i := 0; i < len(z[0]); i++ {
		fmt.Print("-")
	}

	fmt.Println()

	for _, l := range z {
		lstr := ""
		for _, c := range l {
			switch cc := int(c); cc {
			case tetris.Color_nothing:
				lstr += " "
			case tetris.Color_bomb:
				lstr += "*"
			case tetris.Color_stone:
				lstr += "#"
			default:
				if cc > 0 {
					lstr += fmt.Sprintf("%v", cc)
				} else {
					lstr += "p"
				}
			}
		}
		fmt.Println("|", lstr, "|")
	}
	for i := 0; i < len(z[0]); i++ {
		fmt.Print("-")
	}
	fmt.Println()

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

// func fetchNextPiece() {
// 	for {
// 		piece := game.NextPiece()
// 		for _, l := range piece.Dots() {
// 			lstr := ""
// 			for _, c := range l {
// 				if c {
// 					lstr += "#"
// 				} else {
// 					lstr += " "
// 				}
// 			}
// 			fmt.Println("|", lstr, "|")
// 		}
// 	}
// }
//
// func fetchGameScore() {
// 	var score int
// 	for {
// 		score = game.Score()
// 		// do something with score
// 		log.Println("score: ", score)
// 	}
// }
//
// func fetchComboScore() {
// 	var combo int
// 	for {
// 		combo = game.ComboScore()
// 		// do something with score
// 		log.Println("combo: ", combo)
// 	}
// }
//
// func handleGameOver() {
// 	// game over signal is a channel, so only call it once
// 	if game.IsGameOver() {
// 		log.Println("game is over")
// 		os.Exit(1)
// 	}
// }
//
func handleInput() {
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	for {
		key := make([]byte, 1)
		os.Stdin.Read(key)
		switch string(key) {
		case "j":
			g.MoveLeft()
		case "l":
			g.MoveRight()
		case "k":
			g.MoveDown()
		case " ":
			g.DropDown()
		case "i":
			g.Rotate()
		case "r":
			g.Reserve()
		}
	}
}

//
// func main() {
// 	lock := make(chan bool)
// 	go fetchGameScore()
// 	go fetchComboScore()
// 	go fetchGameScreen()
// 	go fetchNextPiece()
// 	go handleGameOver()
// 	go handleInput()
// 	<-lock
// }
