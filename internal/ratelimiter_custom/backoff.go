package ratelimiter_custom

import (
	"math"
	"time"
)

func ExponentialBackoff(n int) time.Duration {
	return time.Duration(math.Pow(2, float64(n/2))) * time.Second
}
