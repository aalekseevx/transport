// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package xtime

import "time"

type tick struct {
	time time.Time
}

func (t tick) Time() time.Time {
	return t.time
}

func (t tick) Done() {
}

type StdTimeManager struct{}

func (r StdTimeManager) NewTicker(duration time.Duration) Ticker {
	ticker := time.NewTicker(duration)
	c := make(chan Tick)
	go func() {
		for now := range ticker.C {
			c <- tick{
				time: now,
			}
		}
	}()
	return &stdTimeTicker{
		ticker: time.NewTicker(duration),
		c:      c,
	}
}

func (t StdTimeManager) NewTimer(d time.Duration, _ bool) Timer {
	c := make(chan Tick)
	timer := time.NewTimer(d)
	go func() {
		for now := range timer.C {
			c <- tick{
				time: now,
			}
		}
	}()
	return stdTimer{
		timer: timer,
		c:     c,
	}
}

func (t StdTimeManager) After(d time.Duration, _ bool) <-chan Tick {
	return t.NewTimer(d, false).C()
}

func (r StdTimeManager) Sleep(duration time.Duration) time.Time {
	time.Sleep(duration)
	return time.Now()
}

func (r StdTimeManager) Now() time.Time {
	return time.Now()
}

func (r StdTimeManager) Since(tm time.Time) time.Duration {
	return time.Since(tm)
}

func (t StdTimeManager) FreezeNow() Tick {
	return tick{
		time: time.Now(),
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

func (t stdTimeTicker) Stop() {
	t.ticker.Stop()
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
		select {
		case <-t.timer.C: // remove on go1.23
		default:

		}
	}
	return stopped
}
