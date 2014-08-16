package bender

import (
	"math"
	"sync"
	"time"
	"fmt"
)

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

func LoadTestThroughput(intervals chan int64, requests chan *Request, requestExec RequestExecutor) chan interface{} {
	reporter := make(chan interface{})

	go func() {
		start := time.Now().UnixNano()
		reporter <- StartMsg{start}

		var wg sync.WaitGroup
		var overage int64
		for request := range requests {
			overageStart := time.Now().UnixNano()

			wait, more := <-intervals
			if !more {
				break
			}

			adjust := int64(math.Min(float64(wait), float64(overage)))
			wait -= adjust
			overage -= adjust
			reporter <- WaitMsg{wait, overage}
			time.Sleep(time.Duration(wait))

			wg.Add(1)
			go func(req *Request) {
				defer wg.Done()
				reqStart := time.Now().UnixNano()
				reporter <- StartRequestMsg{start, request.Rid}
				err := requestExec(time.Now().UnixNano(), req)
				reporter <- EndRequestMsg{reqStart, time.Now().UnixNano(), request.Rid, err}
			}(request)

			overage += time.Now().UnixNano() - overageStart - wait
		}
		wg.Wait()
		reporter <- EndMsg{start, time.Now().UnixNano()}
		close(reporter)
	}()

	return reporter
}

func LoadTestConcurrency(starts chan int64, requests chan *Request, requestExec RequestExecutor) chan interface{} {
	reporter := make(chan interface{})
	permits := make(chan bool)

	go func() {
		for n := range starts {
			for i := 0; i < n; i++ {
				permits <- true
			}
		}
		close(permits)
	}()

	go func() {
		start := time.Now().UnixNano()
		reporter <- StartMsg{start}

		var wg sync.WaitGroup
		for request := range requests {
			<-permits

			wg.Add(1)
			go func(req *Request) {
				defer func() {
					wg.Done()
					permits <- true
				}()

				reqStart := time.Now().UnixNano()
				reporter <- StartRequestMsg{start, request.Rid}
				err := requestExec(time.Now().UnixNano(), req)
				reporter <- EndRequestMsg{reqStart, time.Now().UnixNano(), request.Rid, err}
			}(request)
		}

		wg.Wait()
		reporter <- EndMsg{start, time.Now().UnixNano()}
		close(reporter)
	}()

	return reporter
}
