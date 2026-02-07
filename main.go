package main

import (
	"context"
	"sync"
	"time"

	"fmt"

	"github.com/intentional-mitsake/rate_limiter/pkg/limiter"
	"github.com/intentional-mitsake/rate_limiter/pkg/utils"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := utils.CreateLogger()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// Create Redis-backed token bucket
	capacity := float64(10)
	refillRate := float64(2) // tokens per second
	bucket := limiter.CreateRedisBucket(rdb, capacity, refillRate)

	// Define user keys
	key1 := "user:test" // user 1
	key2 := "user:2"    // user 2

	// Clear previous token state
	ctx := context.Background()
	rdb.Del(ctx, key1)
	rdb.Del(ctx, key2)

	// MAX CAPACITY TEST
	for i := 0; i < int(capacity); i++ {
		allowed, _, err := bucket.ReqLimiter(ctx, key1)
		if err != nil {
			logger.Error(err.Error())
		}
		if !allowed {
			logger.Warning("Expected request beyond capacity for " + key1)
		}
		//logger.Info(fmt.Sprintf("Tokens Left: %f", tokens))
	}

	// MULTI-USER BLOCK TEST
	for i := 0; i < int(capacity); i++ {
		allowed, _, err := bucket.ReqLimiter(ctx, key2)
		if err != nil {
			logger.Error(err.Error())
		}
		if !allowed {
			logger.Warning("Expected request beyond capacity for " + key2)
		}
		//logger.Info(fmt.Sprintf("Tokens Left: %f", tokens))
	}

	// BLOCK AFTER LIMIT REACHED TEST
	allowed, _, _ := bucket.ReqLimiter(ctx, key2)
	if allowed {
		logger.Warning("Expected request beyond capacity to be blocked for " + key2)
	}
	//logger.Info(fmt.Sprintf("Tokens Left: %f", tokens))

	// REFILL AFTER SLEEP TEST
	time.Sleep(1 * time.Second) // allow refill
	allowed, _, _ = bucket.ReqLimiter(ctx, key1)
	if allowed {
		logger.Info("Request allowed after sleep for " + key1 + " (refill works)")
	}
	//logger.Info(fmt.Sprintf("Tokens Left: %f", tokens))
	// CONCURRENT REQUESTS TEST
	users := []string{"user:1", "user:2"} // two new users
	requestsPerUser := 12                 // trying more than capacity
	var wg sync.WaitGroup                 // counter to help wait for goroutines to finish

	for _, u := range users {
		rdb.Del(context.Background(), u)
	}

	logger.Info("Starting concurrent requests...")

	for _, user := range users {
		for i := 1; i <= requestsPerUser; i++ {
			wg.Add(1)
			go func(u string, reqNum int) {
				defer wg.Done()
				allowed, tokens, err := bucket.ReqLimiter(context.Background(), u)
				if err != nil {
					fmt.Printf("[%s] Request %d allowed? %s\n", user, reqNum, err.Error())
					return
				}
				fmt.Printf("[%s] Request %d allowed? %v\n", user, reqNum, allowed)
				logger.Info(fmt.Sprintf("Tokens Left: %f", tokens))
			}(user, i)
		}
	}

	wg.Wait()
	logger.Info("Concurrent test finished")
}
