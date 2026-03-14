# Casbin 权限设计文档

## 概述

基于 Casbin 实现的 **三级权限控制系统**：
1. **API 权限** - 控制接口访问
2. **按钮权限** - 控制页面按钮显示和操作
3. **菜单权限** - 控制菜单显示

## 权限模型

### RBAC 模型定义

```
[request_definition]
r = sub, obj, act, type

[policy_definition]
p = sub, obj, act, type

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.type == p.type
```

### 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `sub` | Subject（主体） | `user:123`, `role:admin` |
| `obj` | Object（对象） | `/api/v1/users`, `user.delete`, `system.user` |
| `act` | Action（操作） | `GET`, `POST`, `click`, `view` |
| `type` | Type（权限类型） | `api`, `button`, `menu` |

## 三级权限详解

### 1. API 权限

**控制接口访问权限**

#### 策略示例

```csv
# 角色权限
p, role:admin, /api/v1/users, GET, api
p, role:admin, /api/v1/users, POST, api
p, role:admin, /api/v1/users/:id, DELETE, api

p, role:user, /api/v1/users, GET, api
p, role:user, /api/v1/profile, GET, api
p, role:user, /api/v1/profile, PUT, api

# 用户角色
g, user:123, role:admin
g, user:456, role:user
```

#### 使用方式

```go
// 在中间件中检查
ok, _ := enforcer.CheckAPI("user:123", "/api/v1/users", "GET")
if !ok {
    c.JSON(403, gin.H{"message": "无权访问"})
    return
}

// 或使用中间件
router.Use(auth.APIAuthMiddleware(enforcer, logger))
```

#### 支持通配符

```csv
# 管理员对所有 API 有权限
p, role:admin, /api/v1/*, *, api

# 用户对自己的资源有权限
p, user:123, /api/v1/users/123/*, *, api
```

### 2. 按钮权限

**控制页面按钮的显示和操作**

#### 策略示例

```csv
# 用户管理页面的按钮权限
p, role:admin, user.create, click, button
p, role:admin, user.edit, click, button
p, role:admin, user.delete, click, button
p, role:admin, user.export, click, button

p, role:user_manager, user.create, click, button
p, role:user_manager, user.edit, click, button

p, role:viewer, user.view, click, button
```

#### 按钮编码规范

```
{模块}.{操作}

示例：
- user.create    - 创建用户按钮
- user.edit      - 编辑用户按钮
- user.delete    - 删除用户按钮
- order.approve  - 订单审批按钮
- order.export   - 导出订单按钮
```

#### 前端使用

```javascript
// 获取用户可用按钮
GET /api/v1/permissions/buttons?page=user

// 响应
{
  "buttons": [
    { "code": "user.create", "name": "新建", "type": "primary" },
    { "code": "user.edit", "name": "编辑", "type": "default" },
    { "code": "user.delete", "name": "删除", "type": "danger" }
  ]
}

// 前端根据按钮权限渲染
<button v-if="hasButton('user.delete')">删除</button>
```

#### 后端检查

```go
// 在 API 中二次检查按钮权限
func DeleteUser(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // 检查删除按钮权限
    ok, _ := enforcer.CheckButton("user:"+userID, "user.delete", "click")
    if !ok {
        c.JSON(403, gin.H{"message": "无删除权限"})
        return
    }
    
    // 执行删除逻辑
}
```

### 3. 菜单权限

**控制菜单的显示**

#### 策略示例

```csv
# 系统管理菜单
p, role:admin, system, view, menu
p, role:admin, system.user, view, menu
p, role:admin, system.role, view, menu
p, role:admin, system.permission, view, menu

# 业务菜单
p, role:user, dashboard, view, menu
p, role:user, profile, view, menu

p, role:user_manager, user, view, menu
p, role:user_manager, user.list, view, menu
```

#### 菜单编码规范

```
{一级菜单}.{二级菜单}.{三级菜单}

示例：
- system              - 系统管理
- system.user         - 用户管理
- system.role         - 角色管理
- business            - 业务管理
- business.order      - 订单管理
- business.order.list - 订单列表
```

#### 前端使用

```javascript
// 获取用户可访问的菜单
GET /api/v1/permissions/menus

// 响应
{
  "menus": [
    {
      "code": "system",
      "name": "系统管理",
      "icon": "setting",
      "children": [
        { "code": "system.user", "name": "用户管理", "path": "/system/user" },
        { "code": "system.role", "name": "角色管理", "path": "/system/role" }
      ]
    }
  ]
}

// 前端根据菜单权限渲染导航
```

## 使用示例

### 1. 初始化 Enforcer

```go
// 在 Pocket 中初始化
func (p *PocketBuilder) initEnforcer() error {
    enforcer, err := auth.NewEnforcer(p.pocket.DB, p.pocket.Logger)
    if err != nil {
        return err
    }
    p.pocket.Enforcer = enforcer
    return nil
}
```

### 2. 初始化权限数据

