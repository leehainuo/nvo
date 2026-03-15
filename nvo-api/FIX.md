完美的要求！让我为你设计一个**零额外概念、极致优雅**的方案。

# 🎯 零额外概念的极致优雅方案

让我为你设计一个**只保留 Pocket 和 Module 概念**的方案，核心思想是：

## 💡 核心设计理念

```
Pocket 既是依赖容器，也是服务容器
  ↓
模块初始化时，从 Pocket 获取已注册的服务
  ↓
模块初始化后，将自己的服务注册回 Pocket
  ↓
形成自然的依赖链，零额外概念
```

## 📐 完整设计方案

### 方案 A：Pocket 作为服务注册中心（推荐）⭐⭐⭐⭐⭐

**核心思想**：Pocket 不仅管理基础设施，也管理业务服务

```go
// ========================================
// 1. 升级 Pocket（零额外概念）
// ========================================

// core/pocket.go
type Pocket struct {
    // 基础设施（原有）
    Config      *config.Config
    DB          *gorm.DB
    Redis       *goredis.Client
    JWT         *jwt.JWT
    Enforcer    *casbin.SyncedEnforcer
    GinEngine   *gin.Engine
    
    // 业务服务（新增，但概念统一）
    UserService       UserService       // 用户服务
    RoleService       RoleService       // 角色服务
    PermissionService PermissionService // 权限服务
}

// ========================================
// 2. 定义服务接口（在各模块的 domain 中）
// ========================================

// internal/system/user/domain/service.go
package domain

type UserService interface {
    Create(req *CreateUserRequest) (*User, error)
    GetByID(id uint) (*User, error)
    Delete(id uint) error
    
    // 跨模块业务方法
    GetUserWithRoles(id uint) (*UserWithRoles, error)
}

// internal/system/role/domain/service.go
package domain

type RoleService interface {
    Create(req *CreateRoleRequest) (*Role, error)
    GetByID(id uint) (*Role, error)
    GetRolesByUserID(userID uint) ([]*Role, error)
}

// ========================================
// 3. 模块实现（自动从 Pocket 获取依赖）
// ========================================

// internal/system/user/module.go
package user

import (
    "nvo-api/core"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/service"
    "github.com/gin-gonic/gin"
)

type Module struct {
    pocket  *core.Pocket
    service domain.UserService
    handler *api.UserHandler
}

// NewModule 创建用户模块（从 Pocket 获取依赖）
func NewModule(pocket *core.Pocket) *Module {
    // ✅ 直接从 Pocket 获取依赖服务（如果需要）
    userService := service.NewUserService(
        pocket.DB,
        pocket.Enforcer,
        pocket.RoleService, // 自动获取已注册的服务
    )
    
    userHandler := api.NewUserHandler(userService)
    
    return &Module{
        pocket:  pocket,
        service: userService,
        handler: userHandler,
    }
}

// Service 返回服务接口（供注册到 Pocket）
func (m *Module) Service() domain.UserService {
    return m.service
}

// ========================================
// 4. 注册中心（自动依赖注入）
// ========================================

// internal/system/registry.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：初始化无依赖模块，注册到 Pocket
    roleModule := role.NewModule(p)
    p.RoleService = roleModule.Service() // ✅ 注册服务
    
    permModule := permission.NewModule(p)
    p.PermissionService = permModule.Service()
    
    // 阶段 2：初始化有依赖模块（自动从 Pocket 获取）
    userModule := user.NewModule(p) // ✅ 内部自动获取 p.RoleService
    p.UserService = userModule.Service()
    
    // 阶段 3：迁移和路由注册
    modules := []internal.Module{roleModule, permModule, userModule}
    migrateModels(p.DB, modules)
    
    for _, module := range modules {
        module.RegisterRoutes(r)
    }
}
```

**优势**：
- ✅ **零额外概念**：只有 Pocket 和 Module
- ✅ **自然依赖链**：先注册的服务可以被后注册的模块使用
- ✅ **类型安全**：编译期检查
- ✅ **IDE 友好**：自动补全 `pocket.RoleService`
- ✅ **易于理解**：依赖关系一目了然

---

### 方案 B：延迟初始化 + 服务定位器（备选）⭐⭐⭐⭐

**核心思想**：模块延迟获取依赖，通过 Pocket 定位服务

