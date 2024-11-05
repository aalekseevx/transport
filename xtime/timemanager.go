package xtime

import "time"

type Tick struct {
	Done chan struct{}
	Time time.Time
}

type Ticker interface {
	// C returns channel, which emit chan struct{} on every tick
	// When job is done, send struct{} to the received channel
	C() <-chan Tick
}

type Timer interface {
	C() <-chan Tick
	Stop() bool
	Reset(time.Duration) bool
}

type TimeManager interface {
	NewTicker(d time.Duration) Ticker
	NewTimer(d time.Duration) Timer
	After(d time.Duration) <-chan Tick
	Sleep(d time.Duration) time.Time
	Now() time.Time
}
