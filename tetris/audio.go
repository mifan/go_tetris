// define some audio effects
package tetris

import (
	"encoding/json"
	"fmt"
)

const (
	backgroundAudio = -1
	bomb            = -2
	ko              = -3
)

const (
	backgroudAudioEffect = "background.avi" // start background audio effect
	bombAudioEffect      = "bomb.avi"       // play bomb.avi audio effect
	koAudioEffect        = "ko.avi"         // play ko.avi audio effect
	comboAudioEffect     = "combo%v.avi"    // play combo1.avi combo2.avi according to # of lines sent to oppenent
)

type audio int

func audioBackground() audio {
	return audio(backgroundAudio)
}

func audioCombo(lines int) audio {
	return audio(lines)
}

func audioHitBomb() audio {
	return audio(bomb)
}

func audioKO() audio {
	return audio(ko)
}

var _ json.Marshaler = audio(backgroundAudio)

func (a audio) MarshalJSON() (b []byte, err error) {
	var v string
	switch intA := int(a); intA {
	case backgroundAudio:
		v = backgroudAudioEffect
	case bomb:
		v = bombAudioEffect
	case ko:
		v = koAudioEffect
	default:
		v = fmt.Sprintf(comboAudioEffect, a)
	}
	return json.Marshal(v)
}
