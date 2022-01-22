package main

import (
	"testing"
)

type StatTest struct {
	input string
	want  int
}

func makeTests() []StatTest {
	return []StatTest{
		{"/api", 8},
		{"/report", 7},
		{"/admin", 6},
		{"/user", 5},
		{"/login", 2},
		{"/logout", 4},
	}
}

func TestStatsHits(t *testing.T) {
	stats := NewStats(int64(10))
	tests := makeTests()

	// Populate hits
	for _, test := range tests {
		for i := 0; i < test.want; i++ {
			stats.Hit(test.input)
		}
	}

	// Check hits
	for _, test := range tests {
		if got := stats.Hits(test.input); got != test.want {
			t.Errorf(`stats.Hits(%q) returned %v, want %v`, test.input, got, test.want)
		}
	}
}

func TestStatsTopK(t *testing.T) {
	stats := NewStats(int64(10))
	tests := makeTests()

	// Populate hits
	for _, test := range tests {
		for i := 0; i < test.want; i++ {
			stats.Hit(test.input)
		}
	}

	for _, k := range []int{2, 3} {
		for index, got := range stats.TopK(k) {
			if index >= len(tests) {
				t.Errorf(`stats.TopK(%v) returned %v results, want less than %v`, k, index+1, len(tests))
				break
			}
			test := tests[index]
			if got.section != test.input || got.hits != test.want {
				t.Errorf(`stats.TopK(%v) returned rank %v as {%q: %v}, want {%q: %v}`, k, index+1, got.section, got.hits, test.input, test.want)
			}
		}
	}
}

func TestStatsClear(t *testing.T) {
	stats := NewStats(int64(10))
	tests := makeTests()

	// Populate hits.
	for _, test := range tests {
		for i := 0; i < test.want; i++ {
			stats.Hit(test.input)
		}
	}

	// Confirm hits registered.
	for _, test := range tests {
		if got := stats.Hits(test.input); got != test.want {
			t.Errorf(`stats.Hits(%q) returned %v, want %v`, test.input, got, test.want)
		}
	}

	// Confirm all sections tracked in top k.
	if got := stats.TopK(len(tests)); len(got) != len(tests) {
		t.Errorf(`stats.TopK(%v) returned %v results, want %v`, len(tests), len(got), len(tests))
	}

	stats.Clear()

	// Confirm all section hits are 0
	for _, test := range tests {
		if got := stats.Hits(test.input); got != 0 {
			t.Errorf(`stats.Clear() then stats.Hits(%q) returned %v, want %v`, test.input, got, 0)
		}
	}

	// Confirm no sections in top k
	if got := stats.TopK(len(tests)); len(got) != 0 {
		t.Errorf(`stats.Clear() then stats.TopK(%v) returned %v results, want %v`, len(tests), len(got), 0)
	}
}
