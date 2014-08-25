package bender

import (
	"math"
	"sync"
	"time"
	"fmt"
)

type IntervalGenerator func(int64) int64

type Request struct {
	Rid int64
	Request interface{}
}

type RequestExecutor func(int64, *Request) error

type StartMsg struct {
	Start int64
}

func (m StartMsg) String() string {
	return fmt.Sprintf("StartMsg{start=%d}", m.Start)
}

type EndMsg struct {
	Start, End int64
}

func (m EndMsg) String() string {
	return fmt.Sprintf("EndMsg{start=%d,end=%d}", m.Start, m.End)
}

type WaitMsg struct {
	Wait, Overage int64
}

func (m WaitMsg) String() string {
	return fmt.Sprintf("WaitMsg{wait=%d,overage=%d}", m.Wait, m.Overage)
}

type StartRequestMsg struct {
	Start, Rid int64
}

func (m StartRequestMsg) String() string {
	return fmt.Sprintf("StartRequestMsg{start=%d,requestId=%d}", m.Start, m.Rid)
}

type EndRequestMsg struct {
	Start, End, Rid int64
	Err       error
}

func (m EndRequestMsg) String() string {
	return fmt.Sprintf("EndRequestMsg{start=%d,end=%d,requestId=%d,error=%s}", m.Start, m.End, m.Rid, m.Err)
}

func LoadTestThroughput(intervals IntervalGenerator, requests chan *Request, requestExec RequestExecutor) chan interface{} {
	reporter := make(chan interface{})

	go func() {
		start := time.Now().UnixNano()
		reporter <- &StartMsg{start}

		var wg sync.WaitGroup
		var overage int64
		for request := range requests {
			overageStart := time.Now().UnixNano()

			wait := intervals(overageStart)
			adjust := int64(math.Min(float64(wait), float64(overage)))
			wait -= adjust
			overage -= adjust
			reporter <- &WaitMsg{wait, overage}
			time.Sleep(time.Duration(wait))

			wg.Add(1)
			go func(req *Request) {
				defer wg.Done()
				reqStart := time.Now().UnixNano()
				// TODO(charles): this can block waiting for the recorder, in which case we will
				// count the time against the service, which isn't right. How to make this not
				// block? Create a buffered channel with a really large buffer? Can we make the
				// buffer infinite? Is that wise? Maybe we send Start and End together, after the
				// request is finished? At that point, maybe just send a single message for each
				// request?
				reporter <- &StartRequestMsg{start, request.Rid}
				err := requestExec(time.Now().UnixNano(), req)
				reporter <- &EndRequestMsg{reqStart, time.Now().UnixNano(), request.Rid, err}
			}(request)

			overage += time.Now().UnixNano() - overageStart - wait
		}
		wg.Wait()
		reporter <- &EndMsg{start, time.Now().UnixNano()}
		close(reporter)
	}()

	return reporter
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

func LoadTestConcurrency(workers *WorkerSemaphore, requests chan *Request, requestExec RequestExecutor) chan interface{} {
	reporter := make(chan interface{})

	go func() {
		start := time.Now().UnixNano()
		reporter <- &StartMsg{start}

		var wg sync.WaitGroup
		for request := range requests {
			workers.Wait(1)

			wg.Add(1)
			go func(req *Request) {
				defer func() {
					wg.Done()
					workers.Signal(1)
				}()

				reqStart := time.Now().UnixNano()
				reporter <- &StartRequestMsg{start, request.Rid}
				err := requestExec(time.Now().UnixNano(), req)
				reporter <- &EndRequestMsg{reqStart, time.Now().UnixNano(), request.Rid, err}
			}(request)
		}

		wg.Wait()
		reporter <- &EndMsg{start, time.Now().UnixNano()}
		close(reporter)
	}()

	return reporter
}
