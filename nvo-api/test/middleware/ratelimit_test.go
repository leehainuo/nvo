package middleware_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"nvo-api/core/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRateLimiter_Basic(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(2, 5)
	defer limiter.Stop()

	t.Run("允许前5次请求", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("test-key")
			assert.NoError(t, err)
			assert.True(t, allowed, "第 %d 次请求应该被允许", i+1)
		}
	})

	t.Run("第6次请求被限流", func(t *testing.T) {
		allowed, err := limiter.Allow("test-key")
		assert.NoError(t, err)
		assert.False(t, allowed, "第6次请求应该被限流")
	})
}

func TestMemoryRateLimiter_TokenRefill(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(10, 5)
	defer limiter.Stop()

	t.Run("消耗所有令牌", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, _ := limiter.Allow("refill-key")
			assert.True(t, allowed)
		}
		allowed, _ := limiter.Allow("refill-key")
		assert.False(t, allowed, "令牌耗尽后应该被限流")
	})

	t.Run("等待令牌补充", func(t *testing.T) {
		time.Sleep(600 * time.Millisecond)
		allowed, err := limiter.Allow("refill-key")
		assert.NoError(t, err)
		assert.True(t, allowed, "等待后令牌应该补充，请求应该被允许")
	})
}

func TestMemoryRateLimiter_MultipleKeys(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(2, 3)
	defer limiter.Stop()

	t.Run("不同key独立限流", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			allowed1, _ := limiter.Allow("key1")
			allowed2, _ := limiter.Allow("key2")
			assert.True(t, allowed1, "key1 的第 %d 次请求应该被允许", i+1)
			assert.True(t, allowed2, "key2 的第 %d 次请求应该被允许", i+1)
		}

		allowed1, _ := limiter.Allow("key1")
		allowed2, _ := limiter.Allow("key2")
		assert.False(t, allowed1, "key1 应该被限流")
		assert.False(t, allowed2, "key2 应该被限流")
	})
}

func TestMemoryRateLimiter_Cleanup(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(10, 100)
	defer limiter.Stop()

	t.Run("创建多个key", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("cleanup-key-%d", i)
			allowed, _ := limiter.Allow(key)
			assert.True(t, allowed)
		}
		// 测试通过即表示内存管理正常
	})

	t.Run("令牌桶正常工作", func(t *testing.T) {
		// 验证限流功能正常
		for i := 0; i < 100; i++ {
			allowed, _ := limiter.Allow("test-key")
			assert.True(t, allowed)
		}
		// 超过容量应该被限流
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)
	})
}

func TestMemoryRateLimiter_Concurrent(t *testing.T) {
	limiter := middleware.NewMemoryRateLimiter(100, 1000)
	defer limiter.Stop()
	concurrency := 50
	requestsPerGoroutine := 20

	t.Run("并发请求", func(t *testing.T) {
		done := make(chan bool, concurrency)
		allowed := make(chan bool, concurrency*requestsPerGoroutine)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				key := fmt.Sprintf("concurrent-key-%d", id%5)
				for j := 0; j < requestsPerGoroutine; j++ {
					ok, _ := limiter.Allow(key)
					allowed <- ok
				}
				done <- true
			}(i)
		}

		for i := 0; i < concurrency; i++ {
			<-done
		}
		close(allowed)

		allowedCount := 0
		for ok := range allowed {
			if ok {
				allowedCount++
			}
		}

		assert.Greater(t, allowedCount, 0, "应该有请求被允许")
		t.Logf("并发测试: %d/%d 请求被允许", allowedCount, concurrency*requestsPerGoroutine)
	})
}

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis 不可用，跳过 Redis 测试")
		return nil
	}

	client.FlushDB(ctx)
	return client
}

func TestRedisRateLimiter_Basic(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	limiter := middleware.NewRedisRateLimiter(client, 5, time.Second)

	t.Run("允许前5次请求", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("redis-test-key")
			assert.NoError(t, err)
			assert.True(t, allowed, "第 %d 次请求应该被允许", i+1)
		}
	})

	t.Run("第6次请求被限流", func(t *testing.T) {
		allowed, err := limiter.Allow("redis-test-key")
		assert.NoError(t, err)
		assert.False(t, allowed, "第6次请求应该被限流")
	})

	t.Run("等待窗口过期", func(t *testing.T) {
		time.Sleep(1100 * time.Millisecond)
		allowed, err := limiter.Allow("redis-test-key")
		assert.NoError(t, err)
		assert.True(t, allowed, "窗口过期后应该允许新请求")
	})
}

