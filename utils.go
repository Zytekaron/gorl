package gorl

import (
	"math"
	"time"
)

// intervalCount counts how many times the interval has completely passed between the start and end.
func intervalCount(start, end time.Time, interval time.Duration) int64 {
	delta := math.Abs(float64(end.Sub(start))) // time difference from start to end
	return int64(delta / float64(interval))    // number of intervals that have passed in that time
}

//// lastBefore returns the most recent time before end which is an exact interval increase from the start.
//func lastBefore(start, end time.Time, interval time.Duration) time.Time {
//	diff := intervalCount(start, end, interval)  // number of times the interval has passed
//	mod := time.Duration(diff * int64(interval)) // amount of time to add to the start time
//
//	t := start.Add(mod)
//	if t.Before(end) {
//		return t
//	}
//	return t.Add(-interval)
//}

// nextAfter returns the most recent time after start which is an exact interval increase from the start.
func nextAfter(start, end time.Time, interval time.Duration) time.Time {
	diff := intervalCount(start, end, interval)  // number of times the interval has passed, +1
	mod := time.Duration(diff * int64(interval)) // amount of time to add to the start time

	t := start.Add(mod)
	if t.After(end) {
		return t
	}
	return t.Add(interval)
}

func min[T int64](a, b T) T {
	if a < b {
		return a
	}
	return b
}
