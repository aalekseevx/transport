// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"github.com/pion/transport/v3/xtime"
	"sync"
	"time"
)

type timer struct {
	c         chan xtime.Tick
	mu        sync.Mutex
	expiresAt time.Time
	simulator *Simulator
}

func (s *Simulator) NewTimer(d time.Duration) xtime.Timer {
	t := &timer{
		c:         make(chan xtime.Tick),
		mu:        sync.Mutex{},
		expiresAt: time.Time{},
		simulator: s,
	}
	t.Reset(d)
	return t
}

func (s *Simulator) After(d time.Duration) <-chan xtime.Tick {
	return s.NewTimer(d).C()
}

func (t *timer) C() <-chan xtime.Tick {
	return t.c
}

func (t *timer) Stop() bool {
	t.mu.Lock()
	t.expiresAt = time.Time{}
	wasSet := !t.expiresAt.IsZero()
	t.mu.Unlock()
	return wasSet
}

func (t *timer) Reset(duration time.Duration) bool {
	if duration < 0 {
		panic("duration must be non negative")
	}
	t.simulator.timeLock.RLock()
	newExpiresAt := t.simulator.now.Add(duration)
	wasReset := t.expiresAt.Before(newExpiresAt)
	t.expiresAt = t.simulator.now.Add(duration)
	t.simulator.timeLock.RUnlock()

	t.simulator.queue.Push(t.expiresAt, func() {
		t.mu.Lock()
		if !t.expiresAt.Equal(t.simulator.now) {
			t.mu.Unlock()
			return
		}
		t.expiresAt = time.Time{}
		t.mu.Unlock()

		tick := xtime.Tick{
			Done: make(chan struct{}),
			Time: t.simulator.now,
		}
		select {
		case t.c <- tick:
			// Somebody is listening, wait for reply
			<-tick.Done
		default:
			// Nobody is actively listening
		}
	})

	return wasReset
}