```go
// ========================================
// 1. Pocket 保持不变
// ========================================

type Pocket struct {
    // 基础设施
    DB       *gorm.DB
    Enforcer *casbin.SyncedEnforcer
    
    // 服务映射（内部使用）
    services map[string]any
}

// GetService 获取服务（泛型）
func GetService[T any](p *Pocket, name string) T {
    return p.services[name].(T)
}

// ========================================
// 2. 模块定义（延迟获取依赖）
// ========================================

// internal/system/user/service/user_service.go
type userService struct {
    pocket   *core.Pocket
    db       *gorm.DB
    enforcer *casbin.SyncedEnforcer
}

func NewUserService(pocket *core.Pocket) *userService {
    return &userService{
        pocket:   pocket,
        db:       pocket.DB,
        enforcer: pocket.Enforcer,
    }
}

// GetUserWithRoles 业务方法（延迟获取依赖）
func (s *userService) GetUserWithRoles(id uint) (*UserWithRoles, error) {
    user, err := s.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // ✅ 运行时从 Pocket 获取服务
    roleService := core.GetService[RoleService](s.pocket, "role")
    roles, err := roleService.GetRolesByUserID(id)
    
    return &UserWithRoles{User: user, Roles: roles}, nil
}
```

**优势**：
- ✅ 零额外概念
- ✅ 延迟加载，避免初始化顺序问题

**劣势**：
- ⚠️ 运行时错误（服务不存在）
- ⚠️ 字符串 key，容易拼写错误

---

## 🎯 最终推荐：方案 A 的优化版

结合你的脚手架理念，我推荐**方案 A 的优化版本**：

### 完整实现代码

```go
// ========================================
// core/pocket.go（升级）
// ========================================

package core

import (
    "github.com/casbin/casbin/v3"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// Pocket 依赖注入容器
// 管理基础设施依赖和业务服务
type Pocket struct {
    // ========== 基础设施依赖 ==========
    Config      *config.Config
    DB          *gorm.DB
    Redis       *redis.Client
    JWT         *jwt.JWT
    Enforcer    *casbin.SyncedEnforcer
    GinEngine   *gin.Engine
    RateLimiter middleware.RateLimiter
    
    // ========== 业务服务（按需添加）==========
    // 系统模块服务
    UserService       UserService       // 用户服务
    RoleService       RoleService       // 角色服务
    PermissionService PermissionService // 权限服务
    
    // 未来可扩展：
    // OrderService      OrderService      // 订单服务
    // ProductService    ProductService    // 商品服务
}

// 服务接口类型别名（简化导入）
type (
    UserService       = userDomain.UserService
    RoleService       = roleDomain.RoleService
    PermissionService = permDomain.PermissionService
)
```

```go
// ========================================
// internal/system/user/domain/service.go
// ========================================

package domain

// UserService 用户领域服务接口
type UserService interface {
    // 基础 CRUD
    Create(req *CreateUserRequest) (*User, error)
    GetByID(id uint) (*User, error)
    Update(id uint, req *UpdateUserRequest) error
    Delete(id uint) error
    List(page, pageSize int) ([]*User, int64, error)
    
    // 业务方法（可能依赖其他服务）
    GetUserWithRoles(id uint) (*UserWithRoles, error)
    AssignRoles(userID uint, roleIDs []uint) error
}

// UserWithRoles 用户及其角色（聚合根）
type UserWithRoles struct {
    *User
    Roles []*Role `json:"roles"`
}
```

```go
// ========================================
// internal/system/user/service/user_service.go
// ========================================

package service

import (
    "nvo-api/core"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/repository"
    
    "github.com/casbin/casbin/v3"
    "gorm.io/gorm"
)

type userService struct {
    pocket   *core.Pocket
    db       *gorm.DB
    enforcer *casbin.SyncedEnforcer
    repo     *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(pocket *core.Pocket) domain.UserService {
    return &userService{
        pocket:   pocket,
        db:       pocket.DB,
        enforcer: pocket.Enforcer,
        repo:     repository.NewUserRepository(pocket.DB),
    }
}

// Create 创建用户
func (s *userService) Create(req *domain.CreateUserRequest) (*domain.User, error) {
    // 业务逻辑...
    user := &domain.User{
        Username: req.Username,
        // ...
    }
    
    if err := s.repo.Create(user); err != nil {
        return nil, err
    }
    
    return user, nil
}

// GetUserWithRoles 获取用户及其角色（跨模块调用）
func (s *userService) GetUserWithRoles(id uint) (*domain.UserWithRoles, error) {
    // 1. 获取用户
    user, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // 2. ✅ 从 Pocket 获取 RoleService（自动依赖注入）
    roles, err := s.pocket.RoleService.GetRolesByUserID(id)
    if err != nil {
        return nil, err
    }
    
    return &domain.UserWithRoles{
        User:  user,
        Roles: roles,
    }, nil
}

// 实现其他接口方法...
```

