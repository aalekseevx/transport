// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vnet

import (
	"context"
	"github.com/pion/transport/v3/xtime"
	"time"
)

type push struct {
	chunk Chunk
	done  chan struct{}
}

// DelayFilter delays outgoing packets by the given delay. Run must be called
// before any packets will be forwarded.
type DelayFilter struct {
	NIC
	delay       time.Duration
	push        chan push
	queue       *chunkQueue
	timeManager xtime.TimeManager
}

type timedChunk struct {
	Chunk
	deadline time.Time
}

type DelayFilterOption func(*DelayFilter)

func DelayFilterWithTimeManager(manager xtime.TimeManager) DelayFilterOption {
	return func(f *DelayFilter) {
		f.timeManager = manager
	}
}

// NewDelayFilter creates a new DelayFilter with the given nic and delay.
func NewDelayFilter(nic NIC, delay time.Duration, opts ...DelayFilterOption) (*DelayFilter, error) {
	f := &DelayFilter{
		NIC:         nic,
		delay:       delay,
		push:        make(chan push),
		queue:       newChunkQueue(0, 0),
		timeManager: xtime.StdTimeManager{},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f, nil
}

func (f *DelayFilter) onInboundChunk(c Chunk) {
	ch := make(chan struct{})
	f.push <- push{
		chunk: c,
		done:  ch,
	}
	<-ch
}

// Run starts forwarding of packets. Packets will be forwarded if they spent
// >delay time in the internal queue. Must be called before any packet will be
// forwarded.
func (f *DelayFilter) Run(ctx context.Context) {
	timer := f.timeManager.NewTimer(0, true)
	for {
		select {
		case <-ctx.Done():
			return
		case chunk := <-f.push:
			nowTick := f.timeManager.FreezeNow()
			f.queue.push(timedChunk{
				Chunk:    chunk.chunk,
				deadline: nowTick.Time.Add(f.delay),
			})
			next := f.queue.peek().(timedChunk) //nolint:forcetypeassert
			timer.Stop()
			udl := f.timeManager.Until(next.deadline)
			timer.Reset(udl)
			nowTick.Done <- struct{}{}
			chunk.done <- struct{}{}
		case tick := <-timer.C():
			f.onTick(timer, tick.Time)
			tick.Done <- struct{}{}
		}
	}
}

func (f *DelayFilter) onTick(timer xtime.Timer, now time.Time) {
	next := f.queue.peek()
	if next == nil {
		timer.Reset(time.Minute)
		return
	}
	if n, ok := next.(timedChunk); ok {
		if n.deadline.Before(now) || n.deadline.Equal(now) {
			f.queue.pop() // ignore result because we already got and casted it from peek
			f.NIC.onInboundChunk(n.Chunk)
		}
	}
	next = f.queue.peek()
	if next == nil {
		timer.Reset(time.Minute)
		return
	}
	if n, ok := next.(timedChunk); ok {
		timer.Reset(f.timeManager.Until(n.deadline))
	}
}
