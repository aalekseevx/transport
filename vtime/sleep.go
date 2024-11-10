// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import "time"

func (s *Simulator) Sleep(duration time.Duration) time.Time {
	ready := make(chan time.Time)
	s.timeLock.RLock()
	s.queue.Push(s.now.Add(duration), func() {
		ready <- s.now
	})
	s.timeLock.RUnlock()
	return <-ready
}