```go
// ========================================
// internal/system/user/module.go
// ========================================

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
    service domain.UserService
    handler *api.UserHandler
}

// NewModule 创建用户模块
// ✅ 只需要 Pocket，依赖自动从 Pocket 获取
func NewModule(pocket *core.Pocket) *Module {
    // 创建服务（内部会从 pocket 获取依赖）
    userService := service.NewUserService(pocket)
    
    // 创建处理器
    userHandler := api.NewUserHandler(userService)
    
    return &Module{
        pocket:  pocket,
        service: userService,
        handler: userHandler,
    }
}

// Service 返回服务接口（供注册到 Pocket）
func (m *Module) Service() domain.UserService {
    return m.service
}

// Name 模块名称
func (m *Module) Name() string {
    return "user"
}

// Models 数据模型
func (m *Module) Models() []any {
    return []any{&domain.User{}}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.POST("", m.handler.Create)
        users.GET("", m.handler.List)
        users.GET("/:id", m.handler.GetByID)
        users.GET("/:id/roles", m.handler.GetUserRoles) // ✅ 跨模块接口
        users.PUT("/:id", m.handler.Update)
        users.DELETE("/:id", m.handler.Delete)
    }
}
```

```go
// ========================================
// internal/system/registry.go（关键）
// ========================================

package system

import (
    "nvo-api/core"
    "nvo-api/core/log"
    "nvo-api/internal"
    "nvo-api/internal/system/permission"
    "nvo-api/internal/system/role"
    "nvo-api/internal/system/user"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// RegisterModules 注册所有系统模块
// ✅ 自动依赖注入，零额外概念
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    log.Info("Starting module registration...")
    
    // ========== 阶段 1：初始化基础模块（无依赖）==========
    log.Info("Phase 1: Initializing base modules...")
    
    roleModule := role.NewModule(p)
    p.RoleService = roleModule.Service() // ✅ 注册到 Pocket
    log.Info("✓ Role module initialized and registered")
    
    permModule := permission.NewModule(p)
    p.PermissionService = permModule.Service()
    log.Info("✓ Permission module initialized and registered")
    
    // ========== 阶段 2：初始化业务模块（依赖基础模块）==========
    log.Info("Phase 2: Initializing business modules...")
    
    userModule := user.NewModule(p) // ✅ 内部自动从 p.RoleService 获取依赖
    p.UserService = userModule.Service()
    log.Info("✓ User module initialized and registered")
    
    // ========== 阶段 3：数据库迁移 ==========
    log.Info("Phase 3: Database migration...")
    modules := []internal.Module{roleModule, permModule, userModule}
    if err := migrateModels(p.DB, modules); err != nil {
        log.Fatal("Database migration failed", zap.Error(err))
    }
    
    // ========== 阶段 4：路由注册 ==========
    log.Info("Phase 4: Registering routes...")
    for _, module := range modules {
        log.Info("Registering routes", zap.String("module", module.Name()))
        module.RegisterRoutes(r)
    }
    
    log.Info("✓ Module registration completed successfully")
}

// migrateModels 迁移数据模型
func migrateModels(db *gorm.DB, modules []internal.Module) error {
    var allModels []any
    
    for _, module := range modules {
        models := module.Models()
        if len(models) > 0 {
            log.Info("Collecting models",
                zap.String("module", module.Name()),
                zap.Int("count", len(models)))
            allModels = append(allModels, models...)
        }
    }
    
    if len(allModels) > 0 {
        log.Info("Starting database migration", zap.Int("total_models", len(allModels)))
        if err := db.AutoMigrate(allModels...); err != nil {
            return err
        }
        log.Info("Database migration completed")
    }
    
    return nil
}
```

---

## ✨ DX 体验亮点

