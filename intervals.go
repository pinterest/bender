package bender

import (
	"time"
	"math/rand"
)

func ExponentialIntervalGenerator(rate float64) IntervalGenerator {
	rate = rate / float64(time.Second)
	return func(_ int64) int64 {
		return int64(rand.ExpFloat64() / rate)
	}
}

func UniformIntervalGenerator(rate float64) IntervalGenerator {
	irate := int64(rate / float64(time.Second))
	return func(_ int64) int64 {
		return irate
	}
}

