/*
Copyright 2014-2016 Pinterest, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bender

import (
	"errors"
	"reflect"
	"testing"
)

type Request struct{}

func assertMessages(t *testing.T, cr chan interface{}, expected ...interface{}) {
	for _, msg := range expected {
		actual, ok := <-cr
		if !ok {
			t.Errorf("Expected a message (%s), but reached end of channel instead", msg)
			return
		}

		if reflect.TypeOf(actual) != reflect.TypeOf(msg) {
			t.Errorf("Expected a message of type %s, but got a message of type %s instead", reflect.TypeOf(actual), reflect.TypeOf(msg))
			return
		}

		switch m := actual.(type) {
		case *EndRequestEvent:
			if m.Err != nil && msg.(*EndRequestEvent).Err == nil {
				t.Errorf("Expected EndRequestEvent with no error (%+v), but got EndRequestEvent with an error (%+v)", msg, m)
			}

			if m.Err == nil && msg.(*EndRequestEvent).Err != nil {
				t.Errorf("Expected EndRequestEvent with an error (%+v), but got EndRequestEvent with no error (%+v)", msg, m)
			}
		}
	}
}

func requests(rs ...interface{}) chan interface{} {
	c := make(chan interface{})
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

func noOpExec(int64, interface{}) (interface{}, error) {
	return nil, nil
}

func errorExec(int64, interface{}) (interface{}, error) {
	return nil, errors.New("fake error")
}

func TestLoadTestThroughputNoRequests(t *testing.T) {
	cr := make(chan interface{})
	LoadTestThroughput(UniformIntervalGenerator(1e9), requests(), noOpExec, cr)
	assertMessages(t, cr, &StartEvent{}, &EndEvent{})
}

func TestLoadTestThroughputOneSuccess(t *testing.T) {
	cr := make(chan interface{})
	LoadTestThroughput(UniformIntervalGenerator(1e9), requests(Request{}), noOpExec, cr)
	assertMessages(t, cr, &StartEvent{}, &WaitEvent{}, &StartRequestEvent{}, &EndRequestEvent{}, &EndEvent{})
}

func TestLoadTestThroughputOneError(t *testing.T) {
	cr := make(chan interface{})
	LoadTestThroughput(UniformIntervalGenerator(1e9), requests(Request{}), errorExec, cr)
	assertMessages(t, cr, &StartEvent{}, &WaitEvent{}, &StartRequestEvent{}, &EndRequestEvent{Err: errors.New("foo")}, &EndEvent{})
}

func TestLoadTestConcurrencyNoRequests(t *testing.T) {
	cr := make(chan interface{})
	LoadTestConcurrency(workers(1), requests(), noOpExec, cr)
	assertMessages(t, cr, &StartEvent{}, &EndEvent{})
}

func TestLoadTestConcurrencyOneSuccess(t *testing.T) {
	cr := make(chan interface{})
	LoadTestConcurrency(workers(1), requests(Request{}), noOpExec, cr)
	assertMessages(t, cr, &StartEvent{}, &StartRequestEvent{}, &EndRequestEvent{}, &EndEvent{})
}

func TestLoadTestConcurrencyOneError(t *testing.T) {
	cr := make(chan interface{})
	LoadTestConcurrency(workers(1), requests(Request{}), errorExec, cr)
	assertMessages(t, cr, &StartEvent{}, &StartRequestEvent{}, &EndRequestEvent{Err: errors.New("foo")}, &EndEvent{})
}
