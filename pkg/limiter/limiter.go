package limiter

import (
	"time"
)

type Bucket struct {
	capacity         float64   //max tokens
	refill_rate      float64   //tokens refilled per second
	tokens           float64   //current amount of tokens
	last_refill_time time.Time //last time token was added
}

func CreateBucket(capacity float64, refill_rate float64) *Bucket {
	return &Bucket{
		capacity:         capacity,
		tokens:           capacity, //tho we start with the max capacity, burst prob of sliding window is avoided due to the refil rate
		refill_rate:      refill_rate,
		last_refill_time: time.Now(),
	}
}
