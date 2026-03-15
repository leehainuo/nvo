# 🎯 简化后的服务注入模式

## 核心理念

**零心智负担** - 注册流程简单直接，无需关心依赖顺序

## 注册流程（仅 3 步）

```go
// internal/system/registry.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 1. 初始化所有模块
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    userModule := user.NewModule(p)

    // 2. 聚合为 SystemService 并注册到 Pocket
    p.System = domain.NewSystemService(
        userModule.Service(),
        roleModule.Service(),
        permModule.Service(),
    )

    // 3. 数据库迁移和路由注册
    modules := []internal.Module{roleModule, permModule, userModule}
    migrateModels(p.DB, modules)
    for _, module := range modules {
        module.RegisterRoutes(r)
    }
}
```

## 为什么这样简单？

### 关键设计：延迟访问

服务在**初始化时**不访问其他服务，只在**运行时**才访问：

```go
// ✅ 初始化时 - 不访问 pocket.System
func NewUserService(pocket *core.Pocket) domain.UserService {
    return &UserService{
        pocket:   pocket,  // 只保存引用
        repo:     repository.NewUserRepository(pocket.DB),
        enforcer: pocket.Enforcer,
        db:       pocket.DB,
    }
}

// ✅ 运行时 - 才访问 pocket.System
func (s *UserService) GetUserWithRoles(id uint) (*UserWithRoles, error) {
    // 此时 pocket.System 已经注册完成
    if s.pocket.System == nil || s.pocket.System.Role == nil {
        return &UserWithRoles{...}, nil
    }
    
    roles := s.pocket.System.Role.GetRolesByUserID(id)
    // ...
}
```

### 时间线

```
时刻 1: roleModule := role.NewModule(p)
        ↓ RoleService 初始化（不访问其他服务）
        
时刻 2: userModule := user.NewModule(p)
        ↓ UserService 初始化（不访问 pocket.System）
        
时刻 3: p.System = NewSystemService(...)
        ↓ SystemService 注册到 Pocket
        
时刻 4: 用户请求 GET /users/1/roles
        ↓ UserService.GetUserWithRoles() 运行
        ↓ 此时访问 pocket.System.Role（已存在）✅
```

## 添加新模块

### 无依赖模块

```go
// 直接添加到初始化列表
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    userModule := user.NewModule(p)
    deptModule := dept.NewModule(p)  // ✅ 新增部门模块
    
    p.System = domain.NewSystemService(
        userModule.Service(),
        roleModule.Service(),
        permModule.Service(),
        deptModule.Service(),  // ✅ 添加到聚合
    )
    
    // ...
}
```

### 有依赖模块

如果新模块在**运行时**需要访问其他服务：

```go
// 新模块的服务实现
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*Order, error) {
    // ✅ 运行时访问 - 完全没问题
    user := s.pocket.System.User.GetByID(req.UserID)
    
    // 业务逻辑...
}
```

**无需修改注册流程！** 只要在运行时访问，就不会有问题。

## 对比：优化前 vs 优化后

### 优化前（5 个阶段，心智负担重）

```go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：初始化基础模块
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    
    // 阶段 2：临时注册（❌ 心智负担）
    tempSystem := &domain.SystemService{
        Role:       roleModule.Service(),
        Permission: permModule.Service(),
    }
    p.System = tempSystem
    
    // 阶段 3：初始化业务模块
    userModule := user.NewModule(p)
    
    // 阶段 4：重新注册完整的 SystemService（❌ 重复）
    p.System = domain.NewSystemService(
        userModule.Service(),
        roleModule.Service(),
        permModule.Service(),
    )
    
    // 阶段 5：数据库迁移
    // 阶段 6：路由注册
}
```

### 优化后（3 步，清晰简洁）

```go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 1. 初始化所有模块
    roleModule := role.NewModule(p)
    permModule := permission.NewModule(p)
    userModule := user.NewModule(p)

    // 2. 聚合并注册
    p.System = domain.NewSystemService(
        userModule.Service(),
        roleModule.Service(),
        permModule.Service(),
    )

    // 3. 迁移和路由
    modules := []internal.Module{roleModule, permModule, userModule}
    migrateModels(p.DB, modules)
    for _, m := range modules { m.RegisterRoutes(r) }
}
```

## 核心原则

### ✅ DO（推荐）

1. **服务初始化时**只访问基础设施（DB, Redis, Enforcer）
2. **运行时**才访问其他业务服务（通过 pocket.System）
3. 访问其他服务时做 nil 检查
4. 保持注册流程简单直接

### ❌ DON'T（避免）

1. ~~在 `NewXxxService()` 中访问 `pocket.System`~~
2. ~~创建临时的服务注册~~
3. ~~复杂的多阶段初始化~~
4. ~~依赖顺序的心智负担~~

## 实战示例

### 场景：添加订单模块（依赖用户服务）

```go
// 1. 实现订单服务
type OrderService struct {
    pocket *core.Pocket
    repo   *repository.OrderRepository
}

func NewOrderService(pocket *core.Pocket) domain.OrderService {
    return &OrderService{
        pocket: pocket,  // ✅ 只保存引用
        repo:   repository.NewOrderRepository(pocket.DB),
    }
}

func (s *OrderService) Create(req *CreateOrderRequest) (*Order, error) {
    // ✅ 运行时访问用户服务
    if s.pocket.System == nil || s.pocket.System.User == nil {
        return nil, errors.New("user service not available")
    }
    
    user, err := s.pocket.System.User.GetByID(req.UserID)
    if err != nil {
        return nil, err
    }
    
    // 创建订单...
}

// 2. 注册模块 - 无需修改注册流程！
func RegisterBusinessModules(r *gin.RouterGroup, p *core.Pocket) {
    orderModule := order.NewModule(p)  // ✅ 直接初始化
    
    p.Business = domain.NewBusinessService(
        orderModule.Service(),
    )
    
    // 迁移和路由...
}
```

## 总结

| 特性 | 优化前 | 优化后 |
|------|--------|--------|
| **注册步骤** | 5-6 步 | 3 步 ✅ |
| **心智负担** | 高（需理解依赖顺序） | 低（无需关心顺序） ✅ |
| **代码行数** | ~50 行 | ~20 行 ✅ |
| **可维护性** | 中（添加模块需修改多处） | 高（添加模块只需一处） ✅ |
| **出错概率** | 高（临时注册易出错） | 低（流程简单） ✅ |

**核心思想：延迟访问 = 零心智负担** 🎉
