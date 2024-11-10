// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vnet

import (
	"context"
	"github.com/pion/transport/v3/vtime"
	"github.com/pion/transport/v3/xtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testSchedulesOnePacketAtATime(t *testing.T, tm xtime.TimeManager) {
	nic := newMockNIC(t)
	df, err := NewDelayFilter(nic, 10*time.Millisecond, DelayFilterWithTimeManager(tm))
	if !assert.NoError(t, err, "should succeed") {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go df.Run(ctx)

	type TimestampedChunk struct {
		ts time.Time
		c  Chunk
	}
	receiveCh := make(chan TimestampedChunk)
	nic.mockOnInboundChunk = func(c Chunk) {
		receivedAt := tm.Now()
		receiveCh <- TimestampedChunk{
			ts: receivedAt,
			c:  c,
		}
	}

	lastNr := -1
	for i := 0; i < 100; i++ {
		sent := tm.FreezeNow()
		df.onInboundChunk(&chunkUDP{
			chunkIP:  chunkIP{timestamp: sent.Time},
			userData: []byte{byte(i)},
		})
		sent.Done <- struct{}{}

		c := <-receiveCh
		nr := int(c.c.UserData()[0])

		assert.Greater(t, nr, lastNr)
		lastNr = nr

		assert.GreaterOrEqual(t, c.ts.Sub(sent.Time), 10*time.Millisecond)
	}
}

func testSchedulesSubsequentManyPackets(t *testing.T, tm xtime.TimeManager) {
	nic := newMockNIC(t)
	df, err := NewDelayFilter(nic, 10*time.Millisecond, DelayFilterWithTimeManager(tm))
	if !assert.NoError(t, err, "should succeed") {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go df.Run(ctx)

	type TimestampedChunk struct {
		ts time.Time
		c  Chunk
	}
	receiveCh := make(chan TimestampedChunk)
	nic.mockOnInboundChunk = func(c Chunk) {
		receivedAt := time.Now()
		receiveCh <- TimestampedChunk{
			ts: receivedAt,
			c:  c,
		}
	}

	// schedule 100 chunks
	sent := tm.FreezeNow()
	for i := 0; i < 100; i++ {
		df.onInboundChunk(&chunkUDP{
			chunkIP:  chunkIP{timestamp: sent.Time},
			userData: []byte{byte(i)},
		})
	}
	sent.Done <- struct{}{}

	// receive 100 chunks with delay>10ms
	for i := 0; i < 100; i++ {
		select {
		case c := <-receiveCh:
			nr := int(c.c.UserData()[0])
			assert.Equal(t, i, nr)
			assert.Greater(t, c.ts.Sub(sent.Time), 10*time.Millisecond)
		case <-tm.After(time.Second, false):
			assert.Fail(t, "expected to receive next chunk")
		}
	}
}

func TestDelayFilter(t *testing.T) {
	t.Run("schedulesOnePacketAtATime_StdTime", func(t *testing.T) {
		testSchedulesOnePacketAtATime(t, xtime.StdTimeManager{})
	})

	t.Run("schedulesOnePacketAtATime_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		tm.Start()
		defer tm.Stop()

		testSchedulesOnePacketAtATime(t, tm)
	})

	t.Run("schedulesSubsequentManyPackets_StdTime", func(t *testing.T) {
		testSchedulesSubsequentManyPackets(t, xtime.StdTimeManager{})
	})

	t.Run("schedulesSubsequentManyPackets_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		tm.Start()
		defer tm.Stop()

		testSchedulesSubsequentManyPackets(t, tm)
	})
}
