// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"github.com/pion/transport/v3/test"
	"github.com/pion/transport/v3/xtime"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var _ xtime.Ticker = &ticker{}

func TestSimulator_NewTicker(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)
	ticker := s.NewTicker(time.Second)

	s.Start()
	defer s.Stop()

	for i := 0; i < 10; i++ {
		tick := <-ticker.C()
		assert.Equal(t, startTime.Add(time.Second*time.Duration(i+1)), tick.Time)
		tick.Done <- struct{}{}
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestSimulator_SynchronizedTickers(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)
	ticker1 := s.NewTicker(time.Second)
	ticker2 := s.NewTicker(time.Second)

	from1 := 0
	from2 := 0

	s.Start()
	defer s.Stop()

	for i := 0; i < 10; i++ {
		select {
		case tick := <-ticker1.C():
			from1++
			assert.Equal(t, startTime.Add(time.Second*time.Duration(from1)), tick.Time)
			assert.LessOrEqual(t, absInt(from1-from2), 1)
			tick.Done <- struct{}{}
		case tick := <-ticker2.C():
			from2++
			assert.Equal(t, startTime.Add(time.Second*time.Duration(from2)), tick.Time)
			assert.LessOrEqual(t, absInt(from1-from2), 1)
			tick.Done <- struct{}{}
		}
	}
}

func TestSimulator_UnsynchronizedTickers(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)

	ticker1 := s.NewTicker(3 * time.Second)
	ticker2 := s.NewTicker(5 * time.Second)

	expectedTicks := []struct {
		now      time.Time
		tickerNo int
	}{
		{
			now:      startTime.Add(3 * time.Second),
			tickerNo: 1,
		},
		{
			now:      startTime.Add(5 * time.Second),
			tickerNo: 2,
		},
		{
			now:      startTime.Add(6 * time.Second),
			tickerNo: 1,
		},
		{
			now:      startTime.Add(9 * time.Second),
			tickerNo: 1,
		},
		{
			now:      startTime.Add(10 * time.Second),
			tickerNo: 2,
		},
	}

	s.Start()
	defer s.Stop()

	for _, expectedTick := range expectedTicks {
		select {
		case tick := <-ticker1.C():
			assert.Equal(t, expectedTick.now, tick.Time)
			assert.Equal(t, 1, expectedTick.tickerNo)
			tick.Done <- struct{}{}
		case tick := <-ticker2.C():
			assert.Equal(t, expectedTick.now, tick.Time)
			assert.Equal(t, 2, expectedTick.tickerNo)
			tick.Done <- struct{}{}
		}
	}
}
