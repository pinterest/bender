package bender

import (
	"time"
	"math/rand"
)

// ExponentialIntervalGenerator creates an IntervalGenerator that outputs exponentially distributed
// intervals. The resulting arrivals constitute a Poisson process. The rate parameter is the average
// queries per second for the generator, and corresponds to the reciprocal of the lambda parameter
// to an exponential distribution. In English, if you want to generate 30 QPS on average, pass 30
// as the value of rate.
func ExponentialIntervalGenerator(rate float64) IntervalGenerator {
	rate = rate / float64(time.Second)
	return func(_ int64) int64 {
		return int64(rand.ExpFloat64() / rate)
	}
}

// UniformIntervalGenerator creates and IntervalGenerator that outputs 1/rate every time it is
// called. Boring, right?
func UniformIntervalGenerator(rate float64) IntervalGenerator {
	irate := int64(rate / float64(time.Second))
	return func(_ int64) int64 {
		return irate
	}
}
