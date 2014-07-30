// Color is not "Color", include other stuffs
// like it may represent "nothing", "stone", "bomb", "transparent-Color"
package tetris

import "fmt"

type Color int

// any use? currently no. may deprecated in the future
func newColor(c int) Color {
	return Color(c % (maxColor + 1))
}

func randomColor() Color {
	return Color(randSeed.Intn(maxColor) + 1)
}

func (c Color) String() string {
	return fmt.Sprintf("%v", int(c))
}

func (c Color) toTransparent() Color {
	return Color(-int(c))
}

func (c Color) isStone() bool {
	return int(c) == Color_stone
}

func (c Color) isNothing() bool {
	return int(c) == Color_nothing
}

func (c Color) isTransparent(oc Color) bool {
	return int(c) == -int(oc)
}

func (c Color) isBomb() bool {
	return int(c) == Color_bomb
}

const (
	maxColor = 6

	Color_nothing = 0
	Color_stone   = -99
	Color_bomb    = -98
)

// Colors
var Colors = map[int]string{
	Color_nothing: "nothing",
	Color_stone:   "stone",
	Color_bomb:    "bomb",

	1:  "black",
	2:  "red",
	3:  "green",
	4:  "blue",
	5:  "yellow",
	6:  "pink",
	7:  "purple",
	-7: "transparent-purple", // the negative value represents the transparent Color
	-6: "transparent-pink",   // ...
}
