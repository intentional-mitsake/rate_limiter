package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBucket struct {
	rdb *redis.Client
	//redis executes scripts atomically
	//meaning all of the stuff inside the script happens at onec without interruption
	//this is imp as prev while using in memory token bucket, we used mutex to avoid race condtisons
	//here using a lua scrip to read tokens, read last refil, calc refil, max capacity, sub tokens etc is done in one operation
	//as no interuption->no race cond
	//when is the scirpt executed tho?->when called->each request
	//basicallt replacemnet for mutex
	script *redis.Script
	//max cap
	cap float64
	//refil rate
	rate float64
}

func CreateRedisBucket(rdb *redis.Client, cap, rate float64) *RedisBucket {
	return &RedisBucket{
		rdb:    rdb,
		script: redis.NewScript("script.lua"),
		cap:    cap,
		rate:   rate,
	}
}

func (b *RedisBucket) ReqLimiter(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	//this takes (context.Context, clinet cmder, keys []string, args ..interface{})
	//ctx is the controller, client(rdb here) is smth that can send a comand to redis and return the result
	//rdb inside the run just tells which rdb connection to use, access to fommadns like (HMGET)
	res, err := b.script.Run(
		ctx,           //bascially a cancellation and timeout controler
		b.rdb,         //which rdb to execute this on
		[]string{key}, //auto recognized as a key
		b.cap,         //from her its argv1 to argv4
		b.rate,
		now,
		1,
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
