/*
Copyright 2014 Pinterest.com

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
	"log"
	"github.com/Pinterest/bender/hist"
)

type Recorder func(interface{})

func Record(c chan interface{}, recorders... Recorder) {
	for msg := range c {
		for _, recorder := range recorders {
			recorder(msg)
		}
	}
}

func logMessage(l *log.Logger, msg interface{}) {
	l.Printf("%+v", msg)
}

func NewLoggingRecorder(l *log.Logger) Recorder {
	return func(msg interface{}) {
		logMessage(l, msg)
	}
}

func NewHistogramRecorder(h *hist.Histogram) Recorder {
	return func(msg interface{}) {
		switch msg := msg.(type) {
		case *StartEvent:
			h.Start(int(msg.Start))
		case *EndEvent:
			h.End(int(msg.End))
		case *EndRequestEvent:
			elapsed := int(msg.End - msg.Start)
			if msg.Err == nil {
				h.Add(elapsed)
			} else {
				h.AddError(elapsed)
			}
		}
	}
}

