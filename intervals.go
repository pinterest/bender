package bender

import (
	"time"
	"math/rand"
)

func ExponentialIntervals(rate float64) chan int64 {
	rate = rate / float64(time.Second)
	c := make(chan int64, 2)
	go func() {
		for {
			c <- int64(rand.ExpFloat64() / rate)
		}
	}()
	return c
}

func UniformIntervals(rate float64) chan int64 {
	rate = rate / float64(time.Second)
	c := make(chan int64, 2)
	go func() {
		for {
			c <- int64(rate)
		}
	}()
	return c
}

