package tetris

import (
	"encoding/json"
	"fmt"
)

type block [defaultNumOfDotsInABlock]dot

var _ json.Marshaler = block{}

var blocks = []block{
	block([defaultNumOfDotsInABlock]dot{
		newDot(0, 0, newColor(1)),
		newDot(1, 0, newColor(1)),
		newDot(2, 0, newColor(1)),
		newDot(3, 0, newColor(1)),
	}),
	block([defaultNumOfDotsInABlock]dot{
		newDot(0, 0, newColor(2)),
		newDot(0, 1, newColor(2)),
		newDot(1, 1, newColor(2)),
		newDot(2, 1, newColor(2)),
	}),
	block([defaultNumOfDotsInABlock]dot{
		newDot(2, 0, newColor(3)),
		newDot(0, 1, newColor(3)),
		newDot(1, 1, newColor(3)),
		newDot(2, 1, newColor(3)),
	}),
	block([defaultNumOfDotsInABlock]dot{
		newDot(1, 0, newColor(4)),
		newDot(0, 1, newColor(4)),
		newDot(1, 1, newColor(4)),
		newDot(2, 1, newColor(4)),
	}),
	block([defaultNumOfDotsInABlock]dot{
		newDot(0, 0, newColor(5)),
		newDot(1, 0, newColor(5)),
		newDot(1, 1, newColor(5)),
		newDot(2, 1, newColor(5)),
	}),
	block([defaultNumOfDotsInABlock]dot{
		newDot(1, 0, newColor(6)),
		newDot(2, 0, newColor(6)),
		newDot(0, 1, newColor(6)),
		newDot(1, 1, newColor(6)),
	}),
}

func newBlock() block {
	return blocks[randSeed.Intn(maxColor)]
}

// implement json marshaler interface for rendering the reserved piece, next several pieces
func (b block) MarshalJSON() ([]byte, error) {
	var v [defaultNumOfDotsInABlock][defaultNumOfDotsInABlock]Color
	for _, d := range b {
		v[d.y][d.x] = b.Color()
	}
	return json.Marshal(v)
}

func (b block) String() (str string) {
	str += "\n"
	for _, d := range b {
		str += fmt.Sprintf("%v\n", d)
	}
	return
}

func (b block) toDots() []dot {
	ds := make([]dot, defaultNumOfDotsInABlock)
	for i, v := range b {
		ds[i] = v
	}
	return ds
}

func (b *block) moveRight() block {
	for i, v := range b {
		b[i] = v.moveRight()
	}
	return *b
}

func (b *block) moveLeft() block {
	for i, v := range b {
		b[i] = v.moveLeft()
	}
	return *b
}

func (b *block) moveDown() block {
	for i, v := range b {
		b[i] = v.moveDown()
	}
	return *b
}

func (b *block) moveUp() block {
	for i, v := range b {
		b[i] = v.moveUp()
	}
	return *b
}

func (b block) Color() Color {
	return b[0].Color
}

func (b *block) transparentBlock() {
	for i, _ := range b {
		b[i].Color = b[i].Color.toTransparent()
	}
}

func (b block) center() dot {
	var x, y int
	for _, v := range b {
		x += v.x*2 + 1
		y += v.y*2 + 1
	}
	return newDot(x/2/defaultNumOfDotsInABlock, y/2/defaultNumOfDotsInABlock, b[0].Color)
}

func (b *block) rotate() block {
	center := b.center()
	for i, v := range b {
		b[i] = v.rotate(center)
	}
	return *b
}

func (b block) outBoundLeft(bound int) bool {
	for _, v := range b {
		if v.x < bound {
			return true
		}
	}
	return false
}

func (b block) outBoundRight(bound int) bool {
	for _, v := range b {
		if v.x > bound {
			return true
		}
	}
	return false
}

func (b block) outBoundTop(bound int) bool {
	for _, v := range b {
		if v.y < bound {
			return true
		}
	}
	return false
}

func (b block) outBoundButtom(bound int) bool {
	for _, v := range b {
		if v.y > bound {
			return true
		}
	}
	return false
}
