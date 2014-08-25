package bender

import (
	"testing"
	"reflect"
	"errors"
)

func assertMessages(t *testing.T, cr chan interface{}, expected_msgs... interface{}) {
	for _, msg := range expected_msgs {
		actual_msg, ok := <-cr
		if !ok {
			t.Errorf("Expected a message (%s), but reached end of channel instead", msg)
			return
		}

		if reflect.TypeOf(actual_msg) != reflect.TypeOf(msg) {
			t.Errorf("Expected a message of type %s, but got a message of type %s instead", reflect.TypeOf(actual_msg), reflect.TypeOf(msg))
			return
		}

		switch m := actual_msg.(type) {
		case *EndRequestMsg:
			if m.Err != nil && msg.(*EndRequestMsg).Err == nil {
				t.Errorf("Expected EndRequestMsg with no error (%s), but got EndRequestMsg with an error (%s)", msg, m)
			}

			if m.Err == nil && msg.(*EndRequestMsg).Err != nil {
				t.Errorf("Expected EndRequestMsg with an error (%s), but got EndRequestMsg with no error (%s)", msg, m)
			}
		}
	}
}

func requests(rs... *Request) chan *Request {
	c := make(chan *Request)
	go func() {
		for _, r := range rs {
			c <- r
		}
		close(c)
	}()
	return c
}

func workers(n int) *WorkerSemaphore {
	s := NewWorkerSemaphore()
	go func() {
		s.Signal(n)
	}()
	return s
}

func noOpExec(int64, *Request) error {
	return nil
}

func errorExec(int64, *Request) error {
	return errors.New("fake error")
}

func TestLoadTestThroughputNoRequests(t *testing.T) {
	cr := LoadTestThroughput(UniformIntervalGenerator(0), requests(), noOpExec)
	assertMessages(t, cr, &StartMsg{}, &EndMsg{})
}

func TestLoadTestThroughputOneSuccess(t *testing.T) {
	cr := LoadTestThroughput(UniformIntervalGenerator(0), requests(&Request{}), noOpExec)
	assertMessages(t, cr, &StartMsg{}, &WaitMsg{}, &StartRequestMsg{}, &EndRequestMsg{}, &EndMsg{})
}

func TestLoadTestThroughputOneError(t *testing.T) {
	cr := LoadTestThroughput(UniformIntervalGenerator(0), requests(&Request{}), errorExec)
	assertMessages(t, cr, &StartMsg{}, &WaitMsg{}, &StartRequestMsg{}, &EndRequestMsg{Err:errors.New("foo")}, &EndMsg{})
}

func TestLoadTestConcurrencyNoRequests(t *testing.T) {
	cr := LoadTestConcurrency(workers(1), requests(), noOpExec)
	assertMessages(t, cr, &StartMsg{}, &EndMsg{})
}

func TestLoadTestConcurrencyOneSuccess(t *testing.T) {
	cr := LoadTestConcurrency(workers(1), requests(&Request{}), noOpExec)
	assertMessages(t, cr, &StartMsg{}, &StartRequestMsg{}, &EndRequestMsg{}, &EndMsg{})
}

func TestLoadTestConcurrencyOneError(t *testing.T) {
	cr := LoadTestConcurrency(workers(1), requests(&Request{}), errorExec)
	assertMessages(t, cr, &StartMsg{}, &StartRequestMsg{}, &EndRequestMsg{Err:errors.New("foo")}, &EndMsg{})
}
