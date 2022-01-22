/*
Stats maintains a ranking of the top sections based on the number of hits. Two data
structures are used to efficiently perform this. An unordered map, with section as key
and hits as value, keeps count of each section’s total number of hits. An ordered map,
with hits as key and sections as value, tracks the top sections. Updating the former
in O(1) time allows the latter to be updated in O(log n) time.
*/
package main

import (
	"fmt"

	"github.com/google/btree"
)

const (
	defaultShowTopK = 10
)

type TopKEntry struct {
	hits     int             // number of hits
	sections map[string]bool // sections with this number of hits
}

// Less is a comparator used by Stats.topK to satisfy the btree.Item interface.
func (entry TopKEntry) Less(than btree.Item) bool {
	return entry.hits < than.(TopKEntry).hits
}

type TopKResult struct {
	section string
	hits    int
}

// Stats tracks the number of section hits over a chosen interval.
type Stats struct {
	hits       map[string]int // tracks the number of section hits, uses O(n) space
	topK       *btree.BTree   // maintains the list of sections ordered by number of hits, uses O(n) space
	tick       int64
	tickReport int64
	interval   int64
	showTopK   int
}

// NewStats returns a new Stats object used to track statistics.
func NewStats(interval int64) *Stats {
	return &Stats{
		hits:       make(map[string]int),
		topK:       btree.New(2),
		tick:       0,
		tickReport: 0 + interval,
		interval:   interval,
		showTopK:   defaultShowTopK,
	}
}

// Tick moves the Stats' internal tick forward.
// If the chosen time interval has been reached then statistics are reported.
func (s *Stats) Tick(t int64) {
	s.tick = t
	if s.tickReport <= s.tick {
		sendStats(s.tick, s.TopK(s.showTopK))
		s.Clear()
		s.tickReport += s.interval
	}
}

// Sync synchronises Stats' internal tick.
func (s *Stats) Sync(t int64) {
	s.tick = t
	s.tickReport = t + s.interval
}

// Hit records an additional hit for a given section.
// This call takes:
//   O(1): to update counts.
//   O(log n):  to update the top k.
func (s *Stats) Hit(key string) {
	count, found := s.hits[key]
	s.hits[key]++

	if found {
		// Remove section from its current position in the topK.
		item := s.topK.Get(TopKEntry{count, nil})
		if item != nil {
			delete(item.(TopKEntry).sections, key)
		}
		if len(item.(TopKEntry).sections) == 0 {
			s.topK.Delete(item)
		}
	}

	// Add section to its new position in the topK.
	item := s.topK.Get(TopKEntry{count + 1, nil})
	if item == nil {
		s.topK.ReplaceOrInsert(TopKEntry{count + 1, map[string]bool{key: true}})
	} else {
		item.(TopKEntry).sections[key] = true
	}
}

// sendStats displays the top sections over the chosen interval.
var sendStats = func(tick int64, topK []TopKResult) {
	fmt.Print(ColourYellow)
	fmt.Printf("[STATS]\t%v\t", tick)
	for _, got := range topK {
		fmt.Printf("%s: %v ", got.section, got.hits)
	}
	fmt.Print("\n")
	fmt.Print(ColourReset)
}

// Clear resets all section hit counts.
func (s *Stats) Clear() {
	for k := range s.hits {
		delete(s.hits, k)
	}
	s.topK.Clear(false)
}

// Hits returns the number of hits for a given section.
// This call takes O(1)
func (s *Stats) Hits(key string) int {
	hits, found := s.hits[key]
	if !found {
		return 0
	} else {
		return hits
	}
}

// TopK returns the top k number of sections with the most hits. If multiple sections have
// the same number of hits, they are all returned.
// This call takes O(log n)
func (s *Stats) TopK(k int) []TopKResult {
	var result []TopKResult
	if k == 0 || s.topK.Len() == 0 {
		return result
	}

	it := func(i btree.Item) bool {
		hits := i.(TopKEntry).hits
		for section := range i.(TopKEntry).sections {
			result = append(result, TopKResult{section, hits})
			k--
		}
		return k > 0
	}
	s.topK.Descend(it)

	return result
}
