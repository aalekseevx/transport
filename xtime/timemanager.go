// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package xtime

import "time"

type Tick interface {
	// Time returns the time of the Tick
	Time() time.Time
	// Done must be called when you are done with the Tick and we can move forward
	Done()
}

type Ticker interface {
	// C returns channel, which will emit Ticks
	// When job is done, call Done()
	C() <-chan Tick
	// Stop stops the ticker
	Stop()
}

type Timer interface {
	// C returns channel, which will emit a single Tick
	C() <-chan Tick
	// Stop stops the timer
	Stop() bool
	// Reset resets the timer
	Reset(time.Duration) bool
}

type TimeManager interface {
	NewTicker(d time.Duration) Ticker
	NewTimer(d time.Duration, blocking bool) Timer
	After(d time.Duration, blocking bool) <-chan Tick
	Sleep(d time.Duration) time.Time
	Now() time.Time
	FreezeNow() Tick
	Until(t time.Time) time.Duration
	Since(t time.Time) time.Duration
}
