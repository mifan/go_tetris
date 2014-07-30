// game controller clarify how the game runs
package tetris

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gogames/go_tetris/timer"
)

const (
	buffer    = 1 << 10
	minWidth  = defaultNumOfDotsInABlock
	minHeight = defaultNumOfDotsInABlock
)

var (
	errWidth  = fmt.Errorf("width should be larger than %v", minWidth)
	errHeight = fmt.Errorf("height should be larger than %v", minHeight)
)

type nextPieces struct{ *ring.Ring }

var (
	_ json.Marshaler = newNextPieces(2)
)

func newNextPieces(numOfNext int) nextPieces {
	return nextPieces{ring.New(numOfNext)}
}

func (np nextPieces) addNewPiece(p *piece) {
	np.Value = p
	np.Ring = np.Next()
}

func (np nextPieces) getOne(newP *piece) *piece {
	p := np.Value.(*piece)
	np.Value = newP
	np.Move(1)
	return p
}

func (np nextPieces) MarshalJSON() ([]byte, error) {
	v := make([]interface{}, np.Len())
	for i := 0; i < np.Len(); i++ {
		v[i] = np.Value
		np.Ring = np.Next()
	}
	return json.Marshal(v)
}

type Game struct {
	sync.Mutex

	// main game zone
	mainZone mainZone

	// the timer
	timer *timer.Timer

	// the pieces
	activePiece *piece
	holdPiece   *piece
	holded      bool
	nextPieces  nextPieces

	// chan
	MsgChan      chan message // directly send to flash client
	AttackChan   chan int
	GameoverChan chan bool
	BeingKOChan  chan bool

	// score
	numOfLineSent, combo, ko int
}

func NewGame(height, width, numOfNextPieces, interval int) (*Game, error) {
	if width < minWidth {
		return nil, errWidth
	}
	if height < minHeight {
		return nil, errHeight
	}
	np := newNextPieces(numOfNextPieces)
	for numOfNextPieces > 0 {
		numOfNextPieces--
		np.addNewPiece(newPiece(width/2 - 2))
		np.Ring = np.Next()
	}
	g := &Game{
		mainZone:     newMainZone(height, width),
		timer:        timer.NewTimer(interval),
		activePiece:  newPiece(width/2 - 2),
		holdPiece:    nil,
		holded:       false,
		nextPieces:   np,
		MsgChan:      make(chan message, buffer),
		GameoverChan: make(chan bool),
		AttackChan:   make(chan int, buffer),
		BeingKOChan:  make(chan bool, 5),
	}
	go g.init()
	return g, nil
}

func (g *Game) init() {
	for {
		g.timer.Wait()
		g.check(true, false)
	}
}

func (g *Game) KoOpponent() {
	g.ko++
	g.send(DescKo, g.ko)
	g.send(DescAudio, audioKO)
}

// params:
// 1. should move down one step
// 2. should drop down to the bottom
// 3. should reset the timer
func (g *Game) check(moveDown, dropDown bool) {
	g.Lock()
	defer g.Unlock()

	g.mainZone.toZoneData()

	var genNewPiece bool
	switch {
	case moveDown:
		if g.mainZone.canBlockMoveDown(g.activePiece.block) {
			g.activePiece.block = g.activePiece.moveDown()
			break
		}
		genNewPiece = true
	case dropDown:
		g.activePiece.block = g.mainZone.dropBlockOnZone(g.activePiece.block)
		genNewPiece = true
	}

	if genNewPiece {
		g.holded = false

		g.mainZone.putBlockOnMainZone(g.activePiece.block)
		if lineSent := g.calculate(); lineSent > 0 {
			g.scoreAdd(lineSent)
			g.AttackChan <- lineSent
			g.send(DescAttack, lineSent)
			g.send(DescLines, g.numOfLineSent)
		}

		g.activePiece = g.nextPieces.getOne(newPiece(g.mainZone.width()/2 - 2))

		g.send(DescNextPiece, g.nextPieces)
	}

	// if it is dropDown or moveDown, the timer should be reset
	if dropDown || moveDown {
		g.timer.Reset()
	}

	// if being ko
	if g.mainZone.beingKO() {
		g.BeingKOChan <- true
		g.mainZone.removeStoneLines()
	}

	// render new zone
	if g.mainZone.canPutBlock(g.activePiece.block) {
		g.send(DescZone, g.mainZone.toZoneData().
			renderProjectionOfBlockOnZone(g.activePiece.block).
			renderBlockOnZone(g.activePiece.block))
	}
}

