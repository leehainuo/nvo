# 限流中间件使用指南

## 设计理念

限流中间件采用**依赖注入 + 自动降级**的优雅设计：

- ✅ **接口抽象**：`RateLimiter` 接口支持多种实现
- ✅ **依赖注入**：通过 Pocket 注入，无全局变量
- ✅ **自动降级**：Redis 不可用时自动切换到内存限流
- ✅ **自动恢复**：Redis 恢复后自动切回
- ✅ **灵活配置**：支持多种限流策略

## 架构设计

```
┌─────────────────────────────────────────┐
│         RateLimiter Interface           │
│         Allow(key) (bool, error)        │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
┌───────▼──────┐   ┌────────▼────────┐
│   Memory     │   │     Redis       │
│ RateLimiter  │   │  RateLimiter    │
└──────────────┘   └─────────────────┘
        │                   │
        └─────────┬─────────┘
                  │
        ┌─────────▼──────────┐
        │  AutoFallback      │
        │  RateLimiter       │
        │  - 自动降级         │
        │  - 自动恢复         │
        └────────────────────┘
```

## 配置说明

```yaml
# config.yml
ratelimit:
  rate: 10                    # 每秒生成令牌数（仅内存模式使用）
  capacity: 100               # 时间窗口内最大请求数
  window: 1m                  # 时间窗口（支持: s, m, h）
```

## 基本使用

### 1. 初始化 Pocket（自动创建限流器）

```go
package main

import (
    "nvo-api/core"
    "nvo-api/core/middleware"
)

func main() {
    // Pocket 会自动初始化限流器
    pocket := core.NewPocketBuilder("config.yml").
        WithEnforcer().
        MustBuild()
    defer pocket.Close()

    router := pocket.GinEngine
    
    // 使用限流器...
}
```

### 2. 全局 IP 限流

```go
// 对所有请求按 IP 限流
router.Use(middleware.RateLimitByIP(pocket.RateLimiter))
```

### 3. 路由组限流

```go
// 公开 API - 严格限流
public := router.Group("/api/v1")
public.Use(middleware.RateLimitByIP(pocket.RateLimiter))
{
    public.POST("/login", loginHandler)
    public.POST("/register", registerHandler)
}

// 认证 API - 用户级限流
auth := router.Group("/api/v1")
auth.Use(
    middleware.Jwt(pocket.JWT),
    middleware.RateLimitByUser(pocket.RateLimiter),
)
{
    auth.GET("/profile", profileHandler)
    auth.POST("/upload", uploadHandler)
}
```

### 4. 路径级限流

```go
// 对特定路径进行限流
router.POST("/api/v1/expensive-operation", 
    middleware.RateLimitByPath(pocket.RateLimiter),
    expensiveHandler,
)
```

## 高级用法

### 自定义 Key 生成函数

```go
// 按 API Key 限流
apiKeyLimiter := middleware.RateLimit(pocket.RateLimiter, func(c *gin.Context) string {
    apiKey := c.GetHeader("X-API-Key")
    if apiKey == "" {
        return c.ClientIP()
    }
    return fmt.Sprintf("apikey:%s", apiKey)
})

router.Use(apiKeyLimiter)
```

### 按租户限流

```go
// 多租户场景
tenantLimiter := middleware.RateLimit(pocket.RateLimiter, func(c *gin.Context) string {
    tenantID := c.GetHeader("X-Tenant-ID")
    if tenantID == "" {
        return c.ClientIP()
    }
    return fmt.Sprintf("tenant:%s", tenantID)
})
```

### 组合限流策略

```go
// 同时限制 IP 和用户
router.Use(
    middleware.RateLimitByIP(pocket.RateLimiter),    // IP 级限流
    middleware.Jwt(pocket.JWT),                       // 认证
    middleware.RateLimitByUser(pocket.RateLimiter),  // 用户级限流
)
```

## 三种限流器实现

### 1. MemoryRateLimiter（内存限流）

**特点**：
- 基于令牌桶算法
- 单机环境
- 自动清理过期记录
- 无外部依赖

**使用场景**：
- 开发环境
- 单机部署
- Redis 不可用时的降级方案

```go
// 手动创建内存限流器
limiter := middleware.NewMemoryRateLimiter(10, 100)
router.Use(middleware.RateLimitByIP(limiter))
```

### 2. RedisRateLimiter（Redis 限流）

**特点**：
- 基于滑动窗口算法
- 分布式环境
- 多实例共享限流状态
- 持久化

**使用场景**：
- 生产环境
- 分布式部署
- 需要精确限流

```go
// 手动创建 Redis 限流器
limiter := middleware.NewRedisRateLimiter(
    pocket.Redis,
    100,           // 容量
    time.Minute,   // 窗口
)
router.Use(middleware.RateLimitByIP(limiter))
```

### 3. AutoFallbackRateLimiter（自动降级）⭐ 推荐

**特点**：
- Redis 可用时使用 Redis
- Redis 不可用时自动降级到内存
- 后台自动检测 Redis 恢复
- 恢复后自动切回 Redis
- 对业务透明

**使用场景**：
- 所有场景（默认推荐）
- Pocket 自动创建此类型

