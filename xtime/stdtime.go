// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package xtime

import "time"

type StdTimeManager struct{}

func (r StdTimeManager) NewTicker(duration time.Duration) Ticker {
	ticker := time.NewTicker(duration)
	c := make(chan Tick)
	go func() {
		for now := range ticker.C {
			c <- Tick{
				Done: make(chan struct{}, 1),
				Time: now,
			}
		}
	}()
	return &stdTimeTicker{
		ticker: time.NewTicker(duration),
		c:      c,
	}
}

func (t StdTimeManager) NewTimer(d time.Duration) Timer {
	c := make(chan Tick)
	timer := time.NewTimer(d)
	go func() {
		for now := range timer.C {
			c <- Tick{
				Done: make(chan struct{}, 1),
				Time: now,
			}
		}
	}()
	return stdTimer{
		timer: timer,
		c:     c,
	}
}

func (t StdTimeManager) After(d time.Duration) <-chan Tick {
	return t.NewTimer(d).C()
}

func (r StdTimeManager) Sleep(duration time.Duration) time.Time {
	time.Sleep(duration)
	return time.Now()
}

func (r StdTimeManager) Now() time.Time {
	return time.Now()
}

func (t StdTimeManager) FreezeNow() Tick {
	return Tick{
		Done: make(chan struct{}, 1),
		Time: time.Now(),
	}
}

func (t StdTimeManager) Until(tm time.Time) time.Duration {
	return time.Until(tm)
}

type stdTimeTicker struct {
	ticker *time.Ticker
	c      <-chan Tick
}

func (t stdTimeTicker) C() <-chan Tick {
	return t.c
}

type stdTimer struct {
	timer *time.Timer
	c     chan Tick
}

func (t stdTimer) Reset(duration time.Duration) bool {
	return t.timer.Reset(duration)
}

func (t stdTimer) C() <-chan Tick {
	return t.c
}

func (t stdTimer) Stop() bool {
	stopped := t.timer.Stop()
	if !stopped {
		<-t.timer.C // remove on go1.23
	}
	return stopped
}
