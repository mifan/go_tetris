package tetris

import (
	"math/rand"
	"time"
)

const defaultNumOfDotsInABlock = 4

var randSeed = rand.New(rand.NewSource(time.Now().UnixNano()))

// rotate dot d 90 degree by dot origin counter-clockwise
func rotate(d, origin dot) dot {
	return newDot(d.y-origin.y+origin.x, origin.x+origin.y-d.x, d.Color)
}

// check if two dots overlap
func isOverlapped(d1, d2 dot) bool {
	return d1.x == d2.x && d1.y == d2.y
}

// check if two dots contiguous
func isContiguous(d1, d2 dot) bool {
	return isOverlapped(d1.moveUp(), d2) ||
		isOverlapped(d1.moveDown(), d2) ||
		isOverlapped(d1.moveLeft(), d2) ||
		isOverlapped(d1.moveRight(), d2)
}

// generate random dot
func randomDot(Color Color) dot {
	x := randSeed.Intn(defaultNumOfDotsInABlock)
	maxY := defaultNumOfDotsInABlock - x
	y := randSeed.Intn(maxY)
	return newDot(x, y, Color)
}