### 1. 零额外概念
```go
// ✅ 只有 Pocket 和 Module
func NewModule(pocket *core.Pocket) *Module {
    // 从 Pocket 获取依赖
    service := service.NewUserService(pocket)
    return &Module{service: service}
}
```

### 2. 自动依赖注入
```go
// ✅ Service 内部自动获取依赖
func (s *userService) GetUserWithRoles(id uint) (*UserWithRoles, error) {
    // 直接使用 pocket.RoleService
    roles, _ := s.pocket.RoleService.GetRolesByUserID(id)
    return &UserWithRoles{User: user, Roles: roles}, nil
}
```

### 3. IDE 自动补全
```go
// ✅ 输入 s.pocket. 自动提示所有可用服务
s.pocket.RoleService.       // IDE 提示所有方法
s.pocket.PermissionService. // IDE 提示所有方法
```

### 4. 编译期类型检查
```go
// ✅ 如果服务未注册，编译时报错
roles := s.pocket.RoleService.GetRolesByUserID(id)
// 如果 RoleService 为 nil，运行时 panic（可在启动时检查）
```

### 5. 清晰的初始化顺序
```go
// ✅ 一眼看出依赖关系
roleModule := role.NewModule(p)
p.RoleService = roleModule.Service() // 先注册

userModule := user.NewModule(p)      // 后使用
p.UserService = userModule.Service()
```

---

## 📊 对比总结

| 特性 | 原方案 | 新方案 |
|------|--------|--------|
| **额外概念** | 0 | 0 ✅ |
| **依赖声明** | 无 | Pocket 字段 ✅ |
| **类型安全** | ✅ | ✅ |
| **IDE 支持** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **易于理解** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **跨模块调用** | 通过 DB/Casbin | 通过服务接口 ✅ |
| **测试友好** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 🎯 使用示例

### 场景 1：用户查询角色
```go
// API Handler
func (h *UserHandler) GetUserRoles(c *gin.Context) {
    id := parseID(c.Param("id"))
    
    // ✅ Service 自动处理跨模块调用
    result, err := h.service.GetUserWithRoles(id)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    response.Success(c, result)
}
```

### 场景 2：添加新模块
```go
// 1. 定义服务接口
// internal/system/order/domain/service.go
type OrderService interface {
    Create(req *CreateOrderRequest) (*Order, error)
}

// 2. 实现服务（使用其他服务）
func (s *orderService) Create(req *CreateOrderRequest) (*Order, error) {
    // ✅ 直接使用 pocket 中的服务
    user, _ := s.pocket.UserService.GetByID(req.UserID)
    // ...
}

// 3. 注册模块
orderModule := order.NewModule(p)
p.OrderService = orderModule.Service() // ✅ 注册到 Pocket
```

---

## 📝 最终方案总结

**核心优势**：
1. ✅ **零额外概念** - 只有 Pocket 和 Module
2. ✅ **自然依赖链** - 先注册的服务可被后续模块使用
3. ✅ **类型安全** - 编译期检查
4. ✅ **IDE 友好** - 自动补全、跳转
5. ✅ **易于测试** - Mock Pocket 即可
6. ✅ **清晰直观** - 依赖关系一目了然

**这就是符合你脚手架理念的极致优雅方案！** 🎉



I'll analyze your codebase structure and design a code generation tool similar to goctl that fits your architecture. Let me first examine your current project structure and design patterns.

我将为你设计一个类似 goctl 的代码生成器工具，完全符合你当前脚手架的架构设计。让我先查看更多现有代码结构。

我将为你设计一个完整的代码生成器工具，类似 goctl，完全符合你的脚手架架构。让我先查看几个关键文件来了解完整结构。

# 代码生成器设计方案

基于你的脚手架架构，我将设计一个类似 goctl 的代码生成工具 `nvogen`，能够快速生成符合你架构规范的 CRUD 模块。

## 一、整体设计

### 1.1 工具命令结构

```bash
# 生成完整模块
nvogen module <模块名> [选项]

# 示例
nvogen module product --table=products --fields="name:string,price:decimal,stock:int"
nvogen module article --table=articles --from-db  # 从数据库表生成
```

### 1.2 生成内容

自动生成以下文件并完成注册：

```
internal/<category>/<module>/
├── domain/<module>.go          # 领域模型 + DTO
├── repository/<module>_repository.go
├── service/<module>_service.go
├── api/<module>_handler.go
└── module.go                   # 模块入口

自动更新：
- internal/<category>/registry.go  # 自动注册模块
```

