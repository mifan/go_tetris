package tetris

import "fmt"

type line []Color

const maxNumOfBombsInALine = 2

func newLine(length int, c Color) line {
	l := make(line, length)
	for length > 0 {
		length--
		l[length] = c
	}
	return l
}

func newClearLine(length int) line {
	return newLine(length, Color_nothing)
}

func newBombLine(length int) line {
	return newLine(length, Color_stone).placeBomb()
}

func (l line) String() (res string) {
	res += "["
	for _, v := range l {
		res += fmt.Sprintf("%v, ", v)
	}
	res += "]"
	return
}

// is clear line
func (l line) isClear() bool {
	for _, v := range l {
		if !v.isNothing() {
			return false
		}
	}
	return true
}

// get a copy of the line
func (l line) getLine() line {
	nl := make(line, len(l))
	copy(nl, l)
	return nl
}

// get the length of the line
func (l line) length() int {
	return len(l)
}

// lines to [width]Color
func (l line) toArray() []Color {
	data := make([]Color, l.length())
	for i, _ := range data {
		data[i] = l[i]
	}
	return data
}

// contains any active dot
func (l line) containAnyActiveDot() bool {
	for _, v := range l {
		if v > 0 {
			return true
		}
	}
	return false
}

// check if the line is a stone line, by checking if it contains any stone in the line
func (l line) isStoneLine() bool {
	for _, v := range l {
		if v.isStone() {
			return true
		}
	}
	return false
}

// check if the line can be clear, by checking if all dot are active
func (l line) canClear() bool {
	for _, v := range l {
		if v <= 0 {
			return false
		}
	}
	return true
}

// place a dot on the line
func (l line) placeDots(index int, Color Color) line {
	l[index] = Color
	return l
}

// place bombs(1~2) on the line
func (l line) placeBomb() line {
	for i := 0; i < maxNumOfBombsInALine; i++ {
		l.placeDots(randSeed.Intn(l.length()), Color_bomb)
	}
	return l
}
