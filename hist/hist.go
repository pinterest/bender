package hist

import (
	"sort"
	"math"
	"fmt"
)

type Histogram struct {
	max int
	n int
	total int
	values []int
}

func NewHistogram(max int) *Histogram {
	return &Histogram{max, 0, 0, make([]int, max + 1)}
}

func (h *Histogram) Add(v int) {
	if v < 1 {
		h.values[0]++
	} else if v >= h.max {
		h.values[h.max]++
	} else {
		h.values[v]++
	}
	h.n++
	h.total += v
}

func (h *Histogram) Percentiles(percentiles ...float64) []int {
	if percentiles == nil {
		return nil
	}

	sort.Sort(sort.Float64Slice(percentiles))
	pcs := make([]int, len(percentiles))
	for i, percentile := range percentiles {
		if percentile > 1.0 || percentile < 0.0 {
			panic("Percentiles must be in [0.0, 1.0]")
		}

		pcs[i] = int(math.Max(1.0, percentile * float64(h.n)))
	}

	result := make([]int, len(pcs))

	count := 0
	for i, j := 0, 0; i < len(pcs) && j < len(h.values); j++ {
		count += h.values[j]

		// one histogram bucket can match multiple percentiles
		for ; i < len(pcs) && count >= pcs[i]; i++ {
			result[i] = j
		}
	}

	return result
}

func (h *Histogram) Average() float64 {
	return float64(h.total) / float64(h.n)
}

func (h *Histogram) String() string {
	ps := h.Percentiles(0.0, 0.5, 0.9, 0.95, 0.99, 0.999, 0.9999, 1.0)
	s := "Percentiles:\n" +
	     " Min:     %d\n" +
	     " Median:  %d\n" +
	     " 90th:    %d\n" +
	     " 99th:    %d\n" +
	     " 99.9th:  %d\n" +
	     " 99.99th: %d\n" +
	     " Max:     %d\n" +
	     "Stats:\n" +
	     " Average: %f\n"
	return fmt.Sprintf(s, ps[0], ps[1], ps[2], ps[3], ps[4], ps[5], ps[6], h.Average())
}