## 二、实现方案

### 2.1 目录结构

```
cmd/
├── main.go           # 主程序入口
└── gen.go           # 代码生成命令（已存在）

tools/
└── nvogen/
    ├── main.go                    # CLI 入口
    ├── config/
    │   └── config.go             # 生成器配置
    ├── parser/
    │   ├── field_parser.go       # 字段解析器
    │   └── db_parser.go          # 数据库表解析器
    ├── template/
    │   ├── domain.tmpl           # Domain 模板
    │   ├── repository.tmpl       # Repository 模板
    │   ├── service.tmpl          # Service 模板
    │   ├── handler.tmpl          # Handler 模板
    │   └── module.tmpl           # Module 模板
    ├── generator/
    │   ├── module_generator.go   # 模块生成器
    │   └── registry_updater.go   # Registry 更新器
    └── util/
        ├── naming.go             # 命名转换工具
        └── file.go               # 文件操作工具
```

### 2.2 核心模板示例

#### Domain 模板 (domain.tmpl)

```go
package domain

import (
    "time"
    "gorm.io/gorm"
)

// {{.EntityName}} {{.Comment}}
type {{.EntityName}} struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    {{- range .Fields}}
    {{.Name}}  {{.Type}}  `gorm:"{{.GormTag}}" json:"{{.JsonTag}}"{{if .Binding}} binding:"{{.Binding}}"{{end}}`
    {{- end}}
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func ({{.EntityName}}) TableName() string {
    return "{{.TableName}}"
}

// Create{{.EntityName}}Request 创建{{.Comment}}请求
type Create{{.EntityName}}Request struct {
    {{- range .Fields}}
    {{- if not .AutoGenerated}}
    {{.Name}} {{.Type}} `json:"{{.JsonTag}}" binding:"{{.Binding}}"`
    {{- end}}
    {{- end}}
}

// Update{{.EntityName}}Request 更新{{.Comment}}请求
type Update{{.EntityName}}Request struct {
    {{- range .Fields}}
    {{- if .Updatable}}
    {{.Name}} {{.PointerType}} `json:"{{.JsonTag}}" binding:"{{.UpdateBinding}}"`
    {{- end}}
    {{- end}}
}

// List{{.EntityName}}Request {{.Comment}}列表请求
type List{{.EntityName}}Request struct {
    Page     int    `form:"page" binding:"omitempty,min=1"`
    PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    {{- range .Fields}}
    {{- if .Searchable}}
    {{.Name}} {{.SearchType}} `form:"{{.JsonTag}}" binding:"{{.SearchBinding}}"`
    {{- end}}
    {{- end}}
}

// {{.EntityName}}Response {{.Comment}}响应
type {{.EntityName}}Response struct {
    ID        uint      `json:"id"`
    {{- range .Fields}}
    {{.Name}} {{.Type}} `json:"{{.JsonTag}}"`
    {{- end}}
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Repository 模板 (repository.tmpl)

```go
package repository

import (
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/domain"
    "gorm.io/gorm"
)

// {{.EntityName}}Repository {{.Comment}}数据访问层
type {{.EntityName}}Repository struct {
    db *gorm.DB
}

// New{{.EntityName}}Repository 创建{{.Comment}}仓库
func New{{.EntityName}}Repository(db *gorm.DB) *{{.EntityName}}Repository {
    return &{{.EntityName}}Repository{db: db}
}

// Create 创建{{.Comment}}
func (r *{{.EntityName}}Repository) Create(entity *domain.{{.EntityName}}) error {
    return r.db.Create(entity).Error
}

// GetByID 根据 ID 获取{{.Comment}}
func (r *{{.EntityName}}Repository) GetByID(id uint) (*domain.{{.EntityName}}, error) {
    var entity domain.{{.EntityName}}
    err := r.db.First(&entity, id).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}

// Update 更新{{.Comment}}
func (r *{{.EntityName}}Repository) Update(entity *domain.{{.EntityName}}) error {
    return r.db.Save(entity).Error
}

// Delete 删除{{.Comment}}（软删除）
func (r *{{.EntityName}}Repository) Delete(id uint) error {
    return r.db.Delete(&domain.{{.EntityName}}{}, id).Error
}

