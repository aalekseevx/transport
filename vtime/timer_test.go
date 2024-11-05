package vtime

import (
	"github.com/pion/transport/v3/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimulator_NewTimer(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)
	timer := s.NewTimer(time.Second)

	s.Start()
	defer s.Stop()

	tick := <-timer.C()
	assert.Equal(t, time.Time{}.Add(time.Second), tick.Time)
	tick.Done <- struct{}{}
}

func TestTimer_Reset(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)
	timer1 := s.NewTimer(time.Second)
	timer2 := s.NewTimer(2 * time.Second)

	s.Start()
	defer s.Stop()

	tick1 := <-timer1.C()
	assert.Equal(t, time.Time{}.Add(time.Second), tick1.Time)
	timer2.Reset(4 * time.Second)
	tick1.Done <- struct{}{}

	tick2 := <-timer2.C()
	assert.Equal(t, time.Time{}.Add(5*time.Second), tick2.Time)
	timer1.Reset(10 * time.Second)
	tick2.Done <- struct{}{}

	tick3 := <-timer1.C()
	assert.Equal(t, time.Time{}.Add(15*time.Second), tick3.Time)
	tick3.Done <- struct{}{}
}
