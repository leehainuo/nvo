# 依赖管理最佳实践

## 📚 目录

1. [核心原则](#核心原则)
2. [单向依赖（推荐）](#单向依赖推荐)
3. [双向依赖（循环依赖）](#双向依赖循环依赖)
4. [完整示例](#完整示例)
5. [常见问题](#常见问题)

---

## 核心原则

### 依赖倒置原则（DIP）

本脚手架严格遵循**依赖倒置原则**，这是解决模块间依赖的核心设计理念：

```
高层模块不应该依赖低层模块，两者都应该依赖抽象
抽象不应该依赖细节，细节应该依赖抽象
```

### 架构分层

```
┌─────────────────────────────────────┐
│         接口层（domain）             │
│  - 定义服务接口                      │
│  - 定义 DTO 和领域模型               │
│  - domain 包之间不互相导入           │
└──────────────┬──────────────────────┘
               │ 依赖（向上）
┌──────────────▼──────────────────────┐
│         实现层（service）            │
│  - 实现服务接口                      │
│  - 依赖其他模块的接口                │
│  - 不依赖其他模块的实现              │
└─────────────────────────────────────┘
```

---

## 单向依赖（推荐）

### 适用场景

- ✅ 模块 A 需要调用模块 B 的功能
- ✅ 模块 B 不需要调用模块 A 的功能
- ✅ 大部分业务场景（90%+）

### 标准写法

#### 1. 定义服务接口

```go
// internal/system/role/domain/service.go
package domain

// RoleService 角色服务接口
type RoleService interface {
    Create(req *CreateRoleRequest) (*Role, error)
    GetByID(id uint) (*Role, error)
    GetRolesByUserID(userID uint) ([]*Role, error)  // 跨模块方法
}
```

```go
// internal/system/user/domain/service.go
package domain

// UserService 用户服务接口
type UserService interface {
    Create(req *CreateUserRequest) (*User, error)
    GetByID(id uint) (*UserResponse, error)
    GetUserWithRoles(id uint) (*UserWithRoles, error)  // 跨模块方法
}
```

#### 2. 实现服务（依赖接口）

```go
// internal/system/user/service/service.go
package service

import (
    "nvo-api/internal/system/user/domain"
    roleDomain "nvo-api/internal/system/role/domain"  // ✅ 导入接口
    "gorm.io/gorm"
)

type UserService struct {
    db          *gorm.DB
    repo        *repository.UserRepository
    roleService roleDomain.RoleService  // ✅ 依赖接口，不是实现
}

// 构造函数：显式声明依赖
func NewUserService(
    db *gorm.DB,
    roleService roleDomain.RoleService,  // ✅ 接口参数
) domain.UserService {
    return &UserService{
        db:          db,
        repo:        repository.NewUserRepository(db),
        roleService: roleService,
    }
}

// 跨模块调用
func (s *UserService) GetUserWithRoles(id uint) (*domain.UserWithRoles, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // ✅ 调用 RoleService 接口方法
    roles, err := s.roleService.GetRolesByUserID(user.ID)
    if err != nil {
        return nil, err
    }
    
    return &domain.UserWithRoles{
        UserResponse: user,
        Roles:        roles,
    }, nil
}
```

#### 3. 模块注册

```go
// internal/system/user/module.go
package user

import (
    "nvo-api/core"
    "nvo-api/internal/system/user/domain"
    "nvo-api/internal/system/user/service"
)

type Module struct {
    service domain.UserService
}

func NewModule(pocket *core.Pocket) *Module {
    // ✅ 从 Pocket 获取依赖的服务
    userService := service.NewUserService(
        pocket.DB,
        pocket.System.Role,  // ✅ 通过 SystemService 获取
    )
    
    return &Module{service: userService}
}

func (m *Module) Service() domain.UserService {
    return m.service
}
```

#### 4. 系统初始化（三阶段）

```go
// internal/system/index.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：初始化无依赖模块
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    
    // 阶段 2：创建 SystemService
    p.System = internal.NewSystemService(
        nil,  // userService 稍后注入
        roleModule.Service(),
        permModule.Service(),
    )
    
    // 阶段 3：初始化有依赖模块
    userModule := user.NewModule(p)  // ✅ 此时 p.System.Role 已可用
    p.System.User = userModule.Service()
    
    // 继续迁移和路由注册...
}
```

### 依赖关系图

```
User ──────> Role（User 依赖 RoleService 接口）
             ↑
             │
        不依赖 User
```

### 优势

- ✅ 简单直观，易于理解
- ✅ 依赖关系清晰
- ✅ 无需额外处理
- ✅ 适用于大部分场景

---

## 双向依赖（循环依赖）

### 适用场景

- ⚠️ 模块 A 需要调用模块 B 的功能
- ⚠️ 模块 B 也需要调用模块 A 的功能
- ⚠️ 少数特殊业务场景（< 10%）

### 标准写法

#### 1. 定义服务接口（与单向依赖相同）

```go
// internal/system/role/domain/service.go
package domain

type RoleService interface {
    Create(req *CreateRoleRequest) (*Role, error)
    GetByID(id uint) (*Role, error)
    GetRolesByUserID(userID uint) ([]*Role, error)
    ValidateRolePermissions(roleID uint) error  // 可能需要调用 User
}
```

```go
// internal/system/user/domain/service.go
package domain

type UserService interface {
    Create(req *CreateUserRequest) (*User, error)
    GetByID(id uint) (*UserResponse, error)
    GetUserWithRoles(id uint) (*UserWithRoles, error)  // 需要调用 Role
}
```

#### 2. 实现服务（双向依赖接口）

```go
// internal/system/user/service/service.go
package service

import (
    roleDomain "nvo-api/internal/system/role/domain"  // ✅ 导入 role 接口
)

type UserService struct {
    db          *gorm.DB
    roleService roleDomain.RoleService  // ✅ 依赖 RoleService
}

func NewUserService(
    db *gorm.DB,
    roleService roleDomain.RoleService,
) domain.UserService {
    return &UserService{
        db:          db,
        roleService: roleService,
    }
}
```

```go
// internal/system/role/service/service.go
package service

import (
    userDomain "nvo-api/internal/system/user/domain"  // ✅ 导入 user 接口
)

type RoleService struct {
    db          *gorm.DB
    userService userDomain.UserService  // ✅ 依赖 UserService
}

func NewRoleService(
    db *gorm.DB,
    userService userDomain.UserService,
) domain.RoleService {
    return &RoleService{
        db:          db,
        userService: userService,
    }
}

// 使用 UserService
func (s *RoleService) ValidateRolePermissions(roleID uint) error {
    // ✅ 安全检查
    if s.userService == nil {
        return errors.New("userService not initialized")
    }
    
    // ✅ 调用 UserService
    users, err := s.userService.GetUsersByRoleID(roleID)
    // ... 业务逻辑
}
```

#### 3. 模块注册（支持 nil 参数）

```go
// internal/system/role/module.go
package role

func NewModule(pocket *core.Pocket, userService userDomain.UserService) *Module {
    // ✅ 接受 userService 参数（可以为 nil）
    roleService := service.NewRoleService(
        pocket.DB,
        userService,  // ✅ 可能为 nil
    )
    
    return &Module{service: roleService}
}
```

```go
// internal/system/user/module.go
package user

func NewModule(pocket *core.Pocket) *Module {
    // ✅ 从 SystemService 获取 roleService
    userService := service.NewUserService(
        pocket.DB,
        pocket.System.Role,
    )
    
    return &Module{service: userService}
}
```

#### 4. 系统初始化（五阶段）

```go
// internal/system/index.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：初始化无依赖模块
    permModule := permission.NewModule(p)
    
    // 阶段 2：初始化循环依赖模块（先传 nil）
    roleModule := role.NewModule(p, nil)  // ✅ userService 传 nil
    
    // 阶段 3：创建 SystemService
    p.System = internal.NewSystemService(
        nil,  // userService 稍后注入
        roleModule.Service(),
        permModule.Service(),
    )
    
    // 阶段 4：初始化 user 模块
    userModule := user.NewModule(p)  // ✅ 使用 p.System.Role
    p.System.User = userModule.Service()
    
    // 阶段 5：重新创建 role 模块并注入 userService
    roleModule = role.NewModule(p, userModule.Service())  // ✅ 传入真实 userService
    p.System.Role = roleModule.Service()  // ✅ 更新 SystemService
    
    // 继续迁移和路由注册...
}
```

### 依赖关系图

```
User ←──────→ Role（双向依赖）
  │            │
  └─ 依赖接口 ─┘
```

### 关键点

- ✅ 两个模块都依赖对方的**接口**，不是实现
- ✅ domain 包之间不互相导入
- ✅ 通过五阶段初始化解决注入顺序
- ✅ 使用时做 nil 检查

---

## 完整示例

### 场景：用户和订单的双向依赖

假设我们要实现：
- User 需要获取自己的订单列表（User → Order）
- Order 需要验证用户状态（Order → User）

#### 1. 定义接口

```go
// internal/business/order/domain/service.go
package domain

type OrderService interface {
    Create(req *CreateOrderRequest) (*Order, error)
    GetByID(id uint) (*Order, error)
    GetOrdersByUserID(userID uint) ([]*Order, error)
}
```

```go
// internal/system/user/domain/service.go
package domain

type UserService interface {
    Create(req *CreateUserRequest) (*User, error)
    GetByID(id uint) (*UserResponse, error)
    ValidateUserStatus(userID uint) error  // Order 需要调用
}
```

#### 2. 实现服务

```go
// internal/system/user/service/service.go
package service

import (
    orderDomain "nvo-api/internal/business/order/domain"
)

type UserService struct {
    db           *gorm.DB
    orderService orderDomain.OrderService  // ✅ 依赖 OrderService
}

func NewUserService(
    db *gorm.DB,
    orderService orderDomain.OrderService,
) domain.UserService {
    return &UserService{
        db:           db,
        orderService: orderService,
    }
}

func (s *UserService) GetUserOrders(userID uint) ([]*orderDomain.Order, error) {
    if s.orderService == nil {
        return nil, errors.New("orderService not initialized")
    }
    
    return s.orderService.GetOrdersByUserID(userID)
}
```

```go
// internal/business/order/service/service.go
package service

import (
    userDomain "nvo-api/internal/system/user/domain"
)

type OrderService struct {
    db          *gorm.DB
    userService userDomain.UserService  // ✅ 依赖 UserService
}

func NewOrderService(
    db *gorm.DB,
    userService userDomain.UserService,
) domain.OrderService {
    return &OrderService{
        db:          db,
        userService: userService,
    }
}

func (s *OrderService) Create(req *domain.CreateOrderRequest) (*domain.Order, error) {
    // ✅ 验证用户状态
    if s.userService != nil {
        if err := s.userService.ValidateUserStatus(req.UserID); err != nil {
            return nil, err
        }
    }
    
    // 创建订单...
}
```

#### 3. 初始化

```go
// internal/index.go
func RegisterAllModules(r *gin.RouterGroup, p *core.Pocket) {
    // 系统模块
    RegisterSystemModules(r.Group("/system"), p)
    
    // 业务模块（有循环依赖）
    RegisterBusinessModules(r.Group("/business"), p)
}

func RegisterBusinessModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：创建 order 模块（userService 传 nil）
    orderModule := order.NewModule(p, nil)
    
    // 阶段 2：创建 BusinessService
    p.Business = internal.NewBusinessService(
        orderModule.Service(),
    )
    
    // 阶段 3：重新创建 order 模块（注入 userService）
    orderModule = order.NewModule(p, p.System.User)
    p.Business.Order = orderModule.Service()
    
    // 继续...
}
```

---

## 常见问题

### Q1: 什么时候使用单向依赖？

**A:** 大部分场景（90%+）都应该使用单向依赖。只有当两个模块真正需要互相调用时，才考虑双向依赖。

### Q2: 为什么不会有循环导入？

**A:** 因为我们遵循依赖倒置原则：
- `user/service` 导入 `role/domain`（接口）
- `role/service` 导入 `user/domain`（接口）
- `user/domain` 和 `role/domain` 之间**不互相导入**

### Q3: 如何判断是否需要双向依赖？

**A:** 问自己三个问题：
1. 模块 A 是否真的需要调用模块 B？
2. 模块 B 是否真的需要调用模块 A？
3. 能否通过重构避免双向依赖？

如果三个问题的答案都是"是"，才使用双向依赖。

### Q4: nil 检查是必须的吗？

**A:** 是的！在双向依赖场景中，服务可能在某个阶段为 nil，必须做检查：

```go
func (s *RoleService) SomeMethod() error {
    if s.userService == nil {
        return errors.New("userService not initialized")
    }
    // 使用 userService...
}
```

### Q5: 如何测试有依赖的服务？

**A:** 使用 mock 接口：

```go
// user_service_test.go
type MockRoleService struct{}

func (m *MockRoleService) GetRolesByUserID(userID uint) ([]*roleDomain.Role, error) {
    return []*roleDomain.Role{{ID: 1, Name: "Admin"}}, nil
}

func TestUserService_GetUserWithRoles(t *testing.T) {
    mockRoleService := &MockRoleService{}
    userService := NewUserService(db, mockRoleService)
    
    // 测试...
}
```

---

## 最佳实践总结

### ✅ 推荐做法

1. **优先使用单向依赖**
   - 简单、清晰、易维护

2. **接口定义在 domain 包**
   - 每个模块的 domain 包定义自己的接口
   - domain 包之间不互相导入

3. **实现类依赖接口**
   - service 包依赖其他模块的 domain 接口
   - 不依赖其他模块的 service 实现

4. **显式声明依赖**
   - 构造函数明确列出所有依赖
   - 不隐藏依赖关系

5. **做好 nil 检查**
   - 双向依赖场景必须检查 nil
   - 提供友好的错误信息

### ❌ 避免做法

1. **不要直接依赖实现类**
   ```go
   // ❌ 错误
   import "nvo-api/internal/system/role/service"
   type UserService struct {
       roleService *service.RoleService  // 依赖实现
   }
   ```

2. **不要在 domain 包之间互相导入**
   ```go
   // ❌ 错误
   // user/domain/service.go
   import roleDomain "nvo-api/internal/system/role/domain"
   
   // role/domain/service.go
   import userDomain "nvo-api/internal/system/user/domain"
   ```

3. **不要隐藏依赖**
   ```go
   // ❌ 错误
   func NewUserService(pocket *core.Pocket) domain.UserService {
       // 依赖不明确
   }
   
   // ✅ 正确
   func NewUserService(
       db *gorm.DB,
       roleService roleDomain.RoleService,
   ) domain.UserService {
       // 依赖清晰
   }
   ```

4. **不要过度使用双向依赖**
   - 大部分场景用单向依赖即可
   - 双向依赖增加复杂度

---

## 架构优势

通过遵循这些最佳实践，你的脚手架具备：

✅ **零循环导入** - 编译期保证  
✅ **类型安全** - 接口约束  
✅ **易于测试** - Mock 接口  
✅ **易于扩展** - 添加新模块简单  
✅ **依赖清晰** - 显式声明  
✅ **符合 SOLID** - 企业级标准  

---

## 参考资料

- [依赖倒置原则（DIP）](https://en.wikipedia.org/wiki/Dependency_inversion_principle)
- [SOLID 原则](https://en.wikipedia.org/wiki/SOLID)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

---

**最后更新：2026-03-15**
