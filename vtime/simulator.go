package vtime

import (
	"container/heap"
	"sync"
	"time"
)

type Simulator struct {
	nextID  int
	now     time.Time
	queue   *eventQueue
	stopped bool
	cond    sync.Cond
}

func NewSimulator(start time.Time) *Simulator {
	eq := &eventQueue{}
	heap.Init(eq)
	return &Simulator{
		nextID: 0,
		now:    start,
		queue:  eq,
		cond:   sync.Cond{L: &sync.Mutex{}},
	}
}

func (s *Simulator) Start() {
	go s.eventLoop()
}

func (s *Simulator) Stop() {
	s.cond.L.Lock()
	s.stopped = true
	s.cond.L.Unlock()
	s.cond.Signal()
}

func (s *Simulator) Now() time.Time {
	s.cond.L.Lock()
	now := s.now
	s.cond.L.Unlock()
	return now
}

func (s *Simulator) eventLoop() {
	for {
		s.cond.L.Lock()
		for s.queue.Len() == 0 && !s.stopped {
			s.cond.Wait() // Wait until a new event is added
		}
		if s.stopped {
			return
		}
		event := heap.Pop(s.queue).(queueElement)
		s.now = event.at
		s.cond.L.Unlock()
		event.do()
	}
}

func (s *Simulator) pushEvent(at time.Time, do func()) {
	event := queueElement{
		id: s.nextID,
		at: at,
		do: do,
	}
	s.nextID += 1
	heap.Push(s.queue, event)
}
