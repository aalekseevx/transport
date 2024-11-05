package xtime

import "time"

type StdTimeManager struct{}

func (r StdTimeManager) NewTicker(duration time.Duration) Ticker {
	return &stdTimeTicker{
		ticker: time.NewTicker(duration),
	}
}

func (t StdTimeManager) NewTimer(d time.Duration) Timer {
	return stdTimer{
		timer: time.NewTimer(d),
	}
}

func (t StdTimeManager) After(d time.Duration) <-chan Tick {
	c := make(chan Tick)
	go func() {
		c <- Tick{
			Done: make(chan struct{}, 1),
			Time: <-time.After(d),
		}
	}()
	return c
}

func (r StdTimeManager) Sleep(duration time.Duration) time.Time {
	time.Sleep(duration)
	return time.Now()
}

func (t StdTimeManager) Now() time.Time {
	return time.Now()
}

type stdTimeTicker struct {
	ticker *time.Ticker
}

func (t stdTimeTicker) C() <-chan Tick {
	ticked := make(chan Tick)
	go func() {
		for now := range t.ticker.C {
			ticked <- Tick{
				Done: make(chan struct{}, 1),
				Time: now,
			}
		}
	}()
	return ticked
}

type stdTimer struct {
	timer *time.Timer
}

func (t stdTimer) Reset(duration time.Duration) bool {
	return t.timer.Reset(duration)
}

func (t stdTimer) C() <-chan Tick {
	c := make(chan Tick)
	go func() {
		c <- Tick{
			Done: make(chan struct{}, 1),
			Time: <-t.timer.C,
		}
	}()
	return c
}

func (t stdTimer) Stop() bool {
	stopped := t.timer.Stop()
	if !stopped {
		<-t.timer.C // remove on go1.23
	}
	return stopped
}
