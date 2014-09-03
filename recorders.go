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
			if msg.Err == nil {
				h.Add(int(msg.End - msg.Start))
			} else {
				h.AddError()
			}
		}
	}
}

