/*
`Reader` is responsible for ingesting the contents of a csv HTTP access log file and parsing
it into suitable format for downstream processes to handle. It is run on a separate thread
(goroutine) and reads the contents into a buffer before being processed. From the example
input file it can be observed that logs are not in a strict order, but it is assumed that
they are in a timely order. To handle this a priority queue has been implemented, of size 50
by default, which is filled to capacity before sending the earliest log line back to the
`Player`. This is effectively a moving window through the csv file which assumes that
timestamp T(n+51) onwards will not be earlier than any time within T(n) to T(n+50).
*/
package main

import (
	"container/heap"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	defaultBufferSize        = 50
	defaultPriorityQueueSize = 50
)

type RequestData struct {
	method   string
	endpoint string
	section  string
	protocol string
}

type LogModel struct {
	remoteHost string
	authServer string
	authUser   string
	date       int64
	status     int
	bytes      int
	request    string
	method     string
	endpoint   string
	section    string
	protocol   string
}

type Reader struct {
	filePath string         // input csv file path
	logMap   map[string]int // maps a request header to the index within a log line
	buff     chan []string  // channel buffer for lines read from csv
}

// NewReader returns a new instance of the Reader.
func NewReader(filePath string) *Reader {
	return &Reader{
		filePath: filePath,
		logMap:   make(map[string]int),
		buff:     make(chan []string, 10),
	}
}

// Process reads the contents of the input file and outputs the results to the out channel.
// Each line is parsed into a LogModel struct. A priority queue is maintained, with the
// earliest timestamp at the front, to handle the input not being in a strict time order.
func (r *Reader) Process(out chan LogModel) {
	queue := make(PriorityQueue, 0)
	header := true
	readBuffer := make(chan []string, defaultBufferSize)

	go r.Read(readBuffer)
	for {
		line, ok := <-readBuffer
		for {
			// Send the next item in the priority queue if the chosen size has been reached.
			if queue.Len() < defaultPriorityQueueSize {
				break
			}
			entry := heap.Pop(&queue).(*LogItem)
			out <- entry.value.(LogModel)
		}
		if ok {
			if header {
				// Process header
				r.parseHeader(line)
				header = false
			} else {
				// Process log line
				logModel := r.parseLog(line)
				heap.Push(&queue, &LogItem{value: logModel, priority: logModel.date})
			}
		} else if queue.Len() > 0 {
			// File fully read, send remaining logs in priority queue
			entry := heap.Pop(&queue).(*LogItem)
			out <- entry.value.(LogModel)
		} else {
			close(out)
			return
		}
	}
}

// Read ingests each line from the csv file and sends the result to a buffer channel
func (r *Reader) Read(buffer chan []string) {
	input, err := os.Open(r.filePath)
	if err != nil {
		log.Fatal("Unable to read file " + r.filePath)
	}
	defer input.Close()

	dec := csv.NewReader(input)
	for {
		record, err := dec.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		buffer <- record
	}
	close(buffer)
}

// parseheader takes the csv header and fills in the logMap which associates a header
// string with an index.
func (r *Reader) parseHeader(header []string) {
	for index, entry := range header {
		r.logMap[entry] = index
	}
}

// parseLog parses a line of the log file and returns a LogModel struct.
func (r *Reader) parseLog(fields []string) LogModel {
	log := LogModel{}
	log.remoteHost = fields[r.logMap["remotehost"]]
	log.authServer = fields[r.logMap["rfc931"]]
	log.authUser = fields[r.logMap["authuser"]]
	t, _ := strconv.Atoi(fields[r.logMap["date"]])
	log.date = int64(t)
	log.status, _ = strconv.Atoi(fields[r.logMap["status"]])
	log.bytes, _ = strconv.Atoi(fields[r.logMap["bytes"]])
	log.request = fields[r.logMap["request"]]
	requestData := parseRequest(log.request)
	log.method = requestData.method
	log.endpoint = requestData.endpoint
	log.section = requestData.section
	log.protocol = requestData.protocol
	return log
}

// parseRequest parses the request string and extracts individual components.
func parseRequest(request string) RequestData {
	result := RequestData{}
	requestSplit := strings.Split(request, " ")
	if len(requestSplit) != 3 {
		return result
	}
	result.method = requestSplit[0]
	result.endpoint = requestSplit[1]
	result.section = getSection(result.endpoint)
	result.protocol = requestSplit[2]
	return result
}

// getSection returns the section from a request endpoint.
func getSection(endpoint string) string {
	if len(endpoint) == 0 || endpoint[0] != '/' {
		return ""
	}
	index := strings.Index(endpoint[1:], "/")
	if index == -1 {
		return endpoint
	}
	return endpoint[0 : index+1]
}
