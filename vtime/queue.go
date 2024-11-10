// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vtime

import (
	"container/heap"
	"sync"
	"time"
)

type queueElement struct {
	id int
	at time.Time
	do func()
}

type eventHeap []queueElement

func (eq eventHeap) Len() int { return len(eq) }

func (eq eventHeap) Less(i, j int) bool {
	return eq[i].at.Before(eq[j].at)
}

func (eq eventHeap) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
}

func (eq *eventHeap) Push(x interface{}) {
	event := x.(queueElement)
	*eq = append(*eq, event)
}

func (eq *eventHeap) Pop() interface{} {
	old := *eq
	n := len(old)
	event := old[n-1]
	*eq = old[0 : n-1]
	return event
}

type eventQueue struct {
	eventHeap *eventHeap
	cond      sync.Cond
	nextID    int
	stopped   bool
}

func newEventQueue() *eventQueue {
	eh := &eventHeap{}
	heap.Init(eh)
	return &eventQueue{
		eventHeap: eh,
		cond:      sync.Cond{L: &sync.Mutex{}},
	}
}

func (q *eventQueue) Len() int {
	return q.eventHeap.Len()
}

func (q *eventQueue) Push(at time.Time, do func()) {
	q.cond.L.Lock()

	event := queueElement{
		id: q.nextID,
		at: at,
		do: do,
	}
	q.nextID += 1
	heap.Push(q.eventHeap, event)
	q.cond.L.Unlock()
	q.cond.Signal()
}

func (q *eventQueue) Pull() (time.Time, func(), bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.eventHeap.Len() == 0 && !q.stopped {
		q.cond.Wait() // Wait until a new event is added
	}
	if q.stopped {
		return time.Time{}, nil, false
	}
	event := heap.Pop(q.eventHeap).(queueElement)
	return event.at, event.do, true
}

func (q *eventQueue) Stop() {
	q.cond.L.Lock()
	q.stopped = true
	q.cond.L.Unlock()
	q.cond.Broadcast()
}
