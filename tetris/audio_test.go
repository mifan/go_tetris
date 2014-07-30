package tetris

import (
	"encoding/json"
	"testing"
)

func Test_Audio(t *testing.T) {
	type a struct{ audio }
	v := a{audioBackground()}
	b, _ := json.Marshal(v)
	t.Log(string(b))
}
