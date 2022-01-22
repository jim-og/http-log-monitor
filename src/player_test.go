package main

import "testing"

func TestPlay(t *testing.T) {
	// Save and restore original sendStats and sendAlert
	savedSendStats := sendStats
	savedSendAlert := sendAlert
	defer func() {
		sendStats = savedSendStats
		sendAlert = savedSendAlert
	}()

	// Install the test's sendStats
	statInterval := int64(10)
	statWant := int64(1549573869)
	sendStats = func(tick int64, topK []TopKResult) {
		if tick != statWant {
			t.Errorf(`Incorrect Stats update, want %v, got %v`, tick, statWant)
		}
		statWant += statInterval
	}

	var alertWant = []struct {
		timestamp int64
		alert     AlertState
		hits      int
	}{
		{1549573957, AlertTraffic, 1206},
		{1549574044, AlertNone, 0},
		{1549574164, AlertTraffic, 1218},
		{1549574303, AlertNone, 0},
	}

	// Install the test's sendAlert
	index := 0
	sendAlert = func(alert AlertState, hits int, alertTime int64) {
		want := alertWant[index]
		if alert != want.alert || (hits != want.hits && alert != AlertNone) || alertTime != want.timestamp {
			t.Errorf(`Incorrect alert, want alert: %v hits: %v time: %v, got alert: %v hits: %v time: %v`,
				alert, hits, alertTime,
				want.alert, want.hits, want.timestamp)
		}
		index++
	}

	p := NewPlayer(defaultFilePath, statInterval, 10, 120)
	p.Play()
}
