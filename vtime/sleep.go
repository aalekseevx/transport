package vtime

import "time"

func (s *Simulator) Sleep(duration time.Duration) time.Time {
	s.cond.L.Lock()
	ready := make(chan time.Time)
	s.pushEvent(s.now.Add(duration), func() {
		ready <- s.now
	})
	s.cond.L.Unlock()
	s.cond.Signal()
	return <-ready
}
