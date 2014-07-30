package tetris

import "encoding/json"

type message struct {
	Description string
	Val         interface{}
}

var _ json.Marshaler = message{}

func (d message) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		d.Description: d.Val,
	})
}

func NewMessage(desc string, val interface{}) message {
	return message{
		Description: desc,
		Val:         val,
	}
}

// descriptions
const (
	DescNextPiece   = "next"     // next piece change
	DescHoldedPiece = "hold"     // hold piece change
	DescZone        = "zone"     // zone change
	DescAudio       = "audio"    // audio play
	DescAttack      = "attack"   // send lines to attack opponent (send line, or T Z spin)
	DescLines       = "lines"    // number of send lines changed
	DescCombo       = "combo"    // combo number changed
	DescKo          = "ko"       // ko the opponent
	DescBeingKo     = "beingKo"  // ko by the opponent
	DescStart       = "start"    // game start, count 3 seconds
	DescPause       = "pause"    // game pause
	DescOver        = "gameover" // game over
	DescClear       = "clear"    // game zone clear
)
