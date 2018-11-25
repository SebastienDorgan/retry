package retry

import (
	"math"
	"time"
)

//UniformStrategy keep same delay between attempts
func UniformStrategy(attempt uint64, interval time.Duration) time.Duration {
	return interval
}

//ExponentialStrategy return an exponential backoff strategy with a growing factor f
func ExponentialStrategy(f float64) BackoffStrategy {
	return func(attempt uint64, interval time.Duration) time.Duration {
		//interval*Pow(f, i)
		return time.Duration(float64(interval.Nanoseconds())*math.Pow(f, float64(attempt))) * time.Nanosecond
	}
}
