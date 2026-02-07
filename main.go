package main

import (
	"context"
	"fmt"
	"time"

	"github.com/intentional-mitsake/rate_limiter/pkg/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	bucket := limiter.CreateRedisBucket(rdb, 10, 2)
	key1 := "user:test" //user 1
	key2 := "user:2"    //user 2

	rdb.Del(context.Background(), key1)
	rdb.Del(context.Background(), key2)
	//MAX CAPACITY TEST
	for i := 0; i < 10; i++ {
		allowed, err := bucket.ReqLimiter(context.Background(), key1)
		if err != nil {
			fmt.Println(err.Error())
		}
		if !allowed {
			fmt.Printf("expected req %d to be allowed", i+1)
		}
	}
	//MULTI-USER BLOCK TEST
	for i := 0; i < 10; i++ {
		other, err := bucket.ReqLimiter(context.Background(), key2)
		if err != nil {
			fmt.Println(err.Error())
		}
		if !other {
			fmt.Printf("expected req %d to be allowed", i+1)
		}
	}
	//BLCOK AFTER LIMIT REACHED TEST
	allowed, _ := bucket.ReqLimiter(context.Background(), key2)
	if allowed {
		fmt.Println("expected this req to be blocked")
	}
	//REFILL AFTER SLEEP TEST
	time.Sleep(1 * time.Second) //refil tokens
	afterrest, _ := bucket.ReqLimiter(context.Background(), key1)
	if afterrest {
		fmt.Println("req allowed after sleep")
	}
	//CONCURRENT REQUESTS TEST
}
