package bender

import (
	"math"
	"sync"
	"time"
)

type IntervalGenerator func(int64) int64

type RequestExecutor func(int64, interface{}) (interface{}, error)

// StartEvent is sent once at the start of the load test.
type StartEvent struct {
	// The Unix epoch time in nanoseconds at which the load test started.
	Start int64
}

// EndEvent is sent once at the end of the load test, after which no more events are sent.
type EndEvent struct {
	// The Unix epoch times in nanoseconds at which the load test started and ended.
	Start, End int64
}

// WaitEvent is sent once for each request before sleeping for the given interval.
type WaitEvent struct {
	// The next wait time (in nanoseconds) and the accumulated overage time (the difference between
	// the actual wait time and the intended wait time).
	Wait, Overage int64
}

// StartRequestEvent is sent before a request is executed. The sending of this event happens before
// the timing of the request starts, to avoid potential issues, so it contains the timestamp of the
// event send, and not the timestamp of the request start.
type StartRequestEvent struct {
	// The Unix epoch time (in nanoseconds) at which this event was created, which will be earlier
	// than the sending of the associated request (for performance reasons)
	Time int64
	// The request that will be sent, nothing good can come from modifying it
	Request interface{}
}

// EndRequestEvent is sent after a request has completed.
type EndRequestEvent struct {
	// The Unix epoch times (in nanoseconds) at which the request was started and finished
	Start, End int64
	// The response data returned by the request executor
	Response interface{}
	// An error or nil if there was no error
	Err       error
}

func LoadTestThroughput(intervals IntervalGenerator, requests chan interface{}, requestExec RequestExecutor, reporter chan interface{}) {
	go func() {
		start := time.Now().UnixNano()
		reporter <- &StartEvent{start}

		var wg sync.WaitGroup
		var overage int64
		for request := range requests {
			overageStart := time.Now().UnixNano()

			wait := intervals(overageStart)
			adjust := int64(math.Min(float64(wait), float64(overage)))
			wait -= adjust
			overage -= adjust
			reporter <- &WaitEvent{wait, overage}
			time.Sleep(time.Duration(wait))

			wg.Add(1)
			go func(req interface{}) {
				defer wg.Done()
				reporter <- &StartRequestEvent{time.Now().UnixNano(), req}
				reqStart := time.Now().UnixNano()
				res, err := requestExec(time.Now().UnixNano(), req)
				reporter <- &EndRequestEvent{reqStart, time.Now().UnixNano(), res, err}
			}(request)

			overage += time.Now().UnixNano() - overageStart - wait
		}
		wg.Wait()
		reporter <- &EndEvent{start, time.Now().UnixNano()}
		close(reporter)
	}()
}

type empty struct{}
type WorkerSemaphore struct {
	permits chan empty
}

func NewWorkerSemaphore() *WorkerSemaphore {
	return &WorkerSemaphore{permits:make(chan empty)}
}

func (s WorkerSemaphore) Signal(n int) {
	e := empty{}
	for i := 0; i < n; i++ {
		s.permits <- e
	}
}

func (s WorkerSemaphore) Wait(n int) bool {
	for i := 0; i < n; i++ {
		<-s.permits
	}
	return true
}

func LoadTestConcurrency(workers *WorkerSemaphore, requests chan interface{}, requestExec RequestExecutor, reporter chan interface{}) {
	go func() {
		start := time.Now().UnixNano()
		reporter <- &StartEvent{start}

		var wg sync.WaitGroup
		for request := range requests {
			workers.Wait(1)

			wg.Add(1)
			go func(req interface{}) {
				defer func() {
					wg.Done()
					workers.Signal(1)
				}()

				reqStart := time.Now().UnixNano()
				reporter <- &StartRequestEvent{start, req}
				res, err := requestExec(time.Now().UnixNano(), req)
				reporter <- &EndRequestEvent{reqStart, time.Now().UnixNano(), res, err}
			}(request)
		}

		wg.Wait()
		reporter <- &EndEvent{start, time.Now().UnixNano()}
		close(reporter)
	}()
}
