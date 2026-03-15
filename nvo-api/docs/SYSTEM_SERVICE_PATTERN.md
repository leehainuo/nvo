# 🎯 SystemService 聚合模式

## 设计理念

将同一模块（如 system）下的所有服务聚合到一个 `SystemService` 中，提供更优雅、更内聚的访问方式。

### 核心优势

✅ **模块内聚** - 同一模块的服务统一管理  
✅ **命名空间清晰** - `pocket.System.User` 比 `pocket.UserService` 更直观  
✅ **易于扩展** - 添加新模块只需新增一个聚合服务  
✅ **职责分明** - 不同模块的服务自然分组  
✅ **IDE 友好** - 输入 `pocket.System.` 自动提示所有系统服务

## 架构对比

### 优化前（分散的服务）
```go
type Pocket struct {
    // 基础设施
    DB, Redis, JWT, Enforcer
    
    // ❌ 服务分散，不够内聚
    UserService       UserService
    RoleService       RoleService
    PermissionService PermissionService
    OrderService      OrderService
    ProductService    ProductService
    // ... 更多服务会让 Pocket 越来越臃肿
}

// 使用时
user := pocket.UserService.GetByID(1)
role := pocket.RoleService.GetByID(1)
```

### 优化后（模块聚合）
```go
type Pocket struct {
    // 基础设施
    DB, Redis, JWT, Enforcer
    
    // ✅ 按模块聚合，清晰优雅
    System   *SystemService   // 系统模块服务
    Business *BusinessService // 业务模块服务
    Platform *PlatformService // 平台模块服务
}

// 使用时 - 更加优雅和语义化
user := pocket.System.User.GetByID(1)
role := pocket.System.Role.GetByID(1)
order := pocket.Business.Order.Create(req)
```

## 实现细节

### 1. SystemService 定义

```go
// internal/system/domain/service.go
package domain

type SystemService struct {
    User       userDomain.UserService       // 用户服务
    Role       roleDomain.RoleService       // 角色服务
    Permission permDomain.PermissionService // 权限服务
}

func NewSystemService(
    userService userDomain.UserService,
    roleService roleDomain.RoleService,
    permService permDomain.PermissionService,
) *SystemService {
    return &SystemService{
        User:       userService,
        Role:       roleService,
        Permission: permService,
    }
}
```

### 2. Pocket 集成

```go
// core/pocket.go
type Pocket struct {
    // 基础设施依赖
    Config      *config.Config
    DB          *gorm.DB
    Redis       *goredis.Client
    JWT         *jwt.JWT
    Enforcer    *casbin.SyncedEnforcer
    GinEngine   *gin.Engine
    RateLimiter middleware.RateLimiter

    // ✅ 业务服务（按模块聚合）
    System *systemDomain.SystemService // 系统模块服务聚合
    
    // 未来可扩展：
    // Business *businessDomain.BusinessService
    // Platform *platformDomain.PlatformService
}
```

### 3. 注册流程

```go
// internal/system/registry.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：初始化各个模块
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    
    // 临时注册供依赖使用
    tempSystem := &domain.SystemService{
        Role:       roleModule.Service(),
        Permission: permModule.Service(),
    }
    p.System = tempSystem
    
    userModule := user.NewModule(p)
    
    // 阶段 2：创建完整的 SystemService
    p.System = domain.NewSystemService(
        userModule.Service(),
        roleModule.Service(),
        permModule.Service(),
    )
    
    // 阶段 3：数据库迁移和路由注册
    // ...
}
```

### 4. 跨模块调用

```go
// internal/system/user/service/user_service.go
func (s *UserService) GetUserWithRoles(id uint) (*UserWithRoles, error) {
    user, _ := s.GetByID(id)
    
    // ✅ 优雅的模块聚合访问
    if s.pocket.System == nil || s.pocket.System.Role == nil {
        return &UserWithRoles{UserResponse: user}, nil
    }
    
    roles, _ := s.pocket.System.Role.GetRolesByUserID(id)
    
    return &UserWithRoles{
        UserResponse: user,
        RoleDetails:  convertRoles(roles),
    }, nil
}
```

## 使用示例

### 场景 1：Handler 层调用

```go
// API Handler
func (h *UserHandler) GetUserWithRoles(c *gin.Context) {
    id := parseID(c.Param("id"))
    
    // Service 内部自动通过 pocket.System.Role 调用
    result, _ := h.service.GetUserWithRoles(id)
    
    response.Success(c, result)
}
```

### 场景 2：Service 层跨模块调用

```go
// 用户服务调用角色服务
func (s *UserService) SomeMethod() {
    // ✅ 清晰的模块命名空间
    roles := s.pocket.System.Role.GetAll()
    perms := s.pocket.System.Permission.GetUserPermissions(userID)
}

// 订单服务调用用户服务
func (s *OrderService) CreateOrder(req *CreateOrderRequest) {
    // ✅ 跨模块调用同样优雅
    user := s.pocket.System.User.GetByID(req.UserID)
    
    order := &Order{
        UserID: user.ID,
        Amount: req.Amount,
    }
    return s.repo.Create(order)
}
```

