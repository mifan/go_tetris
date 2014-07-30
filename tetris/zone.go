// mainZone of game
// a list of lines
package tetris

import "container/list"

type (
	ZoneData [][]Color
	mainZone struct {
		*list.List
		ZoneData
	}
)

func newZoneData(height, width int) ZoneData {
	ZoneData := make([][]Color, height)
	for i, _ := range ZoneData {
		ZoneData[i] = make([]Color, width)
	}
	return ZoneData
}

func newMainZone(height, width int) mainZone {
	l := list.New()
	for i := 0; i < height; i++ {
		l.PushBack(newClearLine(width))
	}
	return mainZone{
		List:     l,
		ZoneData: newZoneData(height, width),
	}
}

// convert mainZone lines to ZoneData
func (m mainZone) toZoneData() ZoneData {
	m.ZoneData.clear()
	i := 0
	for e := m.Front(); e != nil; e = e.Next() {
		m.ZoneData[i] = e.Value.(line).toArray()
		i++
	}
	return m.ZoneData
}

// get line by height
func (m mainZone) getLineByHeight(h int) (l line) {
	var i = 0
	for e := m.Front(); e != nil; e = e.Next() {
		if i == h {
			return e.Value.(line)
		}
		i++
	}
	return
}

// check if it is clear
func (m mainZone) isClear() bool {
	for e := m.Front(); e != nil; e = e.Next() {
		if !e.Value.(line).isClear() {
			return false
		}
	}
	return true
}

// check if the block can be put on zone
func (m mainZone) canPutBlock(b block) bool {
	for _, v := range b {
		if !m.ZoneData[v.y][v.x].isNothing() {
			return false
		}
	}
	return true
}

// place the block into main zone
func (m mainZone) putBlockOnMainZone(b block) {
	for _, v := range b {
		m.getLineByHeight(v.y).placeDots(v.x, v.Color)
	}
}

// drop a block on main zone, return the last location of the block
func (m mainZone) dropBlockOnZone(b block) block {
	zone := m.toZoneData()
	for zone.canBlockMoveDown(b) {
		b = b.moveDown()
	}
	return b
}

// check hit bombs, returns the lines it clears
func (m mainZone) checkHitBombs(b block) int {
	i := m.toZoneData().hitBomb(b.toDots())
	m.removeBombHitLines(i)
	return i
}

// remove the bomb hit lines(the first n stone lines)
func (m mainZone) removeBombHitLines(n int) {
	if n <= 0 {
		return
	}
	v := n
	for e := m.Front(); e != nil; {
		if !e.Value.(line).isStoneLine() {
			e = e.Next()
			continue
		}
		if e.Next() == nil {
			m.Remove(e)
			break
		} else {
			e = e.Next()
			m.Remove(e.Prev())
		}
		n--
		if n <= 0 {
			break
		}
	}
	for v > 0 {
		m.PushFront(newClearLine(m.width()))
		v--
	}
}

// remove the lines and add clear line in the front
func (m mainZone) clearLines() (lines int) {
	for e := m.Front(); e != nil; {
		if e.Value.(line).isStoneLine() {
			break
		}
		if !e.Value.(line).canClear() {
			e = e.Next()
			continue
		}
		lines++
		if e.Next() == nil {
			m.Remove(e)
			m.PushFront(newClearLine(m.width()))
			break
		}
		e = e.Next()
		m.Remove(e.Prev())
		m.PushFront(newClearLine(m.width()))
	}
	return
}

// check if it is able to filled n stone lines
func (m mainZone) canFilledStoneLines(n int) bool {
	for e := m.Front(); e != nil; e = e.Next() {
		l := e.Value.(line)
		if !l.isClear() {
			break
		}
		n--
		if n <= 0 {
			return true
		}
	}
	return false
}

// add stone lines and remove the clear lines
func (m mainZone) addStoneLines(n int) {
	for n > 0 {
		n--
		m.Remove(m.Front())
		m.PushBack(newBombLine(m.width()))
	}
}

