package vtime

import "time"

type queueElement struct {
	id int
	at time.Time
	do func()
}

type eventQueue []queueElement

func (eq eventQueue) Len() int { return len(eq) }

func (eq eventQueue) Less(i, j int) bool {
	return eq[i].at.Before(eq[j].at)
}

func (eq eventQueue) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
}

func (eq *eventQueue) Push(x interface{}) {
	event := x.(queueElement)
	*eq = append(*eq, event)
}

func (eq *eventQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	event := old[n-1]
	*eq = old[0 : n-1]
	return event
}
