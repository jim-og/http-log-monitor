package main

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	var wants = []LogModel{
		{date: 1549573860},
		{date: 1549573859},
		{date: 1549573860},
		{date: 1549573862},
		{date: 1549573861},
		{date: 1549573858},
		{date: 1549570},
		{date: 154957650},
		{date: 15490},
		{date: 154945675670},
	}

	pq := make(PriorityQueue, 0)
	for _, want := range wants {
		heap.Push(&pq, &LogItem{value: want, priority: want.date})
	}

	prevTime := int64(0)
	for pq.Len() > 0 {
		got := heap.Pop(&pq).(*LogItem)
		if got.priority < prevTime {
			t.Errorf(`Queue out of order. Previous time %v, got %v`, prevTime, got.priority)
		}
		prevTime = got.priority
	}
}