// remove stone lines
func (m mainZone) removeStoneLines() {
	var i = 0
	for e := m.Back(); e != nil; {
		if !e.Value.(line).isStoneLine() {
			break
		}
		i++
		if e.Prev() == nil {
			m.Remove(e)
			break
		}
		e = e.Prev()
		m.Remove(e.Next())
	}
	for i > 0 {
		m.PushFront(newClearLine(m.width()))
		i--
	}
}

// being ko ?
func (m mainZone) beingKO() bool {
	return m.Front().Value.(line).containAnyActiveDot()
}

// clear the ZoneData
func (zone ZoneData) clear() ZoneData {
	for h, l := range zone {
		for w, _ := range l {
			zone[h][w] = newColor(Color_nothing)
		}
	}
	return zone
}

// render a block on zone
func (zone ZoneData) renderBlockOnZone(b block) ZoneData {
	for _, v := range b {
		zone[v.y][v.x] = b.Color()
	}
	return zone
}

// render projection of the block on zone
func (zone ZoneData) renderProjectionOfBlockOnZone(b block) ZoneData {
	for zone.canBlockMoveDown(b) {
		b = b.moveDown()
	}
	(&b).transparentBlock()
	return zone.renderBlockOnZone(b)
}

// check if the block hit the bomb
func (zone ZoneData) hitBomb(ds []dot) (lines int) {
	for _, v := range ds {
		var depth int
		if tmpDs := zone.canTraverse(v); tmpDs != nil {
			depth += 1 + zone.hitBomb(tmpDs)
		}
		if lines < depth {
			lines = depth
		}
	}
	return
}

func (zone ZoneData) canTraverse(d dot) []dot {
	if d.y < 0 || d.y >= zone.height()-1 {
		return nil
	}
	if d.x < 0 || d.x > zone.width()-1 {
		return nil
	}
	if zone[d.y+1][d.x].isBomb() {
		ds := make([]dot, 0)
		tmpDot := newDot(d.x, d.y+1, zone[d.y+1][d.x])
		ds = append(ds, tmpDot)
		for tmpDot.x > 0 && zone[tmpDot.y][tmpDot.x-1].isBomb() {
			tmpDot = newDot(tmpDot.x-1, tmpDot.y, zone[tmpDot.y][tmpDot.x-1])
			ds = append(ds, tmpDot)
		}
		tmpDot = newDot(d.x, d.y+1, zone[d.y+1][d.x])
		for tmpDot.x < zone.width()-1 && zone[tmpDot.y][tmpDot.x+1].isBomb() {
			tmpDot = newDot(tmpDot.x+1, tmpDot.y, zone[tmpDot.y][tmpDot.x+1])
			ds = append(ds, tmpDot)
		}
		return ds
	}
	return nil
}

// check if a block can move down
func (zone ZoneData) canBlockMoveDown(b block) bool {
	for _, v := range b {
		if v.y >= zone.height()-1 || !zone[v.y+1][v.x].isNothing() || !zone[v.y+1][v.x].isTransparent(zone[v.y][v.x]) {
			return false
		}
	}
	return true
}

// check if a block can move right
func (zone ZoneData) canBlockMoveRight(b block) bool {
	for _, v := range b {
		if v.x >= zone.width()-1 || !zone[v.y][v.x+1].isNothing() {
			return false
		}
	}
	return true
}

// check if a block can move left
func (zone ZoneData) canBlockMoveLeft(b block) bool {
	for _, v := range b {
		if v.x <= 0 || !zone[v.y][v.x-1].isNothing() {
			return false
		}
	}
	return true
}

// check if a block can rotate
func (zone ZoneData) canBlockRotate(b block) (block, bool) {
	b = b.rotate()
	// better op
	for b.outBoundTop(0) {
		b = b.moveDown()
	}
	for b.outBoundButtom(zone.height() - 1) {
		b = b.moveUp()
	}
	for b.outBoundLeft(0) {
		b = b.moveRight()
	}
	for b.outBoundRight(zone.width() - 1) {
		b = b.moveLeft()
	}
	for _, v := range b {
		if !zone[v.y][v.x].isNothing() {
			return b, false
		}
	}
	return b, true
}

func (zone ZoneData) height() int {
	return len(zone)
}

func (zone ZoneData) width() int {
	return len(zone[0])
}
