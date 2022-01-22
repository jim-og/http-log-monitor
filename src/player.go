/*
`Player` is the controller which facilitates communication between the `Reader`, `Stats`
and `Monitor` components. It is responsible for simulating the passage of time. A requirement
is that any alerts must be accurate to within a second, therefore a second represents a single
unit of time. As the log file is read, access hits for the current second are registered with
the `Stats` and `Monitor`. When a timestamp in the log is greater than the current second the
`Player` “ticks” time forward for each second until it is synchronised with the latest timestamp.
*/
package main

import (
	"fmt"
	"os"
)

const (
	defaultFilePath = "../input/sample_csv.txt"
)

type Player struct {
	reader  *Reader
	stats   *Stats
	monitor *Monitor
}

// NewPlayer returns a new instance of the Player.
func NewPlayer(filePath string, statsInterval int64, monitorRps int, monitorWindow int) *Player {
	return &Player{
		reader:  NewReader(filePath),
		stats:   NewStats(statsInterval),
		monitor: NewMonitor(monitorRps, monitorWindow),
	}
}

// Play starts playback of a log file.
func (p *Player) Play() {
	src := make(chan LogModel)
	tick := int64(0)

	go p.reader.Process(src)
	for line := range src {
		if tick == 0 {
			// First log line, sync monitor and stats.
			p.monitor.Sync(line.date)
			p.stats.Sync(line.date)
			tick = line.date + 1
		}

		// Reject access requests which occured before this second.
		if line.date < tick-1 {
			fmt.Fprintf(os.Stderr, "Access request not processed. Request time %v, current tick started at %v. %v", line.date, tick-1, line)
			continue
		}

		// Bring time forward until it is synchronised with the latest timestamp
		for ; tick <= line.date; tick = tick + 1 {
			p.monitor.Tick(tick)
			p.stats.Tick(tick)
		}

		// Register a hit
		p.monitor.Hit()
		p.stats.Hit(line.section)
	}
}
