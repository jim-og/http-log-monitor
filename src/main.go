/*
http-log-monitor is a Go HTTP log monitoring console programme. It is able to read CSV-encoded
HTTP access logs, display the most popular endpoint sections over a specified time period, and
alert when traffic passes a certain threshold across a given time period.
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

var filePath = flag.String("input", "", "input log file path (required)")
var statsInterval = flag.Int("stats", 10, "time interval between displaying stats in seconds")
var monitorWindow = flag.Int("alert", 120, "duration of the high traffic alert window in seconds")
var monitorRps = flag.Int("rps", 10, "average requests per second threshold for high traffic alert")

func main() {
	flag.Parse()
	if len(*filePath) == 0 {
		fmt.Fprint(os.Stderr, "No input file path provided. Use -input to specify one.\n")
		return
	}
	player := NewPlayer(*filePath, int64(*statsInterval), *monitorRps, *monitorWindow)
	player.Play()
}
