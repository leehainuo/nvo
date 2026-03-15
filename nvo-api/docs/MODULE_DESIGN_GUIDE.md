# 模块化架构设计完整教程

> 从零开始构建优雅的 Go 模块化架构：Pocket 依赖注入 + 模块化设计

## 📚 目录

1. [架构概述](#1-架构概述)
2. [核心概念](#2-核心概念)
3. [Pocket 依赖注入容器](#3-pocket-依赖注入容器)
4. [模块化设计](#4-模块化设计)
5. [完整实战示例](#5-完整实战示例)
6. [最佳实践](#6-最佳实践)
7. [常见问题](#7-常见问题)

---

## 1. 架构概述

### 1.1 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                      Application                         │
│                       (main.go)                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────────┐
│                    core.Pocket                           │
│              (依赖注入容器)                               │
│  ┌──────────┬──────────┬──────────┬──────────┐         │
│  │ Database │  Redis   │ Enforcer │   JWT    │         │
│  └──────────┴──────────┴──────────┴──────────┘         │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ↓            ↓            ↓
   ┌────────┐  ┌────────┐  ┌────────┐
   │  User  │  │  Role  │  │  Perm  │
   │ Module │  │ Module │  │ Module │
   └────────┘  └────────┘  └────────┘
   
   每个模块内部：
   ┌─────────────────────┐
   │  Module (入口)       │
   ├─────────────────────┤
   │  API (HTTP 层)      │
   ├─────────────────────┤
   │  Service (业务层)    │
   ├─────────────────────┤
   │  Repository (数据层) │
   ├─────────────────────┤
   │  Domain (领域模型)   │
   └─────────────────────┘
```

### 1.2 设计理念

**核心原则**：
- ✅ **高内聚低耦合**：模块独立，通过共享资源交互
- ✅ **依赖注入**：通过 Pocket 统一管理依赖
- ✅ **分层架构**：清晰的职责划分
- ✅ **可插拔**：模块可以独立启用/禁用

**优势**：
- 🚀 易于测试：模块可以独立测试
- 🔧 易于维护：修改一个模块不影响其他模块
- 📦 易于扩展：新增模块只需实现接口
- 🎯 职责清晰：每层有明确的职责

---

## 2. 核心概念

### 2.1 Pocket（依赖注入容器）

**什么是 Pocket？**

Pocket 是一个依赖注入容器，管理应用程序的所有共享资源（数据库、缓存、配置等）。

**为什么叫 Pocket（口袋）？**
- 像口袋一样，装着应用程序需要的所有"工具"
- 需要什么就从口袋里拿什么

**Pocket 的职责**：
1. 初始化所有基础设施（DB、Redis、Casbin 等）
2. 提供统一的资源访问接口
3. 管理资源的生命周期

### 2.2 Module（模块）

**什么是 Module？**

Module 是一个独立的业务单元，包含完整的业务逻辑、数据访问和 HTTP 接口。

**Module 的特点**：
- 🎯 **自包含**：包含自己的 domain、repository、service、api
- 🔌 **可插拔**：可以独立启用/禁用
- 🚫 **无依赖**：不直接依赖其他模块
- 📦 **标准化**：实现统一的 Module 接口

### 2.3 分层架构

```
┌─────────────────────────────────────────┐
│  API Layer (HTTP 处理)                   │
│  - 参数验证                              │
│  - 请求响应                              │
│  - 错误处理                              │
└──────────────┬──────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│  Service Layer (业务逻辑)                │
│  - 业务规则                              │
│  - 事务管理                              │
│  - 权限控制                              │
└──────────────┬──────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│  Repository Layer (数据访问)             │
│  - CRUD 操作                            │
│  - 查询构建                              │
│  - 数据映射                              │
└──────────────┬──────────────────────────┘
               ↓
┌─────────────────────────────────────────┐
│  Domain Layer (领域模型)                 │
│  - 数据模型                              │
│  - DTO 定义                             │
│  - 业务实体                              │
└─────────────────────────────────────────┘
```

---

## 3. Pocket 依赖注入容器

### 3.1 创建 Pocket 结构体

**步骤 1：定义 Pocket 结构**

```go
// core/pocket.go
package core

import (
    "github.com/casbin/casbin/v3"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// Pocket 依赖注入容器
// 管理所有应用级依赖，业务模块通过 Pocket 获取所需资源
type Pocket struct {
    Config    *config.Config           // 配置
    DB        *gorm.DB                 // 数据库
    Redis     *redis.Client            // 缓存
    JWT       *jwt.JWT                 // JWT 工具
    Enforcer  *casbin.SyncedEnforcer   // 权限控制
    GinEngine *gin.Engine              // Web 框架
}
```

**为什么这样设计？**
- ✅ 所有依赖集中管理
- ✅ 模块通过 Pocket 获取资源
- ✅ 易于测试（Mock Pocket）

### 3.2 创建 PocketBuilder（构建器模式）

**步骤 2：实现构建器**

```go
// PocketBuilder 构建器模式，用于初始化 Pocket
type PocketBuilder struct {
    configPath string
    config     *config.Config
    pocket     *Pocket
}

// NewPocketBuilder 创建 Pocket 构建器
func NewPocketBuilder(configPath string) *PocketBuilder {
    return &PocketBuilder{
        configPath: configPath,
        pocket:     &Pocket{},
    }
}

// Build 构建并初始化所有依赖
func (b *PocketBuilder) Build() (*Pocket, error) {
    // 1. 加载配置
    if err := b.loadConfig(); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    // 2. 初始化数据库
    if err := b.initDB(); err != nil {
        return nil, fmt.Errorf("failed to init database: %w", err)
    }

    // 3. 初始化 Redis
    if err := b.initRedis(); err != nil {
        log.Warn("Redis init failed, will use memory fallback", zap.Error(err))
    }

    // 4. 初始化 JWT
    if err := b.initJWT(); err != nil {
        return nil, fmt.Errorf("failed to init JWT: %w", err)
    }

    // 5. 初始化权限控制
    if err := b.initEnforcer(); err != nil {
        return nil, fmt.Errorf("failed to init enforcer: %w", err)
    }

    // 6. 初始化 Gin 引擎
    b.initGin()

    log.Info("Pocket initialized successfully")
    return b.pocket, nil
}

// MustBuild 构建 Pocket，失败则 panic
func (b *PocketBuilder) MustBuild() *Pocket {
    pocket, err := b.Build()
    if err != nil {
        panic(fmt.Sprintf("Failed to build pocket: %v", err))
    }
    return pocket
}
```

**为什么使用构建器模式？**
- ✅ 初始化步骤清晰
- ✅ 支持链式调用
- ✅ 易于扩展（可以添加 Skip 选项）

### 3.3 实现各个初始化方法

**步骤 3：初始化数据库**

```go
// initDB 初始化数据库
func (b *PocketBuilder) initDB() error {
    db, err := gorm.Open(mysql.Open(b.config.Database.DSN), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return err
    }
    
    b.pocket.DB = db
    log.Info("Database connected successfully")
    return nil
}
```

**步骤 4：初始化 Redis**

```go
// initRedis 初始化 Redis
func (b *PocketBuilder) initRedis() error {
    rdb := redis.NewClient(&redis.Options{
        Addr:     b.config.Redis.Addr,
        Password: b.config.Redis.Password,
        DB:       b.config.Redis.DB,
    })
    
    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := rdb.Ping(ctx).Err(); err != nil {
        return err
    }
    
    b.pocket.Redis = rdb
    log.Info("Redis connected successfully")
    return nil
}
```

**步骤 5：初始化 Casbin**

```go
// initEnforcer 初始化 Casbin 权限控制
func (b *PocketBuilder) initEnforcer() error {
    adapter, err := gormadapter.NewAdapterByDB(b.pocket.DB)
    if err != nil {
        return err
    }
    
    enforcer, err := casbin.NewSyncedEnforcer(b.config.Casbin.ModelPath, adapter)
    if err != nil {
        return err
    }
    
    // 加载策略
    if err := enforcer.LoadPolicy(); err != nil {
        return err
    }
    
    b.pocket.Enforcer = enforcer
    log.Info("Casbin enforcer initialized successfully")
    return nil
}
```

### 3.4 资源清理

**步骤 6：实现 Close 方法**

```go
// Close 优雅关闭所有资源
func (p *Pocket) Close() error {
    var errs []error

    // 关闭 Redis
    if p.Redis != nil {
        if err := p.Redis.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close redis: %w", err))
        }
    }

    // 关闭数据库
    if p.DB != nil {
        if sqlDB, err := p.DB.DB(); err == nil {
            if err := sqlDB.Close(); err != nil {
                errs = append(errs, fmt.Errorf("failed to close database: %w", err))
            }
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors during close: %v", errs)
    }
    return nil
}
```

---

## 4. 模块化设计

### 4.1 定义 Module 接口

**步骤 1：创建模块接口**

```go
// internal/module.go
package internal

import "github.com/gin-gonic/gin"

// Module 模块接口
// 所有业务模块都应该实现此接口，以实现模块化、可插拔的架构设计
type Module interface {
    Name() string                      // 返回模块名称
    Models() []any                     // 返回需要迁移的数据模型
    RegisterRoutes(r *gin.RouterGroup) // 注册模块路由
}
```

**接口说明**：
- `Name()`：模块标识，用于日志和调试
- `Models()`：返回需要数据库迁移的模型
- `RegisterRoutes()`：注册 HTTP 路由

### 4.2 创建模块目录结构

**步骤 2：创建标准目录结构**

```bash
internal/system/user/
├── domain/              # 领域模型
│   └── user.go         # User 实体 + DTO
├── repository/          # 数据访问层
│   └── user_repository.go
├── service/             # 业务逻辑层
│   └── user_service.go
├── api/                 # HTTP 处理层
│   └── user_handler.go
└── module.go            # 模块入口
```

**为什么这样分层？**
- ✅ **Domain**：定义数据结构，不依赖任何层
- ✅ **Repository**：只负责数据访问，不包含业务逻辑
- ✅ **Service**：实现业务规则，调用 Repository
- ✅ **API**：处理 HTTP 请求，调用 Service

### 4.3 实现 Domain 层

**步骤 3：定义领域模型**

```go
// internal/system/user/domain/user.go
package domain

import (
    "time"
    "gorm.io/gorm"
)

// User 用户领域模型
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    Nickname  string         `gorm:"size:50" json:"nickname"`
    Email     string         `gorm:"uniqueIndex;size:100" json:"email"`
    Status    int8           `gorm:"default:1" json:"status"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
    return "sys_users"
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6"`
    Nickname string `json:"nickname" binding:"max=50"`
    Email    string `json:"email" binding:"omitempty,email"`
}

// UserResponse 用户响应
type UserResponse struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Nickname  string    `json:"nickname"`
    Email     string    `json:"email"`
    Status    int8      `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

**设计要点**：
- ✅ 实体定义（User）
- ✅ 请求 DTO（CreateUserRequest）
- ✅ 响应 DTO（UserResponse）
- ✅ 表名映射（TableName）

### 4.4 实现 Repository 层

**步骤 4：实现数据访问**

```go
// internal/system/user/repository/user_repository.go
package repository

import (
    "nvo-api/internal/system/user/domain"
    "gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *domain.User) error {
    return r.db.Create(user).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id uint) (*domain.User, error) {
    var user domain.User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
    var user domain.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *domain.User) error {
    return r.db.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(id uint) error {
    return r.db.Delete(&domain.User{}, id).Error
}

// List 获取用户列表
func (r *UserRepository) List(page, pageSize int) ([]*domain.User, int64, error) {
    var users []*domain.User
    var total int64

    query := r.db.Model(&domain.User{})
    
    // 统计总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    offset := (page - 1) * pageSize
    if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}
```

**Repository 职责**：
- ✅ 只负责数据库操作
- ✅ 不包含业务逻辑
- ✅ 返回领域模型

### 4.5 实现 Service 层

**步骤 5：实现业务逻辑**

```go
// internal/system/user/service/user_service.go
package service

import (
    "errors"
    "fmt"

    "nvo-api/core/log"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/repository"

    "github.com/casbin/casbin/v3"
    "go.uber.org/zap"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

// UserService 用户业务逻辑层
type UserService struct {
    repo     *repository.UserRepository
    enforcer *casbin.SyncedEnforcer
    db       *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, enforcer *casbin.SyncedEnforcer) *UserService {
    return &UserService{
        repo:     repository.NewUserRepository(db),
        enforcer: enforcer,
        db:       db,
    }
}

// Create 创建用户
func (s *UserService) Create(req *domain.CreateUserRequest) (*domain.User, error) {
    // 1. 检查用户名是否存在
    existingUser, _ := s.repo.GetByUsername(req.Username)
    if existingUser != nil {
        return nil, errors.New("用户名已存在")
    }

    // 2. 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("密码加密失败: %w", err)
    }

    // 3. 创建用户
    user := &domain.User{
        Username: req.Username,
        Password: string(hashedPassword),
        Nickname: req.Nickname,
        Email:    req.Email,
        Status:   1,
    }

    // 4. 保存到数据库
    if err := s.repo.Create(user); err != nil {
        return nil, err
    }

    log.Info("user created", zap.String("username", user.Username), zap.Uint("id", user.ID))
    return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id uint) (*domain.UserResponse, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("用户不存在")
        }
        return nil, err
    }

    return &domain.UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Nickname:  user.Nickname,
        Email:     user.Email,
        Status:    user.Status,
        CreatedAt: user.CreatedAt,
    }, nil
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
    // 开启事务
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 删除用户
        if err := s.repo.Delete(id); err != nil {
            return err
        }

        // 2. 删除用户的权限关联
        subject := fmt.Sprintf("user:%d", id)
        if _, err := s.enforcer.DeleteUser(subject); err != nil {
            return err
        }

        return nil
    })
}
```

**Service 职责**：
- ✅ 实现业务规则
- ✅ 事务管理
- ✅ 权限控制
- ✅ 调用 Repository

### 4.6 实现 API 层

**步骤 6：实现 HTTP 处理**

```go
// internal/system/user/api/user_handler.go
package api

import (
    "errors"
    "strconv"

    "nvo-api/core/log"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/service"
    "nvo-api/pkg/response"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
    service *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(service *service.UserService) *UserHandler {
    return &UserHandler{
        service: service,
    }
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
    var req domain.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.New("参数错误: "+err.Error()))
        return
    }

    user, err := h.service.Create(&req)
    if err != nil {
        log.Error("create user failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, user)
}

// GetByID 根据 ID 获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        response.Error(c, errors.New("无效的用户ID"))
        return
    }

    user, err := h.service.GetByID(uint(id))
    if err != nil {
        log.Error("get user failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, user)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        response.Error(c, errors.New("无效的用户ID"))
        return
    }

    if err := h.service.Delete(uint(id)); err != nil {
        log.Error("delete user failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, nil)
}
```

**API 职责**：
- ✅ 参数验证
- ✅ 请求响应
- ✅ 错误处理
- ✅ 调用 Service

### 4.7 实现 Module 入口

**步骤 7：创建模块入口**

```go
// internal/system/user/module.go
package user

import (
    "nvo-api/core"
    "nvo-api/internal/system/user/api"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/service"

    "github.com/gin-gonic/gin"
)

// Module 用户模块
type Module struct {
    pocket  *core.Pocket
    handler *api.UserHandler
}

// NewModule 创建用户模块
func NewModule(pocket *core.Pocket) *Module {
    // 初始化服务
    userService := service.NewUserService(pocket.DB, pocket.Enforcer)

    // 初始化处理器
    userHandler := api.NewUserHandler(userService)

    return &Module{
        pocket:  pocket,
        handler: userHandler,
    }
}

// Name 模块名称
func (m *Module) Name() string {
    return "user"
}

// Models 返回需要迁移的数据模型
func (m *Module) Models() []any {
    return []any{
        &domain.User{},
    }
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.POST("", m.handler.Create)
        users.GET("/:id", m.handler.GetByID)
        users.DELETE("/:id", m.handler.Delete)
    }
}
```

**Module 职责**：
- ✅ 模块初始化
- ✅ 依赖组装
- ✅ 路由注册
- ✅ 模型声明

### 4.8 创建模块注册器

**步骤 8：统一注册所有模块**

```go
// internal/system/registry.go
package system

import (
    "nvo-api/core"
    "nvo-api/core/log"
    "nvo-api/internal"
    "nvo-api/internal/system/user"
    "nvo-api/internal/system/role"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// RegisterModules 注册所有系统模块
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 1. 初始化所有模块
    modules := []internal.Module{
        user.NewModule(p),
        role.NewModule(p),
    }

    // 2. 收集并迁移所有模型
    if err := migrateModels(p.DB, modules); err != nil {
        log.Fatal("Database migration failed", zap.Error(err))
    }

    // 3. 注册路由
    for _, module := range modules {
        log.Info("Registering system module", zap.String("name", module.Name()))
        module.RegisterRoutes(r)
    }
}

// migrateModels 收集并迁移所有模块的数据模型
func migrateModels(db *gorm.DB, modules []internal.Module) error {
    var allModels []any

    // 收集所有模型
    for _, module := range modules {
        models := module.Models()
        if len(models) > 0 {
            log.Info("Collecting models from module",
                zap.String("module", module.Name()),
                zap.Int("count", len(models)))
            allModels = append(allModels, models...)
        }
    }

    // 统一迁移
    if len(allModels) > 0 {
        log.Info("Starting database migration", zap.Int("total_models", len(allModels)))
        if err := db.AutoMigrate(allModels...); err != nil {
            return err
        }
        log.Info("Database migration completed successfully")
    }

    return nil
}
```

**Registry 职责**：
- ✅ 模块注册
- ✅ 数据库迁移
- ✅ 路由注册
- ✅ 启动日志

---

## 5. 完整实战示例

### 5.1 主程序入口

**步骤 9：编写 main.go**

```go
// cmd/main.go
package main

import (
    "fmt"

    "nvo-api/core"
    "nvo-api/core/log"
    "nvo-api/internal/system"
)

func main() {
    // 1. 构建 Pocket（依赖注入容器）
    pocket := core.NewPocketBuilder("config.yml").MustBuild()
    defer pocket.Close()

    log.Info("Pocket initialized successfully")

    // 2. 注册系统模块
    api := pocket.GinEngine.Group("/api/v1")
    system.RegisterModules(api, pocket)

    log.Info("All modules registered successfully")

    // 3. 启动服务器
    addr := fmt.Sprintf(":%d", pocket.Config.Server.Port)
    log.Info("Starting server", zap.String("addr", addr))
    
    if err := pocket.GinEngine.Run(addr); err != nil {
        log.Fatal("Failed to start server", zap.Error(err))
    }
}
```

### 5.2 配置文件

**步骤 10：创建配置文件**

```yaml
# config.yml
server:
  host: 0.0.0.0
  port: 8080

database:
  driver: mysql
  host: localhost
  port: 3306
  database: nvo_api
  username: root
  password: password
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100

redis:
  addr: localhost:6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key
  expire: 7200

casbin:
  model_path: ./config/casbin_model.conf

log:
  level: info
  output: stdout
```

### 5.3 启动流程

**步骤 11：运行应用**

```bash
# 1. 安装依赖
go mod tidy

# 2. 运行应用
go run cmd/main.go

# 输出：
# INFO  Pocket initialized successfully
# INFO  Collecting models from module  module=user count=1
# INFO  Collecting models from module  module=role count=1
# INFO  Starting database migration  total_models=2
# INFO  Database migration completed successfully
# INFO  Registering system module  module=user
# INFO  Registering system module  module=role
# INFO  All modules registered successfully
# INFO  Starting server  addr=:8080
```

### 5.4 测试 API

**步骤 12：测试接口**

```bash
# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "123456",
    "nickname": "管理员",
    "email": "admin@example.com"
  }'

# 响应：
# {
#   "code": 0,
#   "message": "success",
#   "data": {
#     "id": 1,
#     "username": "admin",
#     "nickname": "管理员",
#     "email": "admin@example.com"
#   }
# }

# 获取用户
curl http://localhost:8080/api/v1/users/1

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/1
```

---

## 6. 最佳实践

### 6.1 模块设计原则

#### ✅ DO（推荐做法）

1. **模块独立**
```go
// ✅ 模块只依赖 Pocket 资源
type UserService struct {
    db       *gorm.DB                // 来自 Pocket
    enforcer *casbin.SyncedEnforcer  // 来自 Pocket
}
```

2. **通过共享资源交互**
```go
// ✅ 通过数据库查询
func (s *UserService) ValidateRole(roleID uint) error {
    var count int64
    s.db.Model(&Role{}).Where("id = ?", roleID).Count(&count)
    if count == 0 {
        return errors.New("角色不存在")
    }
    return nil
}
```

3. **清晰的分层**
```go
// ✅ API → Service → Repository → Domain
func (h *UserHandler) Create(c *gin.Context) {
    // API 层：参数验证
    var req domain.CreateUserRequest
    c.ShouldBindJSON(&req)
    
    // 调用 Service
    user, err := h.service.Create(&req)
    
    // 返回响应
    response.Success(c, user)
}
```

#### ❌ DON'T（避免做法）

1. **模块间直接依赖**
```go
// ❌ 不要直接依赖其他模块
import "nvo-api/internal/system/role/service"

type UserService struct {
    roleService *roleService.RoleService  // ❌ 危险！
}
```

2. **跨层调用**
```go
// ❌ API 不要直接调用 Repository
func (h *UserHandler) Create(c *gin.Context) {
    user := &User{...}
    h.repository.Create(user)  // ❌ 跳过了 Service 层
}
```

3. **业务逻辑放在 Repository**
```go
// ❌ Repository 不要包含业务逻辑
func (r *UserRepository) CreateWithDefaultRole(user *User) error {
    // ❌ 这是业务逻辑，应该在 Service 层
    if user.RoleID == 0 {
        user.RoleID = 1  // 默认角色
    }
    return r.db.Create(user).Error
}
```

### 6.2 错误处理

```go
// ✅ 统一的错误处理
func (s *UserService) Create(req *CreateUserRequest) (*User, error) {
    // 1. 参数验证
    if req.Username == "" {
        return nil, errors.New("用户名不能为空")
    }
    
    // 2. 业务规则检查
    exists, _ := s.repo.ExistsByUsername(req.Username)
    if exists {
        return nil, errors.New("用户名已存在")
    }
    
    // 3. 数据库操作
    user := &User{...}
    if err := s.repo.Create(user); err != nil {
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }
    
    return user, nil
}
```

### 6.3 日志记录

```go
// ✅ 关键操作记录日志
func (s *UserService) Create(req *CreateUserRequest) (*User, error) {
    log.Info("creating user", zap.String("username", req.Username))
    
    user, err := s.repo.Create(&User{...})
    if err != nil {
        log.Error("failed to create user", 
            zap.String("username", req.Username),
            zap.Error(err))
        return nil, err
    }
    
    log.Info("user created successfully", 
        zap.Uint("id", user.ID),
        zap.String("username", user.Username))
    
    return user, nil
}
```

### 6.4 事务管理

```go
// ✅ Service 层管理事务
func (s *UserService) Delete(id uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 删除用户
        if err := s.repo.Delete(id); err != nil {
            return err
        }
        
        // 2. 删除权限关联
        subject := fmt.Sprintf("user:%d", id)
        if _, err := s.enforcer.DeleteUser(subject); err != nil {
            return err
        }
        
        // 事务自动提交
        return nil
    })
}
```

---

## 7. 常见问题

### Q1: 为什么不用 interface 而直接用 struct？

**A**: 在 Go 中，过早抽象会增加复杂度。

```go
// ❌ 过度设计
type UserServiceInterface interface {
    Create(*CreateUserRequest) (*User, error)
    GetByID(uint) (*User, error)
    // ... 20+ 方法
}

// ✅ 简单直接
type UserService struct {
    repo *UserRepository
}
```

**何时使用 interface**：
- 需要 Mock 测试
- 有多种实现
- 需要依赖倒置

### Q2: 模块之间如何通信？

**A**: 通过共享资源，不直接依赖。

```go
// ✅ 方式 1：通过数据库
func (s *UserService) ValidateRole(roleID uint) error {
    var count int64
    s.db.Model(&Role{}).Where("id = ?", roleID).Count(&count)
    return count > 0
}

// ✅ 方式 2：通过 Casbin
func (s *UserService) GetUserRoles(userID uint) ([]string, error) {
    subject := fmt.Sprintf("user:%d", userID)
    return s.enforcer.GetRolesForUser(subject)
}

// ✅ 方式 3：通过事件（高级）
eventBus.Publish("user.created", UserCreatedEvent{...})
```

### Q3: 如何处理循环依赖？

**A**: Go 编译器会阻止循环依赖，你的设计不会有这个问题。

```go
// ❌ Go 不允许
package user
import "app/role"  // user 导入 role

package role
import "app/user"  // role 导入 user
// 编译错误：import cycle not allowed

// ✅ 通过 Pocket 隔离
package user
// 不导入 role 包

package role
// 不导入 user 包

// 都依赖 Pocket，通过共享资源交互
```

### Q4: 如何测试模块？

**A**: Mock Pocket 或直接传入测试依赖。

```go
// 测试 Service
func TestUserService_Create(t *testing.T) {
    // 1. 准备测试数据库
    db := setupTestDB()
    enforcer := setupTestEnforcer()
    
    // 2. 创建 Service
    service := NewUserService(db, enforcer)
    
    // 3. 测试
    user, err := service.Create(&CreateUserRequest{
        Username: "test",
        Password: "123456",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

### Q5: 如何添加新模块？

**A**: 按照标准结构创建，然后在 registry 中注册。

```bash
# 1. 创建目录结构
mkdir -p internal/system/order/{domain,repository,service,api}

# 2. 实现各层代码
# - domain/order.go
# - repository/order_repository.go
# - service/order_service.go
# - api/order_handler.go
# - module.go

# 3. 在 registry.go 中注册
modules := []internal.Module{
    user.NewModule(p),
    role.NewModule(p),
    order.NewModule(p),  // 新增
}
```

### Q6: 如何处理数据库迁移失败？

**A**: 迁移失败会立即终止程序。

```go
// registry.go
if err := migrateModels(p.DB, modules); err != nil {
    log.Fatal("Database migration failed", zap.Error(err))
    // 程序终止，不会启动服务
}
```

### Q7: 性能如何？

**A**: 模块化设计对性能几乎无影响。

- ✅ 初始化开销：< 100ms
- ✅ 运行时开销：0（直接方法调用）
- ✅ 内存占用：每个模块 < 1KB

---

## 8. 总结

### 核心要点

1. **Pocket 管理依赖**
   - 所有共享资源集中管理
   - 模块通过 Pocket 获取资源

2. **模块独立设计**
   - 实现 Module 接口
   - 标准的分层结构
   - 不直接依赖其他模块

3. **清晰的职责划分**
   - Domain：数据模型
   - Repository：数据访问
   - Service：业务逻辑
   - API：HTTP 处理

4. **统一的注册机制**
   - 自动收集模型
   - 统一迁移
   - 统一路由注册

### 下一步

- 📖 阅读代码示例
- 🔨 动手实践
- 🧪 编写测试
- 📦 添加新模块

---

**祝你构建优雅的 Go 应用！** 🎉
