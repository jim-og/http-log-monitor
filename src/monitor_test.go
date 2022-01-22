package main

import (
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	currTime := time.Now().Unix()
	monitor := NewMonitor(1, 3)

	var wants = []struct {
		alert     AlertState
		hits      int
		alertTime int64
	}{
		{AlertTraffic, 3, currTime + 1},
		{AlertNone, 1, currTime + 4},
	}

	// Save and restore original sendAlert
	savedSendAlert := sendAlert
	defer func() {
		sendAlert = savedSendAlert
	}()

	// Install the test's sendAlert
	index := 0
	sendAlert = func(alert AlertState, hits int, alertTime int64) {
		if index >= len(wants) {
			t.Errorf(`Alert unexpected: sendAlert(%v, %v, %v)`, alert, hits, alertTime)
			return
		}
		want := wants[index]
		if alert != want.alert || hits != want.hits || alertTime != want.alertTime {
			t.Errorf(`Alert mismatch: (alert=%v, hits=%v, time=%v), want (alert=%v, hits=%v, time=%v)`,
				alert, hits, alertTime, want.alert, want.hits, want.alertTime)
		}
		index++
	}

	monitor.Hit()
	monitor.Hit()
	monitor.Hit()
	currTime++
	monitor.Tick(currTime)
	monitor.Hit()
	currTime++
	monitor.Tick(currTime)
	currTime++
	monitor.Tick(currTime)
	currTime++
	monitor.Tick(currTime)
}
