package tetris

import "fmt"

type dot struct {
	x, y  int
	Color Color
}

func newDot(x, y int, Color Color) dot {
	return dot{x: x, y: y, Color: Color}
}

func (d dot) String() string {
	return fmt.Sprintf("x: %d, y: %d, Color: %v", d.x, d.y, d.Color)
}

func (d dot) rotate(origin dot) dot {
	return rotate(d, origin)
}

func (d dot) moveLeft() dot {
	d.x -= 1
	return d
}

func (d dot) moveRight() dot {
	d.x += 1
	return d
}

func (d dot) moveDown() dot {
	d.y += 1
	return d
}

func (d dot) moveUp() dot {
	d.y -= 1
	return d
}

func (d dot) isOverlapped(d2 dot) bool {
	return isOverlapped(d, d2)
}

func (d dot) isContiguous(d2 dot) bool {
	return isContiguous(d, d2)
}
