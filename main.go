package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/intentional-mitsake/rate_limiter/pkg/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
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
		allowed, err := bucket.ReqLimiter(ctx, key1)
		if err != nil {
			fmt.Println("Error:", err)
		}
		if !allowed {
			fmt.Printf("Expected request %d to be allowed for %s\n", i+1, key1)
		}
	}

	// MULTI-USER BLOCK TEST
	//.ie. if one user has hit their max cap, another user shouldnt be blocked
	for i := 0; i < int(capacity); i++ {
		allowed, err := bucket.ReqLimiter(ctx, key2)
		if err != nil {
			fmt.Println("Error:", err)
		}
		if !allowed {
			fmt.Printf("Expected request %d to be allowed for %s\n", i+1, key2)
		}
	}

	// BLOCK AFTER LIMIT REACHED TEST
	allowed, _ := bucket.ReqLimiter(ctx, key2)
	if allowed {
		fmt.Printf("Expected request beyond capacity to be blocked for %s\n", key2)
	}

	// REFILL AFTER SLEEP TEST
	time.Sleep(1 * time.Second) // allow refill
	allowed, _ = bucket.ReqLimiter(ctx, key1)
	if allowed {
		fmt.Printf("Request allowed after sleep for %s (refill works)\n", key1)
	}

	//CONCURRENT REQUESTS TEST
	//chatgpteed this one but in hindsight looks simple enough
	users := []string{"user:1", "user:2"} //two new users
	requestsPerUser := 12                 // trying more than capacity
	var wg sync.WaitGroup                 //counter to help wait to goroutines to finish before continkuing
	//incr for each goroutine added, decr for each finish, wiat till zero

	for _, u := range users {
		rdb.Del(context.Background(), u)
	}

	fmt.Println("Starting concurrent requests...")

	for _, user := range users {
		//going for more than 10(max cap) to see perfomance
		for i := 1; i <= requestsPerUser; i++ {
			wg.Add(1) //incremnet the counter by one for new goroutine launched
			//goroutine definition-->takes the reqnum and the users name/id
			//this loop wont be waiting till this func is done,
			//this func will start with the reqNum of this iteration and move on to next iteration
			//without waitng for this iteratioins func to end execution
			//pretty much async/await
			//multi goroutines run parallel simulating concurrent requests
			go func(u string, reqNum int) {
				defer wg.Done() //decrement the counter after this go routine is done
				allowed, err := bucket.ReqLimiter(context.Background(), u)
				if err != nil {
					fmt.Printf("[%s] Request %d error: %v\n", u, reqNum, err)
					return
				}
				fmt.Printf("[%s] Request %2d allowed? %v\n", u, reqNum, allowed)
			}(user, i) //schedules teh anon goroutine immediately to run concurently using current user and i(reqnum) as parameters
		}
		//at the end whats basically happening in thsi nested loop is:
		//the first loop goes over the tow users, second one runs for each user
		//in the second loop for ech user, all their respective requests are added to the wait group
		//and lauched immediately---THIS DOESNT MEAN ITS EXECUTEED IMMEDIATELY--
		//cuz the loop doesnt wait to execute each req before moving on to next
		//all it does is :
		//add to wg, schedules teh goroutine,(NO WAIT FOR GOROUTINE TO EXECUTE) moves to next itertinl
		//this way all req are added to the wg and scheduled pretyy much immediately
		//a scheduler decides when to run each goroutine
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Concurrent test finished")
}
