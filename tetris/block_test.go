package tetris

import "testing"

func Test_Block(t *testing.T) {
	b := newBlock()
	t.Log(b)
	c := (&b).rotate()
	t.Log(b)
	t.Log(c)
}
