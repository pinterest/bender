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
	t int64
}

func (m StartMsg) String() string {
	return fmt.Sprintf("StartMsg{time=%d}", m.t)
}

type EndMsg struct {
	s, t int64
}

func (m EndMsg) String() string {
	return fmt.Sprintf("EndMsg{start=%d,end=%d}", m.s, m.t)
}

type WaitMsg struct {
	w, o int64
}

func (m WaitMsg) String() string {
	return fmt.Sprintf("WaitMsg{wait=%d,overage=%d}", m.w, m.o)
}

type StartRequestMsg struct {
	s, rid int64
}

func (m StartRequestMsg) String() string {
	return fmt.Sprintf("StartRequestMsg{start=%d,requestId=%d}", m.s, m.rid)
}

type EndRequestMsg struct {
	s, t, rid int64
	err       error
}

func (m EndRequestMsg) String() string {
	return fmt.Sprintf("EndRequestMsg{start=%d,end=%d,requestId=%d,error=%s}", m.s, m.t, m.rid, m.err)
}

func Bender(intervals chan int64, requests chan *Request, requestExec RequestExecutor) chan interface{} {
	reporter := make(chan interface{})

	go func() {
		start := time.Now().UnixNano()
		reporter <- StartMsg{start}

		var wg sync.WaitGroup
		var overage int64
		for request := range requests {
			overageStart := time.Now().UnixNano()

			wait := <-intervals
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
