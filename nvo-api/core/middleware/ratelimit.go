package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nvo-api/core/log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type bucket struct {
	tokens     int
	lastUpdate time.Time
	lastAccess time.Time
}

// RateLimiter - 限流器接口
type RateLimiter interface {
	Allow(key string) (bool, error)
}

// MemoryRateLimiter 基于内存的令牌桶限流器（单机）
type MemoryRateLimiter struct {
	rate     int
	capacity int
	buckets  map[string]*bucket
	mu       sync.RWMutex
	ttl      time.Duration
	stopChan chan struct{}
}

func NewMemoryRateLimiter(rate, capacity int) *MemoryRateLimiter {
	rl := &MemoryRateLimiter{
		rate:     rate,
		capacity: capacity,
		buckets:  make(map[string]*bucket),
		ttl:      5 * time.Minute,
		stopChan: make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *MemoryRateLimiter) Stop() {
	close(rl.stopChan)
}

func (rl *MemoryRateLimiter) Allow(key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     rl.capacity - 1,
			lastUpdate: now,
			lastAccess: now,
		}
		rl.buckets[key] = b
		return true, nil
	}

	elapsed := now.Sub(b.lastUpdate)
	newTokens := int(elapsed.Seconds() * float64(rl.rate))
	b.tokens += newTokens
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastUpdate = now
	b.lastAccess = now

	if b.tokens > 0 {
		b.tokens--
		return true, nil
	}

	return false, nil
}

func (rl *MemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, b := range rl.buckets {
				if now.Sub(b.lastAccess) > rl.ttl {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopChan:
			return
		}
	}
}

// RedisRateLimiter 基于 Redis 的限流器（分布式）
type RedisRateLimiter struct {
	client     *redis.Client
	capacity   int
	window     time.Duration
	expiration time.Duration
}

func NewRedisRateLimiter(client *redis.Client, capacity int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:     client,
		capacity:   capacity,
		window:     window,
		expiration: window + time.Second,
	}
}

func (rl *RedisRateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixNano()
	windowKey := fmt.Sprintf("ratelimit:%s", key)
	windowStart := now - int64(rl.window)

	pipe := rl.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, windowKey)
	pipe.ZAdd(ctx, windowKey, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, windowKey, rl.expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := countCmd.Val()
	return count < int64(rl.capacity), nil
}

// AutoFallbackRateLimiter 自动降级限流器（Redis 失败时降级到内存）
type AutoFallbackRateLimiter struct {
	redis      *RedisRateLimiter
	memory     *MemoryRateLimiter
	useRedis   bool
	checkMu    sync.RWMutex
	lastCheck  time.Time
	checkEvery time.Duration
}

func NewAutoFallbackRateLimiter(redisClient *redis.Client, rate, capacity int, window time.Duration) *AutoFallbackRateLimiter {
	afl := &AutoFallbackRateLimiter{
		memory:     NewMemoryRateLimiter(rate, capacity),
		checkEvery: 30 * time.Second,
		lastCheck:  time.Now(),
	}

	if redisClient != nil {
		afl.redis = NewRedisRateLimiter(redisClient, capacity, window)
		if err := redisClient.Ping(context.Background()).Err(); err == nil {
			afl.useRedis = true
			log.Info("Rate limiter using Redis backend")
		} else {
			log.Warn("Redis unavailable, rate limiter using memory backend", zap.Error(err))
		}
	} else {
		log.Info("Rate limiter using memory backend (Redis not configured)")
	}

	return afl
}

func (afl *AutoFallbackRateLimiter) Allow(key string) (bool, error) {
	afl.checkMu.RLock()
	useRedis := afl.useRedis
	afl.checkMu.RUnlock()

	if useRedis {
		allowed, err := afl.redis.Allow(key)
		if err != nil {
			afl.tryFallback(err)
			return afl.memory.Allow(key)
		}
		return allowed, nil
	}

	return afl.memory.Allow(key)
}

func (afl *AutoFallbackRateLimiter) tryFallback(err error) {
	afl.checkMu.Lock()
	defer afl.checkMu.Unlock()

	now := time.Now()
	if now.Sub(afl.lastCheck) < afl.checkEvery {
		return
	}
	afl.lastCheck = now

	if afl.useRedis {
		log.Warn("Redis rate limiter failed, falling back to memory", zap.Error(err))
		afl.useRedis = false
		go afl.tryReconnect()
	}
}

func (afl *AutoFallbackRateLimiter) tryReconnect() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		afl.checkMu.RLock()
		if afl.useRedis {
			afl.checkMu.RUnlock()
			return
		}
		afl.checkMu.RUnlock()

		if err := afl.redis.client.Ping(context.Background()).Err(); err == nil {
			afl.checkMu.Lock()
			afl.useRedis = true
			afl.checkMu.Unlock()
			log.Info("Redis reconnected, rate limiter switched back to Redis backend")
			return
		}
	}
}

// RateLimit 限流中间件
func RateLimit(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	if keyFunc == nil {
		keyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		key := keyFunc(c)
		allowed, err := limiter.Allow(key)

		if err != nil {
			log.Error("rate limit check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "rate limit check failed",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByIP IP 限流中间件
func RateLimitByIP(limiter RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// RateLimitByUser 用户限流中间件
func RateLimitByUser(limiter RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		if userID, exists := c.Get("claims"); exists {
			return fmt.Sprintf("user:%v", userID)
		}
		return c.ClientIP()
	})
}

// RateLimitByPath 路径限流中间件
func RateLimitByPath(limiter RateLimiter) gin.HandlerFunc {
	return RateLimit(limiter, func(c *gin.Context) string {
		return fmt.Sprintf("%s:%s", c.ClientIP(), c.Request.URL.Path)
	})
}
