// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vnet

import (
	"context"
	"fmt"
	"github.com/pion/transport/v3/xtime"
	"time"
)

// DelayFilter delays outgoing packets by the given delay. Run must be called
// before any packets will be forwarded.
type DelayFilter struct {
	NIC
	delay       time.Duration
	push        chan struct{}
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
		push:        make(chan struct{}),
		queue:       newChunkQueue(0, 0),
		timeManager: xtime.StdTimeManager{},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f, nil
}

func (f *DelayFilter) onInboundChunk(c Chunk) {
	dl := f.timeManager.Now().Add(f.delay)
	f.queue.push(timedChunk{
		Chunk:    c,
		deadline: f.timeManager.Now().Add(f.delay),
	})
	fmt.Println("deadline ", dl)
	f.push <- struct{}{}
}

// Run starts forwarding of packets. Packets will be forwarded if they spent
// >delay time in the internal queue. Must be called before any packet will be
// forwarded.
func (f *DelayFilter) Run(ctx context.Context) {
	timer := f.timeManager.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-f.push:
			fmt.Println("read pushed")
			next := f.queue.peek().(timedChunk) //nolint:forcetypeassert
			timer.Stop()
			timer.Reset(time.Until(next.deadline))
			fmt.Println("reset at ", next.deadline)
		case tick := <-timer.C():
			fmt.Println("ticked")
			f.onTick(timer, tick.Time)
			fmt.Println("tick done")
			tick.Done <- struct{}{}
		}
	}
}

func (f *DelayFilter) onTick(timer xtime.Timer, now time.Time) {
	next := f.queue.peek()
	fmt.Println("next ", next)
	if next == nil {
		timer.Reset(time.Minute)
		return
	}
	if n, ok := next.(timedChunk); ok && n.deadline.Before(now) {
		f.queue.pop() // ignore result because we already got and casted it from peek
		f.NIC.onInboundChunk(n.Chunk)
	}
	next = f.queue.peek()
	if next == nil {
		fmt.Println("!!!MINUTE RESET!!!")
		timer.Reset(time.Minute)
		return
	}
	if n, ok := next.(timedChunk); ok {
		timer.Reset(time.Until(n.deadline))
	}
}
