package vtime

import (
	"github.com/pion/transport/v3/xtime"
	"time"
)

type ticker struct {
	c       chan xtime.Tick
	stopped bool
}

func (s *Simulator) NewTicker(duration time.Duration) xtime.Ticker {
	t := &ticker{
		c:       make(chan xtime.Tick),
		stopped: false,
	}
	var do func()
	do = func() {
		if t.stopped {
			return
		}
		tick := xtime.Tick{
			Done: make(chan struct{}),
			Time: s.now,
		}
		t.c <- tick
		<-tick.Done
		s.cond.L.Lock()
		s.pushEvent(s.now.Add(duration), do)
		s.cond.L.Unlock()
		s.cond.Signal()
	}

	s.cond.L.Lock()
	s.pushEvent(s.now.Add(duration), do)
	s.cond.L.Unlock()
	s.cond.Signal()
	return t
}

func (t *ticker) C() <-chan xtime.Tick {
	return t.c
}

func (t *ticker) Stop() {
	t.stopped = true
}
