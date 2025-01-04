// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"github.com/pion/transport/v3/xtime"
	"sync/atomic"
	"time"
)

type Simulator struct {
	nextID      int
	now         time.Time
	queue       *eventQueue
	timeLock    recursiveRWLock
	timeStopped atomic.Bool
}

func NewSimulator(start time.Time) *Simulator {
	return &Simulator{
		nextID:      0,
		now:         start,
		queue:       newEventQueue(),
		timeLock:    recursiveRWLock{},
		timeStopped: atomic.Bool{},
	}
}

func (s *Simulator) Start() {
	go s.eventLoop()
}

func (s *Simulator) Stop() {
	s.queue.Stop()
}

func (s *Simulator) Now() time.Time {
	return s.now
}

func (s *Simulator) Since(tm time.Time) time.Duration {
	return s.now.Sub(tm)
}

func (s *Simulator) FreezeNow() xtime.Tick {
	s.timeLock.RLock()
	now := s.now
	ch := make(chan struct{})
	go func() {
		<-ch
		s.timeLock.RUnlock()
	}()
	return tick{
		done: ch,
		time: now,
	}
}

func (s *Simulator) Until(tm time.Time) time.Duration {
	return tm.Sub(s.now)
}

func (s *Simulator) eventLoop() {
	for {
		now, do, run := s.queue.Pull()
		if !run {
			return
		}
		s.timeLock.Lock()
		s.now = now
		s.timeLock.Unlock()

		s.timeStopped.Store(true)
		do()
		s.timeStopped.Store(false)
	}
}
