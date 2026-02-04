package limiter

import (
	"math"
	"sync"
	"time"
)

type Bucket struct {
	capacity         float64    //max tokens
	refill_rate      float64    //tokens refilled per second
	tokens           float64    //current amount of tokens
	last_refill_time time.Time  //last time token was added
	m                sync.Mutex //avoid race conditions
}

func CreateBucket(capacity float64, refill_rate float64) *Bucket {
	return &Bucket{
		capacity:         capacity,
		tokens:           capacity, //tho we start with the max capacity, burst prob of sliding window is avoided due to the refil rate
		refill_rate:      refill_rate,
		last_refill_time: time.Now(),
	}
}

func (b *Bucket) ReqLimiter() bool {
	//inevitably when used in projects that deal with http req/res
	//go will use go routines; hence race conditon is a risk
	b.m.Lock()         //no other goroutines can access while another is accceing the bucket
	defer b.m.Unlock() //unlock onece the access is done

	now := time.Now()
	ts_last_refill := now.Sub(b.last_refill_time).Seconds()
	//refill
	b.tokens = math.Min(
		//cant exceed max capacity
		b.capacity,
		b.tokens+(ts_last_refill*b.refill_rate),
	)
	b.last_refill_time = time.Now()
	//if theres at least one token we allow the request
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}
