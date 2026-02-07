package limiter

import (
	"context"
	"time"

	_ "embed"

	"github.com/redis/go-redis/v9"
)

type RedisBucket struct {
	Rdb *redis.Client
	//redis executes scripts atomically
	//meaning all of the stuff inside the script happens at onec without interruption
	//this is imp as prev while using in memory token bucket, we used mutex to avoid race condtisons
	//here using a lua scrip to read tokens, read last refil, calc refil, max capacity, sub tokens etc is done in one operation
	//as no interuption->no race cond
	//when is the scirpt executed tho?->when called->each request
	//basicallt replacemnet for mutex
	Script *redis.Script
	//max cap
	Cap float64
	//refil rate
	Rate float64
}

//go:embed script.lua
var luascript string

func CreateRedisBucket(rdb *redis.Client, cap, rate float64) *RedisBucket {
	return &RedisBucket{
		Rdb:    rdb,
		Script: redis.NewScript(luascript),
		Cap:    cap,
		Rate:   rate,
	}
}

func (b *RedisBucket) ReqLimiter(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	//this takes (context.Context, clinet cmder, keys []string, args ..interface{})
	//ctx is the controller, client(rdb here) is smth that can send a comand to redis and return the result
	//rdb inside the run just tells which rdb connection to use, access to fommadns like (HMGET)
	res, err := b.Script.Run(
		ctx,   //bascially a cancellation and timeout controler
		b.Rdb, //which rdb to execute this on
		//btw, we will be passing each users id or smth wrapping the last and tokens here
		//meaning each key-value stored is separate for each user
		[]string{key}, //auto recognized as a key
		b.Cap,         //from her its argv1 to argv4
		b.Rate,
		now,
		1,
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
