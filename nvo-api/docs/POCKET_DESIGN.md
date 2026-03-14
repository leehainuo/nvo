# Pocket 设计文档

## 概述

**Pocket（口袋）** 是一个优雅的依赖注入容器，负责管理应用的所有依赖，让业务模块通过统一的接口获取所需资源。

## 核心设计

### 1. Pocket 结构

```go
type Pocket struct {
    Config    *core.Config        // 应用配置
    DB        *gorm.DB            // 数据库连接
    Redis     *redis.Client       // Redis 客户端
    Logger    *zap.Logger         // 日志实例
    Enforcer  *casbin.Enforcer    // 权限控制（可选）
    GinEngine *gin.Engine         // Web 引擎
}
```

**特点：**
- ✅ 集中管理所有依赖
- ✅ 业务模块只需要 Pocket 一个参数
- ✅ 依赖关系清晰可见

### 2. Builder 模式初始化

```go
// 最简单的使用方式
pocket := core.NewPocketBuilder("config.yml").MustBuild()
defer pocket.Close()

// 带可选配置
pocket := core.NewPocketBuilder("config.yml").
    WithEnforcer("model.conf", "policy.csv").
    SkipRedis().  // 不需要 Redis
    MustBuild()
```

**优势：**
- ✅ 一行代码完成初始化
- ✅ 链式调用，可读性强
- ✅ 支持可选依赖
- ✅ 自动处理初始化顺序

### 3. 初始化流程

```
NewPocketBuilder("config.yml")
    ↓
1. loadConfig()      // 加载配置，获取 viper 实例
    ↓
2. initLogger()      // 初始化日志（依赖配置）
    ↓
3. initDB()          // 初始化数据库（依赖日志）
    ↓
4. initRedis()       // 初始化 Redis（依赖日志）
    ↓
5. initGin()         // 初始化 Gin（依赖配置）
    ↓
Build() / MustBuild()
```

**关键点：**
- 依赖按顺序初始化
- 每一步都有错误处理
- Logger 最先初始化，用于记录后续步骤

### 4. 资源清理

```go
// Close 方法自动清理所有资源
func (p *Pocket) Close() error {
    // 1. 关闭 Redis
    // 2. 关闭数据库
    // 3. 同步日志
    // 4. 收集并返回错误
}

// 使用方式
defer pocket.Close()  // 一行搞定！
```

**优势：**
- ✅ 统一的资源清理接口
- ✅ 自动处理 nil 检查
- ✅ 收集所有错误
- ✅ 不会遗漏资源

## 使用示例

### 基础使用

```go
package main

import (
    core "nvo-api/core/pocket"
    "nvo-api/internal/system/user"
)

func main() {
    // 1. 初始化 Pocket
    pocket := core.NewPocketBuilder("config.yml").MustBuild()
    defer pocket.Close()

    // 2. 注册业务模块
    userModule := user.NewModule(pocket)
    api := pocket.GinEngine.Group("/api/v1")
    userModule.RegisterRoutes(api)

    // 3. 启动服务
    addr := fmt.Sprintf("%s:%d", 
        pocket.Config.Server.Host, 
        pocket.Config.Server.Port)
    pocket.Logger.Info("Server starting", zap.String("addr", addr))
    
    if err := pocket.GinEngine.Run(addr); err != nil {
        pocket.Logger.Fatal("Failed to start server", zap.Error(err))
    }
}
```

### 业务模块使用

```go
// internal/system/user/module.go
package user

import core "nvo-api/core/pocket"

type Module struct {
    pocket *core.Pocket
}

func NewModule(pocket *core.Pocket) *Module {
    // 从 Pocket 获取依赖
    userRepo := repository.NewUserRepository(pocket.DB)
    userService := service.NewUserService(userRepo)
    userHandler := api.NewUserHandler(userService)

    // 使用 Logger 记录日志
    pocket.Logger.Info("User module initialized")

    // 自动迁移表结构
    pocket.DB.AutoMigrate(&domain.User{})

    return &Module{pocket: pocket}
}
```

### 可选依赖

```go
// 不需要 Redis 的服务
pocket := core.NewPocketBuilder("config.yml").
    SkipRedis().
    MustBuild()

// 不需要数据库的服务（如纯 API 网关）
pocket := core.NewPocketBuilder("config.yml").
    SkipDB().
    SkipRedis().
    MustBuild()

// 需要权限控制
pocket := core.NewPocketBuilder("config.yml").
    WithEnforcer("model.conf", "policy.csv").
    MustBuild()
```

## 设计优势

### 1. 完全解耦

```
业务模块 → Pocket → 各种依赖
```

- 业务模块只依赖 Pocket 接口
- 不需要知道依赖如何初始化
- 依赖变更不影响业务代码