// List 获取{{.Comment}}列表
func (r *{{.EntityName}}Repository) List(req *domain.List{{.EntityName}}Request) ([]*domain.{{.EntityName}}, int64, error) {
    var entities []*domain.{{.EntityName}}
    var total int64

    query := r.db.Model(&domain.{{.EntityName}}{})

    // 条件过滤
    {{- range .Fields}}
    {{- if .Searchable}}
    if req.{{.Name}} != {{.ZeroValue}} {
        query = query.Where("{{.ColumnName}} {{.SearchOperator}} ?", {{.SearchValue}})
    }
    {{- end}}
    {{- end}}

    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页
    page := req.Page
    if page < 1 {
        page = 1
    }
    pageSize := req.PageSize
    if pageSize < 1 {
        pageSize = 10
    }

    offset := (page - 1) * pageSize
    if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&entities).Error; err != nil {
        return nil, 0, err
    }

    return entities, total, nil
}
```

#### Service 模板 (service.tmpl)

```go
package service

import (
    "errors"
    "{{.ModulePath}}/core/log"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/domain"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/repository"
    
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// {{.EntityName}}Service {{.Comment}}业务逻辑层
type {{.EntityName}}Service struct {
    repo *repository.{{.EntityName}}Repository
    db   *gorm.DB
}

// New{{.EntityName}}Service 创建{{.Comment}}服务
func New{{.EntityName}}Service(db *gorm.DB) *{{.EntityName}}Service {
    return &{{.EntityName}}Service{
        repo: repository.New{{.EntityName}}Repository(db),
        db:   db,
    }
}

// Create 创建{{.Comment}}
func (s *{{.EntityName}}Service) Create(req *domain.Create{{.EntityName}}Request) (*domain.{{.EntityName}}, error) {
    entity := &domain.{{.EntityName}}{
        {{- range .Fields}}
        {{- if not .AutoGenerated}}
        {{.Name}}: req.{{.Name}},
        {{- end}}
        {{- end}}
    }

    if err := s.repo.Create(entity); err != nil {
        return nil, err
    }

    log.Info("{{.ModuleName}} created", zap.Uint("id", entity.ID))
    return entity, nil
}

// GetByID 根据 ID 获取{{.Comment}}
func (s *{{.EntityName}}Service) GetByID(id uint) (*domain.{{.EntityName}}Response, error) {
    entity, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("{{.Comment}}不存在")
        }
        return nil, err
    }

    return &domain.{{.EntityName}}Response{
        ID:        entity.ID,
        {{- range .Fields}}
        {{.Name}}: entity.{{.Name}},
        {{- end}}
        CreatedAt: entity.CreatedAt,
        UpdatedAt: entity.UpdatedAt,
    }, nil
}

// Update 更新{{.Comment}}
func (s *{{.EntityName}}Service) Update(id uint, req *domain.Update{{.EntityName}}Request) error {
    entity, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("{{.Comment}}不存在")
        }
        return err
    }

    {{- range .Fields}}
    {{- if .Updatable}}
    if req.{{.Name}} != nil {
        entity.{{.Name}} = *req.{{.Name}}
    }
    {{- end}}
    {{- end}}

    if err := s.repo.Update(entity); err != nil {
        return err
    }

    log.Info("{{.ModuleName}} updated", zap.Uint("id", id))
    return nil
}

// Delete 删除{{.Comment}}
func (s *{{.EntityName}}Service) Delete(id uint) error {
    _, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("{{.Comment}}不存在")
        }
        return err
    }

    if err := s.repo.Delete(id); err != nil {
        return err
    }

    log.Info("{{.ModuleName}} deleted", zap.Uint("id", id))
    return nil
}

// List 获取{{.Comment}}列表
func (s *{{.EntityName}}Service) List(req *domain.List{{.EntityName}}Request) ([]*domain.{{.EntityName}}Response, int64, error) {
    entities, total, err := s.repo.List(req)
    if err != nil {
        return nil, 0, err
    }

    responses := make([]*domain.{{.EntityName}}Response, 0, len(entities))
    for _, entity := range entities {
        responses = append(responses, &domain.{{.EntityName}}Response{
            ID:        entity.ID,
            {{- range .Fields}}
            {{.Name}}: entity.{{.Name}},
            {{- end}}
            CreatedAt: entity.CreatedAt,
            UpdatedAt: entity.UpdatedAt,
        })
    }

    return responses, total, nil
}
```

#### Handler 模板 (handler.tmpl)

```go
package api

