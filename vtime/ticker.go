// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

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
	if duration < 0 {
		panic("ticker duration must not be negative")
	}
	t := &ticker{
		c:       make(chan xtime.Tick),
		stopped: false,
	}
	var do func()
	do = func() {
		if t.stopped {
			return
		}
		tick := tick{
			done: make(chan struct{}),
			time: s.now,
		}
		t.c <- tick
		<-tick.done
		s.queue.Push(s.now.Add(duration), do)
	}

	s.timeLock.RLock()
	s.queue.Push(s.now.Add(duration), do)
	s.timeLock.RUnlock()
	return t
}

func (t *ticker) C() <-chan xtime.Tick {
	return t.c
}

func (t *ticker) Stop() {
	t.stopped = true
}