```go
func InitPermissions(enforcer *auth.Enforcer) error {
    // 1. 创建角色
    roles := []string{"admin", "user_manager", "user"}
    
    // 2. 添加 API 权限
    enforcer.AddPolicy("role:admin", "/api/v1/*", "*", auth.PermissionAPI)
    enforcer.AddPolicy("role:user", "/api/v1/profile", "GET", auth.PermissionAPI)
    
    // 3. 添加按钮权限
    enforcer.AddPolicy("role:admin", "user.delete", "click", auth.PermissionButton)
    enforcer.AddPolicy("role:user_manager", "user.edit", "click", auth.PermissionButton)
    
    // 4. 添加菜单权限
    enforcer.AddPolicy("role:admin", "system", "view", auth.PermissionMenu)
    enforcer.AddPolicy("role:admin", "system.user", "view", auth.PermissionMenu)
    
    // 5. 分配角色
    enforcer.AddRoleForUser("user:1", "role:admin")
    enforcer.AddRoleForUser("user:2", "role:user_manager")
    
    return nil
}
```

### 3. API 中使用

```go
// 用户管理 API
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup, enforcer *auth.Enforcer, logger *zap.Logger) {
    // 使用 API 权限中间件
    users := r.Group("/users")
    users.Use(auth.APIAuthMiddleware(enforcer, logger))
    {
        users.GET("", h.ListUsers)
        users.POST("", h.CreateUser)
        users.DELETE("/:id", h.DeleteUser)
    }
}

// 在 Handler 中二次检查按钮权限
func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // 检查删除按钮权限
    ok, _ := h.enforcer.CheckButton("user:"+userID, "user.delete", "click")
    if !ok {
        c.JSON(403, gin.H{"message": "无删除权限"})
        return
    }
    
    // 执行删除
    id := c.Param("id")
    if err := h.service.DeleteUser(id); err != nil {
        c.JSON(500, gin.H{"message": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "删除成功"})
}
```

### 4. 获取用户权限

```go
// 获取用户菜单
func GetUserMenus(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // 所有菜单
    allMenus := []auth.Menu{
        {Code: "system", Name: "系统管理", Icon: "setting"},
        {Code: "system.user", Name: "用户管理", Path: "/system/user"},
        {Code: "system.role", Name: "角色管理", Path: "/system/role"},
    }
    
    // 过滤用户可访问的菜单
    menus, _ := auth.GetUserMenus(enforcer, userID, allMenus)
    
    c.JSON(200, gin.H{"menus": menus})
}

// 获取页面按钮
func GetPageButtons(c *gin.Context) {
    userID := c.GetString("user_id")
    page := c.Query("page") // 如 "user"
    
    // 该页面的所有按钮
    allButtons := []auth.Button{
        {Code: "user.create", Name: "新建", Type: "primary"},
        {Code: "user.edit", Name: "编辑", Type: "default"},
        {Code: "user.delete", Name: "删除", Type: "danger"},
    }
    
    // 过滤用户可用的按钮
    buttons, _ := auth.GetUserButtons(enforcer, userID, allButtons)
    
    c.JSON(200, gin.H{"buttons": buttons})
}
```

## 权限管理 API

### 角色管理

```go
// 创建角色
POST /api/v1/roles
{
  "name": "产品经理",
  "code": "product_manager"
}

// 为角色分配权限
POST /api/v1/roles/:id/permissions
{
  "permissions": [
    { "object": "/api/v1/products", "action": "GET", "type": "api" },
    { "object": "product.edit", "action": "click", "type": "button" },
    { "object": "product", "action": "view", "type": "menu" }
  ]
}
```

### 用户管理

```go
// 为用户分配角色
POST /api/v1/users/:id/roles
{
  "roles": ["role:admin", "role:user_manager"]
}

// 获取用户权限
GET /api/v1/users/:id/permissions
{
  "roles": ["admin", "user_manager"],
  "permissions": {
    "apis": [...],
    "buttons": [...],
    "menus": [...]
  }
}
```

## 最佳实践

### 1. 权限粒度

- **API 权限**：粗粒度，按路由分组
- **按钮权限**：中粒度，按功能分组
- **菜单权限**：粗粒度，按模块分组

### 2. 命名规范

```
API:    /api/v1/{模块}/{资源}
按钮:   {模块}.{操作}
菜单:   {一级}.{二级}.{三级}
```

### 3. 角色设计

```
超级管理员 (super_admin)  - 所有权限
管理员 (admin)            - 系统管理权限
业务管理员 (manager)      - 业务管理权限
普通用户 (user)           - 基础权限
访客 (guest)              - 只读权限
```

### 4. 性能优化

```go
// 使用缓存的 Enforcer
enforcer := casbin.NewSyncedEnforcer(model, adapter)

// 批量检查权限
results := enforcer.BatchEnforce(requests)

// 定期刷新策略
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        enforcer.LoadPolicy()
    }
}()
```

## 总结

这个权限系统实现了：

✅ **三级权限控制** - API、按钮、菜单
✅ **RBAC 模型** - 基于角色的访问控制
✅ **灵活扩展** - 支持自定义权限类型
✅ **持久化存储** - 使用 GORM Adapter
✅ **优雅的 API** - 简洁易用的接口
✅ **完整的日志** - 记录所有权限检查

**这是一个生产级别的权限系统设计！** 🎉
