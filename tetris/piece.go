package tetris

import "fmt"

func newPiece(mid int) *piece {
	originBlock := newBlock()
	block := originBlock
	for ; mid > 0; mid-- {
		block = block.moveRight()
	}
	return &piece{
		block:       block,
		originBlock: originBlock,
		resPosition: block,
	}
}

type piece struct {
	block
	originBlock block
	resPosition block
}

func (p piece) String() string {
	return fmt.Sprintf("\nblock: %v\nColor: %v\n", p.block, p.Color())
}

func (p piece) MarshalJSON() ([]byte, error) {
	return p.originBlock.MarshalJSON()
}
