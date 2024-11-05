package vtime

import (
	"github.com/pion/transport/v3/xtime"
	"time"
)

type timer struct {
	c         chan xtime.Tick
	expiresAt time.Time
	simulator *Simulator
}

func (s *Simulator) NewTimer(d time.Duration) xtime.Timer {
	t := &timer{
		c:         make(chan xtime.Tick),
		expiresAt: time.Time{},
		simulator: s,
	}
	t.Reset(d)
	return t
}

func (s *Simulator) After(d time.Duration) <-chan xtime.Tick {
	return s.NewTicker(d).C()
}

func (t *timer) C() <-chan xtime.Tick {
	return t.c
}

func (t *timer) Stop() bool {
	t.simulator.cond.L.Lock()

	t.expiresAt = time.Time{}
	wasSet := !t.expiresAt.IsZero()

	t.simulator.cond.L.Unlock()
	return wasSet
}

func (t *timer) Reset(duration time.Duration) bool {
	t.simulator.cond.L.Lock()

	newExpiresAt := t.simulator.now.Add(duration)
	wasReset := t.expiresAt.Before(newExpiresAt)
	t.expiresAt = t.simulator.now.Add(duration)

	t.simulator.pushEvent(t.expiresAt, func() {
		t.simulator.cond.L.Lock()
		if !t.expiresAt.Equal(t.simulator.now) {
			t.simulator.cond.L.Unlock()
			return
		}
		t.expiresAt = time.Time{}
		tick := xtime.Tick{
			Done: make(chan struct{}),
			Time: t.simulator.now,
		}
		t.simulator.cond.L.Unlock()
		t.c <- tick
		<-tick.Done
	})

	t.simulator.cond.L.Unlock()
	t.simulator.cond.Signal()
	return wasReset
}
