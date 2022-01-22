package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetSection(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"/api/user", "/api"},
		{"/report", "/report"},
		{"/fee/fi/fo/fum", "/fee"},
		{"/", "/"},
		{"", ""},
		{"///", "/"},
	}
	for _, test := range tests {
		if got := getSection(test.input); got != test.want {
			t.Errorf(`getSession(%q) returned %q, want %q`, test.input, got, test.want)
		}
	}
}

func TestParseRequest(t *testing.T) {
	var tests = []struct {
		input string
		want  RequestData
	}{
		{"GET /api/user HTTP/1.0", RequestData{"GET", "/api/user", "/api", "HTTP/1.0"}},
		{"POST /report HTTP/1.0", RequestData{"POST", "/report", "/report", "HTTP/1.0"}},
		{"DELETE /fee/fi/fo/fum HTTP/1.0", RequestData{"DELETE", "/fee/fi/fo/fum", "/fee", "HTTP/1.0"}},
	}
	for _, test := range tests {
		if got := parseRequest(test.input); got != test.want {
			t.Errorf(`parseRequest(%q) returned %q, want %q`, test.input, got, test.want)
		}
	}
}

func TestProcess(t *testing.T) {
	r := NewReader(defaultFilePath)
	out := make(chan LogModel)
	go r.Process(out)
}

func TestParseHeader(t *testing.T) {
	var tests = []struct {
		input []string
		want  map[string]int
	}{
		{[]string{"remotehost", "rfc931", "authuser", "date", "request", "status", "bytes"},
			map[string]int{"remotehost": 0, "rfc931": 1, "authuser": 2, "date": 3, "request": 4, "status": 5, "bytes": 6}},
		{[]string{"request", "authuser", "remotehost", "status", "bytes", "date", "rfc931"},
			map[string]int{"remotehost": 2, "rfc931": 6, "authuser": 1, "date": 5, "request": 0, "status": 3, "bytes": 4}},
	}

	r := NewReader("")
	for _, test := range tests {
		r.parseHeader(test.input)
		if !cmp.Equal(r.logMap, test.want) {
			t.Errorf(`reader.parseHeader(%q) set logMap to %q, want %q`, test.input, r.logMap, test.want)
		}
	}
}

func TestParseLog(t *testing.T) {
	var tests = []struct {
		header []string
		log    []string
		want   LogModel
	}{
		{[]string{"remotehost", "rfc931", "authuser", "date", "request", "status", "bytes"},
			[]string{"10.0.0.2", "-", "apache", "1549573860", "GET /api/user HTTP/1.0", "200", "1234"},
			LogModel{"10.0.0.2", "-", "apache", 1549573860, 200, 1234, "GET /api/user HTTP/1.0", "GET", "/api/user", "/api", "HTTP/1.0"}},
		{[]string{"bytes", "remotehost", "authuser", "rfc931", "status", "request", "date"},
			[]string{"1194", "10.0.0.5", "apache", "-", "500", "POST /report HTTP/1.0", "1549574134"},
			LogModel{"10.0.0.5", "-", "apache", 1549574134, 500, 1194, "POST /report HTTP/1.0", "POST", "/report", "/report", "HTTP/1.0"}},
	}

	r := NewReader("")
	for _, test := range tests {
		r.parseHeader(test.header)
		if got := r.parseLog(test.log); got != test.want {
			t.Errorf(`parseLog(%s) returned %v, want %v`, test.log, got, test.want)
		}
	}
}