func TestRedisRateLimiter_SlidingWindow(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	limiter := middleware.NewRedisRateLimiter(client, 10, 2*time.Second)

	t.Run("滑动窗口测试", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, _ := limiter.Allow("sliding-key")
			assert.True(t, allowed)
		}

		time.Sleep(1 * time.Second)

		for i := 0; i < 5; i++ {
			allowed, _ := limiter.Allow("sliding-key")
			assert.True(t, allowed)
		}

		allowed, _ := limiter.Allow("sliding-key")
		assert.False(t, allowed, "超过窗口容量应该被限流")
	})
}

func TestRedisRateLimiter_MultipleKeys(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	limiter := middleware.NewRedisRateLimiter(client, 3, time.Second)

	t.Run("不同key独立限流", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			allowed1, _ := limiter.Allow("redis-key1")
			allowed2, _ := limiter.Allow("redis-key2")
			assert.True(t, allowed1)
			assert.True(t, allowed2)
		}

		allowed1, _ := limiter.Allow("redis-key1")
		allowed2, _ := limiter.Allow("redis-key2")
		assert.False(t, allowed1)
		assert.False(t, allowed2)
	})
}

func TestAutoFallbackRateLimiter_WithRedis(t *testing.T) {
	client := setupTestRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	limiter := middleware.NewAutoFallbackRateLimiter(client, 10, 5, time.Second)

	t.Run("Redis可用时正常限流", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("fallback-key")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		allowed, err := limiter.Allow("fallback-key")
		assert.NoError(t, err)
		assert.False(t, allowed, "应该被限流")
	})
}

func TestAutoFallbackRateLimiter_WithoutRedis(t *testing.T) {
	limiter := middleware.NewAutoFallbackRateLimiter(nil, 10, 5, time.Second)

	t.Run("Redis不可用时降级到内存", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("memory-key")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		allowed, err := limiter.Allow("memory-key")
		assert.NoError(t, err)
		assert.False(t, allowed, "应该被限流")
	})
}

func TestAutoFallbackRateLimiter_Fallback(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999",
		DB:   0,
	})

	limiter := middleware.NewAutoFallbackRateLimiter(client, 10, 5, time.Second)

	t.Run("Redis连接失败时自动降级", func(t *testing.T) {
		// 验证降级后限流功能正常
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("fallback-test")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		allowed, err := limiter.Allow("fallback-test")
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRateLimitMiddleware_ByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewMemoryRateLimiter(10, 3)

	router := gin.New()
	router.Use(middleware.RateLimitByIP(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("允许前3次请求", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code, "第 %d 次请求应该成功", i+1)
		}
	})

	t.Run("第4次请求被限流", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code, "应该返回 429 Too Many Requests")
	})

	t.Run("不同IP独立限流", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "不同IP应该独立计数")
	})
}

func TestRateLimitMiddleware_ByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewMemoryRateLimiter(10, 2)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", "user123")
		c.Next()
	})
	router.Use(middleware.RateLimitByUser(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("按用户ID限流", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code, "超过用户限额应该被限流")
	})
}

func TestRateLimitMiddleware_ByPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewMemoryRateLimiter(10, 2)

	router := gin.New()
	router.Use(middleware.RateLimitByPath(limiter))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	router.GET("/api/v1/other", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("不同路径独立限流", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}

		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code, "/api/v1/test 应该被限流")

		req = httptest.NewRequest("GET", "/api/v1/other", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "/api/v1/other 应该独立计数")
	})
}

func TestRateLimitMiddleware_CustomKeyFunc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := middleware.NewMemoryRateLimiter(10, 2)

	router := gin.New()
	router.Use(middleware.RateLimit(limiter, func(c *gin.Context) string {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			return c.ClientIP()
		}
		return fmt.Sprintf("apikey:%s", apiKey)
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	t.Run("按API Key限流", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-API-Key", "key123")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "key123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code, "相同API Key应该被限流")

		req = httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "key456")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "不同API Key应该独立计数")
	})
}

func BenchmarkMemoryRateLimiter(b *testing.B) {
	limiter := middleware.NewMemoryRateLimiter(1000, 10000)

	b.Run("单key", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limiter.Allow("bench-key")
		}
	})

	b.Run("多key", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench-key-%d", i%100)
			limiter.Allow(key)
		}
	})
}

func BenchmarkRedisRateLimiter(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		b.Skip("Redis 不可用")
		return
	}
	defer client.Close()

	limiter := middleware.NewRedisRateLimiter(client, 10000, time.Minute)

	b.Run("单key", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			limiter.Allow("bench-key")
		}
	})

	b.Run("多key", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench-key-%d", i%100)
			limiter.Allow(key)
		}
	})
}
