// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"fmt"
	"time"
)

func (s *Simulator) Sleep(duration time.Duration) time.Time {
	ready := make(chan time.Time)
	s.timeLock.RLock()
	fmt.Println("Simulator.Sleep", s.now)
	fmt.Println("Simulator.Sleep until", s.now.Add(duration))
	s.queue.Push(s.now.Add(duration), func() {
		ready <- s.now
	})
	s.timeLock.RUnlock()
	return <-ready
}