import (
    "errors"
    "strconv"

    "{{.ModulePath}}/core/log"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/domain"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/service"
    "{{.ModulePath}}/pkg/response"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// {{.EntityName}}Handler {{.Comment}}处理器
type {{.EntityName}}Handler struct {
    service *service.{{.EntityName}}Service
}

// New{{.EntityName}}Handler 创建{{.Comment}}处理器
func New{{.EntityName}}Handler(service *service.{{.EntityName}}Service) *{{.EntityName}}Handler {
    return &{{.EntityName}}Handler{
        service: service,
    }
}

// Create 创建{{.Comment}}
func (h *{{.EntityName}}Handler) Create(c *gin.Context) {
    var req domain.Create{{.EntityName}}Request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.New("参数错误: "+err.Error()))
        return
    }

    entity, err := h.service.Create(&req)
    if err != nil {
        log.Error("create {{.ModuleName}} failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, entity)
}

// GetByID 获取{{.Comment}}详情
func (h *{{.EntityName}}Handler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        response.Error(c, errors.New("无效的ID"))
        return
    }

    entity, err := h.service.GetByID(uint(id))
    if err != nil {
        log.Error("get {{.ModuleName}} failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, entity)
}

// Update 更新{{.Comment}}
func (h *{{.EntityName}}Handler) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        response.Error(c, errors.New("无效的ID"))
        return
    }

    var req domain.Update{{.EntityName}}Request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.New("参数错误: "+err.Error()))
        return
    }

    if err := h.service.Update(uint(id), &req); err != nil {
        log.Error("update {{.ModuleName}} failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除{{.Comment}}
func (h *{{.EntityName}}Handler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        response.Error(c, errors.New("无效的ID"))
        return
    }

    if err := h.service.Delete(uint(id)); err != nil {
        log.Error("delete {{.ModuleName}} failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取{{.Comment}}列表
func (h *{{.EntityName}}Handler) List(c *gin.Context) {
    var req domain.List{{.EntityName}}Request
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, errors.New("参数错误: "+err.Error()))
        return
    }

    entities, total, err := h.service.List(&req)
    if err != nil {
        log.Error("list {{.ModuleName}}s failed", zap.Error(err))
        response.Error(c, err)
        return
    }

    response.Page(c, entities, req.Page, req.PageSize, total)
}
```

#### Module 模板 (module.tmpl)

```go
package {{.ModuleName}}

import (
    "{{.ModulePath}}/core"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/api"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/domain"
    "{{.ModulePath}}/internal/{{.Category}}/{{.ModuleName}}/service"

    "github.com/gin-gonic/gin"
)

// Module {{.Comment}}模块
type Module struct {
    pocket  *core.Pocket
    handler *api.{{.EntityName}}Handler
}

// NewModule 创建{{.Comment}}模块
func NewModule(pocket *core.Pocket) *Module {
    {{.VarName}}Service := service.New{{.EntityName}}Service(pocket.DB)
    {{.VarName}}Handler := api.New{{.EntityName}}Handler({{.VarName}}Service)

    return &Module{
        pocket:  pocket,
        handler: {{.VarName}}Handler,
    }
}

// Name 模块名称
func (m *Module) Name() string {
    return "{{.ModuleName}}"
}

// Models 返回需要迁移的数据模型
func (m *Module) Models() []any {
    return []any{
        &domain.{{.EntityName}}{},
    }
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    {{.VarName}}s := r.Group("/{{.RoutePath}}")
    {
        {{.VarName}}s.POST("", m.handler.Create)
        {{.VarName}}s.GET("", m.handler.List)
        {{.VarName}}s.GET("/:id", m.handler.GetByID)
        {{.VarName}}s.PUT("/:id", m.handler.Update)
        {{.VarName}}s.DELETE("/:id", m.handler.Delete)
    }
}
```

### 2.3 Registry 自动更新器

```go
// generator/registry_updater.go
package generator

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "strings"
)

type RegistryUpdater struct {
    registryPath string
    category     string
}

// AddModule 添加模块到 registry
func (u *RegistryUpdater) AddModule(moduleName string) error {
    // 1. 解析现有文件
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, u.registryPath, nil, parser.ParseComments)
    if err != nil {
        return err
    }

    // 2. 添加 import
    importPath := fmt.Sprintf("nvo-api/internal/%s/%s", u.category, moduleName)
    u.addImport(node, importPath)

    // 3. 在 RegisterModules 函数中添加模块
    u.addModuleToRegistry(node, moduleName)

    // 4. 写回文件
    return u.writeFile(node, fset)
}