### 2. 配置解耦

```
core/config → viper 实例 → pkg/client/*
```

- `core/config` 不依赖客户端包
- 客户端包自己读取配置
- 完全的依赖倒置

### 3. 易于测试

```go
// 测试时可以 Mock Pocket
type MockPocket struct {
    DB     *gorm.DB
    Logger *zap.Logger
}

func TestUserModule(t *testing.T) {
    mockDB := setupTestDB()
    mockLogger := zap.NewNop()
    
    pocket := &core.Pocket{
        DB:     mockDB,
        Logger: mockLogger,
    }
    
    module := user.NewModule(pocket)
    // 测试...
}
```

### 4. 生命周期管理

| 阶段 | 操作 | 说明 |
|------|------|------|
| **创建** | `NewPocketBuilder()` | 创建构建器 |
| **配置** | `.WithXXX()` | 可选配置 |
| **初始化** | `.Build()` / `.MustBuild()` | 执行初始化 |
| **使用** | `pocket.DB`, `pocket.Logger` | 获取依赖 |
| **清理** | `pocket.Close()` | 释放资源 |

## 对比其他框架

### vs go-zero ServiceContext

| 特性 | go-zero | Pocket |
|------|---------|--------|
| 初始化方式 | 手动组装 | Builder 模式 |
| 配置管理 | 集中式 | 分散式（解耦） |
| 资源清理 | 手动 | 自动 `Close()` |
| 可选依赖 | 不支持 | 支持 `Skip*()` |
| 链式调用 | 不支持 | 支持 |

### vs Spring IoC

| 特性 | Spring | Pocket |
|------|--------|--------|
| 依赖注入 | 注解/XML | 构造函数 |
| 复杂度 | 高（反射） | 低（显式） |
| 类型安全 | 弱 | 强 |
| 学习曲线 | 陡峭 | 平缓 |

## 最佳实践

### 1. 业务模块设计

```go
// ✅ 推荐：通过构造函数注入 Pocket
func NewModule(pocket *core.Pocket) *Module {
    return &Module{pocket: pocket}
}

// ❌ 不推荐：直接依赖具体实现
func NewModule(db *gorm.DB, logger *zap.Logger) *Module {
    // 参数太多，难以维护
}
```

### 2. 依赖获取

```go
// ✅ 推荐：在初始化时获取依赖
func NewModule(pocket *core.Pocket) *Module {
    repo := repository.New(pocket.DB)
    service := service.New(repo)
    return &Module{service: service}
}

// ❌ 不推荐：每次使用时获取
func (m *Module) Handle() {
    repo := repository.New(m.pocket.DB)  // 重复创建
}
```

### 3. 错误处理

```go
// ✅ 推荐：使用 MustBuild() 快速失败
pocket := core.NewPocketBuilder("config.yml").MustBuild()

// ✅ 推荐：需要自定义错误处理时使用 Build()
pocket, err := core.NewPocketBuilder("config.yml").Build()
if err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}
```

### 4. 资源清理

```go
// ✅ 推荐：使用 defer 确保清理
func main() {
    pocket := core.NewPocketBuilder("config.yml").MustBuild()
    defer pocket.Close()
    // ...
}

// ❌ 不推荐：手动清理（容易遗漏）
func main() {
    pocket := core.NewPocketBuilder("config.yml").MustBuild()
    // ... 业务逻辑
    pocket.DB.Close()  // 可能因为 panic 而不执行
}
```

## 扩展 Pocket

### 添加新依赖

```go
// 1. 在 Pocket 中添加字段
type Pocket struct {
    // ...
    MongoDB *mongo.Client  // 新增
}

// 2. 在 Builder 中添加初始化方法
func (b *PocketBuilder) initMongoDB() error {
    client, err := mongodb.InitFromViper(b.viper, "mongodb")
    if err != nil {
        return err
    }
    b.pocket.MongoDB = client
    return nil
}

// 3. 在 Build() 中调用
func (b *PocketBuilder) Build() (*Pocket, error) {
    // ...
    if err := b.initMongoDB(); err != nil {
        return nil, err
    }
    // ...
}

// 4. 在 Close() 中清理
func (p *Pocket) Close() error {
    // ...
    if p.MongoDB != nil {
        p.MongoDB.Disconnect(context.Background())
    }
    // ...
}
```

## 总结

Pocket 的设计哲学：

1. **简单优于复杂** - 一行代码完成初始化
2. **显式优于隐式** - 依赖关系清晰可见
3. **解耦优于耦合** - 配置和依赖完全分离
4. **安全优于便利** - 强类型，编译时检查
5. **优雅优于粗糙** - Builder 模式，链式调用

**这是一个优雅、简洁、易用的依赖注入容器！** ✨
