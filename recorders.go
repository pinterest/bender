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
	switch msg := msg.(type) {
	case StartMsg:
		l.Printf("Start time:%d\n", msg.Start)
	case EndMsg:
		l.Printf("End start:%d end:%d\n", msg.Start, msg.End)
	case WaitMsg:
		l.Printf("Wait duration:%d overage:%d\n", msg.Wait, msg.Overage)
	case StartRequestMsg:
		l.Printf("StartRequest time:%d, rid:%d\n", msg.Start, msg.Rid)
	case EndRequestMsg:
		if msg.Err == nil {
			l.Printf("EndRequest start:%d end:%d duration:%d rid:%d\n", msg.Start, msg.End, msg.End-msg.Start, msg.Rid)
		} else {
			l.Printf("EndRequest start:%d end:%d duration:%d rid:%d err: %s\n", msg.Start, msg.End, msg.End-msg.Start, msg.Rid, msg.Err)
		}
	default:
		l.Printf("Unknown %s", msg)
	}
}

func NewLoggingRecorder(l *log.Logger) Recorder {
	return func(msg interface{}) {
		logMessage(l, msg)
	}
}

func NewHistogramRecorder(max int, c chan *hist.Histogram) Recorder {
	h := hist.NewHistogram(max)

	return func(msg interface{}) {
		switch msg := msg.(type) {
		case EndRequestMsg:
			if msg.Err == nil {
				h.Add(int(msg.End - msg.Start))
			}
		case EndMsg:
			c <- h
			close(c)
		}
	}
}

