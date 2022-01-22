/*
`Monitor` is responsible for high traffic and recovery alerts. It is initialised with
a duration and average request per second value. A FIFO queue is used of size duration where
each entry holds the number of hits for a second of time. As time ticks forward the number
of hits for this second are appended to the end. Once the queue reaches capacity, subsequent
appends cause the front entry to be popped. This allows the total number of hits for the
chosen duration to be efficiently maintained.
*/
package main

import (
	"container/list"
	"fmt"
	"os"
)

type AlertState int

const (
	AlertNone AlertState = iota
	AlertTraffic
)

type Monitor struct {
	queue     *list.List
	capacity  int
	tickHits  int
	totalHits int
	threshold int
	alert     AlertState
	tick      int64
}

// NewMonitor returns a new instance of the Monitor.
func NewMonitor(rps int, window int) *Monitor {
	return &Monitor{
		queue:     list.New(),
		capacity:  window,
		threshold: rps * window,
		alert:     AlertNone,
	}
}

// Sync synchronises the monitor's internal tick.
func (m *Monitor) Sync(t int64) {
	m.tick = t
}

// Hit registers a new hit in the current second.
func (m *Monitor) Hit() {
	m.tickHits++
}

// Tick moves the monitor's internal tick forward. Hits during the last tick are added
// to the queue. If the queue has reached capacity then the front element is removed.
func (m *Monitor) Tick(t int64) {
	if m.queue.Len() == m.capacity {
		front := m.queue.Front()
		m.totalHits -= front.Value.(int)
		m.queue.Remove(front)
	}

	m.tick = t
	m.queue.PushBack(m.tickHits)
	m.totalHits += m.tickHits
	m.tickHits = 0
	m.checkAlerts()
}

// checkAlerts tests whether a new alert should be sent.
func (m *Monitor) checkAlerts() {
	if m.totalHits >= m.threshold && m.alert != AlertTraffic {
		m.alert = AlertTraffic
	} else if m.totalHits < m.threshold && m.alert == AlertTraffic {
		m.alert = AlertNone
	} else {
		return
	}
	sendAlert(m.alert, m.totalHits, m.tick)
}

// sendAlert sends a new alert message based on the monitor's current alert state.
var sendAlert = func(alert AlertState, hits int, alertTime int64) {
	switch alert {
	case AlertTraffic:
		fmt.Print(ColourRed)
		fmt.Printf("[ALERT]\t%v\tHigh traffic generated an alert - hits = %v\n", alertTime, hits)
	case AlertNone:
		fmt.Print(ColourGreen)
		fmt.Printf("[ALERT]\t%v\tAlert recovered\n", alertTime)
	default:
		fmt.Fprintf(os.Stderr, "sendAlert unknown alert state: %v hits: %v time: %v \n", alert, hits, alertTime)
	}
	fmt.Print(ColourReset)
}
