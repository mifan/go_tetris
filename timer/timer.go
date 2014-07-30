// could be a better timer, pause supported
// may not be accurate, but fair enough for me
package timer

import (
	"sync"
	"time"
)

const (
	defaultInterval = 500
	tickFrequency   = 10
)

func i2Duration(i int) time.Duration {
	return time.Millisecond * time.Duration(i)
}

type Timer struct {
	sync.Mutex
	timerInterval, currentTick int // in ms
	ticker                     *time.Ticker
	isPaused                   bool
	tick                       chan bool
}

func NewTimer(intervalInMs ...int) *Timer {
	var interval int = defaultInterval
	if len(intervalInMs) > 0 {
		interval = intervalInMs[0]
	}
	t := &Timer{
		timerInterval: interval,
		currentTick:   tickFrequency,
		ticker:        time.NewTicker(i2Duration(tickFrequency)),
		tick:          make(chan bool),
		isPaused:      true,
	}
	return t.init()
}

func (t *Timer) init() *Timer {
	go t.startTick()
	return t
}

func (t *Timer) setTick() {
	t.Lock()
	defer t.Unlock()
	if !t.isPaused {
		t.currentTick = (t.currentTick + tickFrequency) % t.timerInterval
	}
}

func (t *Timer) shouldTick() bool {
	t.Lock()
	defer t.Unlock()
	return !t.isPaused && t.currentTick == 0
}

func (t *Timer) startTick() {
	for {
		select {
		case <-t.ticker.C:
			t.setTick()
			if t.shouldTick() {
				t.tick <- true
			}
		}
	}
}

// pause
func (t *Timer) Pause() {
	t.Lock()
	defer t.Unlock()
	t.isPaused = true
}

// break pause
func (t *Timer) Start() {
	t.Lock()
	defer t.Unlock()
	t.isPaused = false
}

// reset
func (t *Timer) Reset() {
	t.Lock()
	defer t.Unlock()
	t.currentTick = tickFrequency
}

// wait for next tick
func (t *Timer) Wait() {
	<-t.tick
}