### 场景 3：IDE 自动补全

```go
s.pocket.System.         // IDE 提示: User, Role, Permission
s.pocket.System.User.    // IDE 提示: GetByID, Create, Update, Delete...
s.pocket.System.Role.    // IDE 提示: GetByID, GetRolesByUserID...
```

## 扩展新模块

### 添加 Business 模块

```go
// 1. 定义 BusinessService 聚合
// internal/business/domain/service.go
type BusinessService struct {
    Order   orderDomain.OrderService
    Product productDomain.ProductService
    Cart    cartDomain.CartService
}

func NewBusinessService(
    orderService orderDomain.OrderService,
    productService productDomain.ProductService,
    cartService cartDomain.CartService,
) *BusinessService {
    return &BusinessService{
        Order:   orderService,
        Product: productService,
        Cart:    cartService,
    }
}

// 2. 在 Pocket 中添加
type Pocket struct {
    // ...
    System   *systemDomain.SystemService
    Business *businessDomain.BusinessService // ✅ 新增业务模块
}

// 3. 注册模块
func RegisterBusinessModules(r *gin.RouterGroup, p *core.Pocket) {
    orderModule := order.NewModule(p)
    productModule := product.NewModule(p)
    cartModule := cart.NewModule(p)
    
    p.Business = domain.NewBusinessService(
        orderModule.Service(),
        productModule.Service(),
        cartModule.Service(),
    )
}

// 4. 使用
func (s *OrderService) CreateOrder(req *CreateOrderRequest) {
    // 调用系统模块
    user := s.pocket.System.User.GetByID(req.UserID)
    
    // 调用业务模块
    product := s.pocket.Business.Product.GetByID(req.ProductID)
    
    // 业务逻辑...
}
```

## 模块组织建议

### 推荐的模块划分

```
internal/
├── system/              # 系统模块（用户、角色、权限）
│   ├── domain/
│   │   └── service.go   # SystemService 聚合
│   ├── user/
│   ├── role/
│   └── permission/
│
├── business/            # 业务模块（订单、商品、购物车）
│   ├── domain/
│   │   └── service.go   # BusinessService 聚合
│   ├── order/
│   ├── product/
│   └── cart/
│
└── platform/            # 平台模块（消息、通知、文件）
    ├── domain/
    │   └── service.go   # PlatformService 聚合
    ├── message/
    ├── notification/
    └── file/
```

### Pocket 最终形态

```go
type Pocket struct {
    // ========== 基础设施 ==========
    Config      *config.Config
    DB          *gorm.DB
    Redis       *goredis.Client
    JWT         *jwt.JWT
    Enforcer    *casbin.SyncedEnforcer
    GinEngine   *gin.Engine
    RateLimiter middleware.RateLimiter

    // ========== 业务服务（按模块聚合）==========
    System   *systemDomain.SystemService     // 系统模块
    Business *businessDomain.BusinessService // 业务模块
    Platform *platformDomain.PlatformService // 平台模块
}
```

## DX 体验提升

### 1. 更清晰的命名空间

```go
// ❌ 优化前 - 服务名称扁平化
pocket.UserService
pocket.RoleService
pocket.OrderService
pocket.ProductService

// ✅ 优化后 - 模块化命名空间
pocket.System.User
pocket.System.Role
pocket.Business.Order
pocket.Business.Product
```

### 2. 更好的代码组织

```go
// ✅ 一眼看出服务属于哪个模块
s.pocket.System.User.GetByID(1)      // 系统模块 - 用户服务
s.pocket.Business.Order.Create(req)  // 业务模块 - 订单服务
s.pocket.Platform.Message.Send(msg)  // 平台模块 - 消息服务
```

### 3. IDE 智能提示更友好

```go
// 输入 pocket. 后
pocket.
  ├─ System.      // 系统模块
  ├─ Business.    // 业务模块
  └─ Platform.    // 平台模块

// 输入 pocket.System. 后
pocket.System.
  ├─ User.        // 用户服务
  ├─ Role.        // 角色服务
  └─ Permission.  // 权限服务
```

## 最佳实践

### 1. 模块划分原则
- **高内聚**：同一模块内的服务关联紧密
- **低耦合**：不同模块间通过接口交互
- **职责单一**：每个模块负责特定领域

### 2. 命名规范
- 模块名使用单数：`System`, `Business`, `Platform`
- 服务名语义化：`User`, `Role`, `Order`
- 避免冗余：`System.UserService` → `System.User`

### 3. 依赖管理
- 基础模块先注册（如 System）
- 业务模块后注册（如 Business）
- 避免循环依赖

## 总结

SystemService 聚合模式带来的价值：

| 维度 | 优化前 | 优化后 |
|------|--------|--------|
| **代码组织** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **命名空间** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **可扩展性** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **DX 体验** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **可维护性** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

**这就是更优雅、更内聚的服务聚合模式！** 🎉