```go
// Pocket 自动创建，直接使用
router.Use(middleware.RateLimitByIP(pocket.RateLimiter))
```

## 降级机制说明

### 自动降级流程

```
1. 启动时检测 Redis
   ├─ Redis 可用 → 使用 Redis 限流
   └─ Redis 不可用 → 使用内存限流

2. 运行时 Redis 故障
   ├─ 检测到错误 → 立即降级到内存
   └─ 后台定期检测 Redis（每 30 秒）

3. Redis 恢复
   ├─ 检测到连接成功 → 自动切回 Redis
   └─ 记录日志
```

### 日志输出示例

```
# 启动时 Redis 可用
INFO  Rate limiter using Redis backend

# 启动时 Redis 不可用
WARN  Redis unavailable, rate limiter using memory backend

# 运行时降级
WARN  Redis rate limiter failed, falling back to memory

# 自动恢复
INFO  Redis reconnected, rate limiter switched back to Redis backend
```

## 中间件分类

### 需要依赖注入的中间件（放入 Pocket）

```go
// 这些中间件需要外部依赖，通过 Pocket 注入
middleware.Jwt(pocket.JWT)                       // 需要 JWT 实例
middleware.Casbin(pocket.Enforcer)               // 需要 Enforcer 实例
middleware.RateLimitByIP(pocket.RateLimiter)     // 需要限流器实例
middleware.RateLimitByUser(pocket.RateLimiter)   // 需要限流器实例
```

### 无状态中间件（直接使用）

```go
// 这些中间件无外部依赖，直接调用
middleware.Cors()        // 跨域处理
middleware.Logger()      // 日志记录
middleware.Recovery()    // Panic 恢复
```

## 完整示例

```go
package main

import (
    "fmt"
    "nvo-api/core"
    "nvo-api/core/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    // 1. 初始化 Pocket
    pocket := core.NewPocketBuilder("config.yml").
        WithEnforcer().
        MustBuild()
    defer pocket.Close()

    router := pocket.GinEngine

    // 2. 全局中间件（无状态）
    router.Use(middleware.Cors())

    // 3. 公开路由 - IP 限流
    public := router.Group("/api/v1")
    public.Use(middleware.RateLimitByIP(pocket.RateLimiter))
    {
        public.POST("/login", loginHandler)
        public.POST("/register", registerHandler)
    }

    // 4. 认证路由 - 用户限流
    auth := router.Group("/api/v1")
    auth.Use(
        middleware.Jwt(pocket.JWT),
        middleware.RateLimitByUser(pocket.RateLimiter),
    )
    {
        auth.GET("/profile", profileHandler)
        auth.POST("/upload", uploadHandler)
    }

    // 5. 管理员路由 - 权限控制 + 限流
    admin := router.Group("/api/v1/admin")
    admin.Use(
        middleware.Jwt(pocket.JWT),
        middleware.Casbin(pocket.Enforcer),
        middleware.RateLimitByPath(pocket.RateLimiter),
    )
    {
        admin.GET("/users", listUsersHandler)
        admin.POST("/users", createUserHandler)
    }

    // 6. 特殊路由 - 自定义限流
    router.POST("/api/v1/webhook",
        middleware.RateLimit(pocket.RateLimiter, func(c *gin.Context) string {
            // 按 webhook 签名限流
            signature := c.GetHeader("X-Webhook-Signature")
            return fmt.Sprintf("webhook:%s", signature)
        }),
        webhookHandler,
    )

    router.Run(":8080")
}
```

## 性能优化

### 内存限流器优化

- 自动清理过期记录（每分钟）
- TTL 默认 5 分钟
- 读写锁优化并发性能

### Redis 限流器优化

- 使用 Pipeline 减少网络往返
- 自动设置过期时间
- 滑动窗口算法精确限流

## 测试建议

```go
// 测试限流器
func TestRateLimiter(t *testing.T) {
    limiter := middleware.NewMemoryRateLimiter(2, 5)
    
    // 前 5 次应该通过
    for i := 0; i < 5; i++ {
        allowed, _ := limiter.Allow("test-key")
        assert.True(t, allowed)
    }
    
    // 第 6 次应该被限流
    allowed, _ := limiter.Allow("test-key")
    assert.False(t, allowed)
}
```

## 最佳实践

1. **生产环境**：使用 `AutoFallbackRateLimiter`（Pocket 默认）
2. **分层限流**：全局 IP 限流 + 用户级限流
3. **合理配置**：根据业务调整 `capacity` 和 `window`
4. **监控日志**：关注降级和恢复日志
5. **错误处理**：限流失败返回明确错误码（1006/1007）

## 错误码说明

| 错误码 | 说明 | 场景 |
|--------|------|------|
| 1006 | 请求过于频繁 | 触发限流 |
| 1007 | 限流检查失败 | 系统错误 |

## 总结

这是一个**最优雅**的限流实现方案：

- ✅ 符合项目的依赖注入架构
- ✅ 自动降级保证高可用
- ✅ 接口抽象易于扩展
- ✅ 配置灵活满足多场景
- ✅ 对业务代码透明