// get data
func (g *Game) GetData() message {
	return <-g.MsgChan
}

// get number of ko
func (g *Game) GetKo() int {
	return g.ko
}

// get score
func (g *Game) GetScore() int {
	return g.numOfLineSent
}

// move down
func (g *Game) MoveDown() {
	g.check(true, false)
}

// drop down
func (g *Game) DropDown() {
	g.check(false, true)
}

// move left
func (g *Game) MoveLeft() {
	func() {
		g.Lock()
		defer g.Unlock()
		if g.mainZone.toZoneData().canBlockMoveLeft(g.activePiece.block) {
			g.activePiece.block = g.activePiece.block.moveLeft()
		}
	}()
	g.check(false, false)
}

// move right
func (g *Game) MoveRight() {
	func() {
		g.Lock()
		defer g.Unlock()
		if g.mainZone.toZoneData().canBlockMoveRight(g.activePiece.block) {
			g.activePiece.block = g.activePiece.block.moveRight()
		}
	}()
	g.check(false, false)
}

// rotate
func (g *Game) Rotate() {
	func() {
		g.Lock()
		defer g.Unlock()
		if b, can := g.mainZone.toZoneData().canBlockRotate(g.activePiece.block); can {
			g.activePiece.block = b
		}
	}()
	g.check(false, false)
}

// hold
func (g *Game) Hold() {
	func() {
		g.Lock()
		defer g.Unlock()
		if !g.canHold() {
			return
		}
		g.holded = true
		if g.holdPiece == nil {
			g.holdPiece, g.activePiece = g.activePiece, g.nextPieces.getOne(newPiece(g.mainZone.width()/2-2))
			return
		}
		g.activePiece, g.holdPiece = g.holdPiece, g.activePiece
		g.activePiece.block = g.activePiece.resPosition
		g.send(DescHoldedPiece, g.holdPiece)
	}()
	g.check(false, false)
}

// being attacked, return true if being KO
func (g *Game) BeingAttacked(n int) {
	if ko := func() bool {
		g.Lock()
		defer g.Unlock()
		if g.mainZone.canFilledStoneLines(n) {
			g.mainZone.addStoneLines(n)
			return false
		}
		g.mainZone.removeStoneLines()
		return true
	}(); ko {
		g.BeingKOChan <- true
	}
	g.check(false, false)
}

// start the game
func (g *Game) Start() {
	g.timer.Start()
	g.send(DescAudio, audioBackground())
}

// pause the game
func (g *Game) Pause() {
	g.timer.Pause()
	g.send(DescPause, true)
	g.send(DescAudio, audioBackground())
}

// stop the game
func (g *Game) Stop() {
	g.timer.Pause()
}

// end the game
func (g *Game) End() {
	g.timer.Pause()
	g.send(DescOver, true)
	g.GameoverChan <- true
}

// combo add one
func (g *Game) comboAdd() {
	g.combo++
}

// combo reset
func (g *Game) comboReset() {
	g.combo = 0
}

// combo attack ?
func (g *Game) comboAttack() (combo int) {
	switch c := g.combo - 1; {
	case c <= 0:
		combo = 0
	case c <= 2:
		combo = 1
	case c <= 4:
		combo = 2
	case c <= 6:
		combo = 3
	default:
		combo = 4
	}
	return
}

// score add
func (g *Game) scoreAdd(n int) {
	g.numOfLineSent += n
}

// check if it is able to hold the current block
func (g *Game) canHold() bool {
	return !g.holded
}

// send MsgChan
func (g *Game) send(desc string, val interface{}) {
	g.MsgChan <- NewMessage(desc, val)
}

// calculate score
// score = bomb + clear_lines + combo + if_zone_clear_then_10
func (g *Game) calculate() (lineSent int) {
	// num of bombs hit and lines clear
	hitBombs := g.mainZone.checkHitBombs(g.activePiece.block)
	if hitBombs > 0 {
		g.send(DescAudio, audioHitBomb())
	}
	l := g.mainZone.clearLines()

	// clear
	if g.mainZone.isClear() {
		lineSent += 10
		g.send(DescClear, true)
	}

	// not combo, reset combo, return
	if (l + hitBombs + lineSent) <= 1 {
		g.comboReset()
		return
	}

	// combo
	g.comboAdd()
	if c := g.comboAttack(); c > 0 {
		lineSent += c
		g.send(DescCombo, g.combo)
		g.send(DescAudio, audioCombo(c))
	}
	if 0 < l && l < 4 {
		l--
	}

	// num of lines should sent to opponent
	lineSent += l + hitBombs
	return
}
