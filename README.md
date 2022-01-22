# http-log-monitor

http-log-monitor is a Go HTTP log monitoring console programme. It is able to read CSV-encoded HTTP access logs, display the most popular endpoint sections over a specified time period, and alert when traffic passes a certain threshold across a given time period.

I am new to Go and have taken this assignment as an opportunity to improve my proficiency with the language.

## Installation

http-log-monitor was developed and tested with `Go 1.17.6`. Please install the latest version for your target platform:
```
https://go.dev/doc/install
```

Navigate to the `src` directory and run the following command to download the required modules:

```
$ cd src
$ go mod download
```

Build an executable of the project: 

```
$ go build -o http-log-monitor
```

## Usage

It is suggested that the executable be initially run with the `-h` flag to display the available command line flags:

```
$ ./http-log-monitor -h
Usage of ./http-log-monitor:
  -alert int
        duration of the high traffic alert window in seconds (default 120)
  -input string
        input log file path (required)
  -rps int
        average requests per second threshold for high traffic alert (default 10)
  -stats int
        time interval between displaying stats in seconds (default 10)
```

Output messages take the form:

```
event type | event timestamp | event message
```

Execute the programme, including the required `-input` input file path:

```
$ ./http-log-monitor -input ../input/sample_csv.txt
[STATS] 1549573939      /api: 147 /report: 31 
[STATS] 1549573949      /api: 152 /report: 29 
[ALERT] 1549573957      High traffic generated an alert - hits = 1206
[STATS] 1549573959      /api: 150 /report: 32 
...
```

## Testing
Tests are executed using the following:

```
$ go test -v ./... -cover
```

A visual representation of code coverage for each file can be created and viewed in a web browser:

```
$ go test -coverprofile=c.out
$ go tool cover -html=c.out
```

## Design

The following assumptions have been made:
* Whilst logs are not in a strict time order, it is assumed that they arrive in a timely manner. This is explained more in the `Reader` section.
* The input log is well formed and all headers listed in the sample input are present.
* The request string is formatted in the order: method, endpoint, protocol (e.g. `"GET /api/user HTTP/1.0"`)

### Player

The `Player` is the controller which facilitates communication between the `Reader`, `Stats` and `Monitor` components. It is responsible for simulating the passage of time. A requirement is that any alerts must be accurate to within a second, therefore a second represents a single unit of time. As the log file is read, access hits during the current second are registered with the `Stats` and `Monitor`. When a timestamp in the log is greater than the current second the `Player` “ticks” time forward for each second until it is synchronised with the latest timestamp.

### Reader

The `Reader` component is responsible for ingesting the contents of a csv log file and parsing it into a suitable format for downstream processes to handle. It is run on a separate thread (goroutine) and reads the contents into a buffer before being processed. From the example input file it can be observed that logs are not in a strict order, but it is assumed that they are in a timely order. To handle this a priority queue has been implemented, of size 50 by default, which is filled to capacity before sending the earliest log line back to the `Player`. This is effectively a moving window through the csv file which assumes that timestamp T<sub>n+51</sub> onwards will not be earlier than any time within T<sub>n</sub> to T<sub>n+50</sub>.

### Monitor

The `Monitor` is responsible for high traffic and recovery alerts. It is initialised with a duration and average request per second value. A FIFO queue is used of size duration where each entry holds the number of hits for a second of time. As time ticks forward the number of hits for this second are appended to the end. Once the queue reaches capacity, subsequent appends cause the front entry to be popped. This allows the total number of hits for the chosen duration to be efficiently maintained. The time complexity for insertions and removals is O(1), whilst the required space is O(n) where n is the number of seconds in the alert window. 

### Stats

`Stats` maintains a ranking of the top sections based on the number of hits. Two data structures are used to efficiently perform this. An unordered map, with section as key and hits as value, keeps count of each section’s total number of hits. An ordered map, with hits as key and sections as value, tracks the top sections. Updating the former in O(1) time allows the latter to be updated in O(log n) time where n is the number of unique sections. The space required is O(n).

## Improvements

* Currently all data is held in memory. An improvement would be to store processed data in a log or database table such that a crash or loss of service could be recovered by another instance. 
* Functionality such as reporting statistics could be partitioned into 10 second intervals and processed by separate instances in parallel. A message queue could be created which manages these jobs for parallel workers to process.
* A better solution should be explored for handling out of order timestamps. The method presented here uses an arbitrary number for the size of the priority queue but if there is a significant number of access requests in a short period of time it will lead to late arrivals being rejected. Reading ahead by a certain number of seconds was considered but this would be unbounded and could lead to an extremely large queue held in memory.
* Many of the components have default values defined as constants which could be exposed for the user to configure.
* The solution makes the assumption that a second is a single unit of time. This should be configurable by the user as a future requirement may be to improve the accuracy of high traffic alerting to less than a second.
* Both `Stats` and `Monitor` could have a more general API to allow any attribute of a HTTP access request to be tracked and alerted.
