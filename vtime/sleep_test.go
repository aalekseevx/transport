// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"github.com/pion/transport/v3/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimulator_Sleep(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()

	startTime := time.Time{}
	s := NewSimulator(startTime)
	s.Start()
	defer s.Stop()

	assert.Equal(t, startTime.Add(time.Second), s.Sleep(time.Second))
	assert.Equal(t, startTime.Add(3*time.Second), s.Sleep(2*time.Second))
	assert.Equal(t, startTime.Add(6*time.Second), s.Sleep(3*time.Second))
}
