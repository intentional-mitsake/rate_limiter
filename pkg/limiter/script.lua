--local var are scoped to the script
--all this runs once each call without interruption
--all we are doing here is implementing the same token bucket as the prev in-memory one

--argv is string initially, co conv to numbers 
--we need to assign values from go in order of argv1, argv2, argv3 etc
local capacity = tonumber(ARGV[1]) --max capacity
local refillRate = tonumber(ARGV[2]) --refil rate
local now        = tonumber(ARGV[3]) --time now
local tokenCost       = tonumber(ARGV[4]) --cost of each request

--redis returns smth like this --> data = { [1]="7" [2]="149242987"}
local data = redis.call("HMGET", KEYS[1], "tokens", "last") 
local tokens = tonumber(data[1])--just taking the respective values from the redis.call
local last = tonumber(data[2])

--initially if the bucket is empty then that means its never been used before
--so we fill it to max capacity here
--next time this bucket is used, this con is no true so skips
if tokens == nil then
    tokens = capacity
    last = now
end

--either no time has passed since last req, or it the diff btwn now and last time refiled
local timeElapsed = math.max(0, now-last)
--just the formula for token refil
tokens = math.min(capacity, tokens + timeElapsed * refillRate)

local allowed = 0
if tokens >= tokenCost then
    tokens = tokens - tokenCost
    allowed = 1
end
--setting new values for tokens and last refil time
redis.call("HMSET", KEYS[1], "tokens", tokens, "last", now)
--expire the tokens if their life cylcle is done
--if 100 max cap, and 5 seconds per refil, then life of a token = 100/5 = 20 seconds
redis.call("EXPIRE", KEYS[1], math.ceil(capacity/refillRate))

return {allowed, tokens}