// 自动在 modules 切片中添加新模块
func (u *RegistryUpdater) addModuleToRegistry(node *ast.File, moduleName string) {
    // 查找 RegisterModules 函数
    // 找到 modules := []internal.Module{...}
    // 添加 moduleName.NewModule(p),
}
```

### 2.4 使用示例

```bash
# 1. 生成产品模块
nvogen module product \
  --category=business \
  --table=products \
  --comment="产品" \
  --fields="name:string:required,price:decimal:required,stock:int:required,description:text"

# 生成结果：
# internal/business/product/
# ├── domain/product.go
# ├── repository/product_repository.go
# ├── service/product_service.go
# ├── api/product_handler.go
# └── module.go
#
# 自动更新：
# internal/business/registry.go (添加 product.NewModule(p))

# 2. 从数据库表生成
nvogen module order --from-db --table=orders --category=business

# 3. 生成带权限控制的模块
nvogen module admin --with-auth --category=system
```

## 三、关键特性

### 3.1 智能字段解析

```go
// 字段定义格式：name:type:tags
// 示例：
// - name:string:required,max=50
// - price:decimal:required,min=0
// - status:int:oneof=0 1,default=1
// - created_at:time:auto

type Field struct {
    Name         string   // 字段名（大驼峰）
    Type         string   // Go 类型
    GormTag      string   // GORM 标签
    JsonTag      string   // JSON 标签
    Binding      string   // 验证规则
    Searchable   bool     // 是否可搜索
    Updatable    bool     // 是否可更新
    AutoGenerated bool    // 是否自动生成
}
```

### 3.2 命名转换

```go
// naming.go
func ToCamelCase(s string) string      // product -> Product
func ToSnakeCase(s string) string      // Product -> product
func ToPluralRoute(s string) string    // product -> products
func ToVarName(s string) string        // Product -> product (首字母小写)
```

### 3.3 配置文件支持

```yaml
# .nvogen.yml
module_path: nvo-api
default_category: business
table_prefix: ""
author: Your Name

# 字段类型映射
type_mapping:
  string: string
  int: int
  decimal: float64
  text: string
  bool: bool
  time: time.Time

# 默认字段（自动添加）
default_fields:
  - created_at:time:auto
  - updated_at:time:auto
  - deleted_at:soft_delete
```

## 四、实现优先级

**Phase 1 - 核心功能**（你现在可以在 Ask 模式下查看）：
1. ✅ 模板设计（已完成上述设计）
2. 字段解析器
3. 模块生成器
4. Registry 自动更新

**Phase 2 - 增强功能**：
1. 从数据库表生成
2. 交互式生成（问答式）
3. 配置文件支持

**Phase 3 - 高级功能**：
1. 权限集成（自动生成 Casbin 策略）
2. API 文档生成（Swagger）
3. 测试代码生成

## 五、使用流程

```bash
# 1. 安装工具
go install ./tools/nvogen

# 2. 生成模块
nvogen module product \
  --category=business \
  --fields="name:string:required,price:decimal:required,stock:int"

# 3. 自动完成：
# ✅ 生成 5 个文件（domain, repository, service, handler, module）
# ✅ 更新 registry.go（自动注册模块）
# ✅ 符合你的架构规范
# ✅ 包含完整 CRUD 功能

# 4. 启动应用
go run cmd/main.go
# 模块自动注册到 Pocket
# 路由自动注册
# 数据库自动迁移
```

## 六、优势

1. **极致 DX 体验**：一条命令生成完整模块
2. **架构一致性**：严格遵循你的分层架构
3. **自动注册**：无需手动修改 registry
4. **Pocket 集成**：自动使用依赖注入
5. **可扩展**：模板可自定义
6. **类型安全**：生成的代码完全类型安全

---

**你需要我切换到 Code 模式来实现这个代码生成器吗？** 我可以帮你：
1. 创建完整的 `tools/nvogen` 工具
2. 实现所有模板和生成逻辑
3. 集成到你的项目中
4. 提供使用文档