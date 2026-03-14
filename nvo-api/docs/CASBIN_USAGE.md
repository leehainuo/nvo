# Casbin 权限系统使用指南

## ✨ 已完成的优雅设计

我已经为你设计并实现了一个**完整的三级权限控制系统**：

### 1. 核心文件

- **`core/auth/casbin.go`** - Casbin Enforcer 封装，提供优雅的 API
- **`core/auth/middleware.go`** - API 权限中间件和辅助函数
- **`core/auth/example.go`** - 默认权限初始化示例
- **`docs/CASBIN_DESIGN.md`** - 完整的设计文档

### 2. 三级权限

#### API 权限
```go
// 检查用户是否可以访问接口
ok, _ := enforcer.CheckAPI("user:123", "/api/v1/users", "GET")
```

#### 按钮权限
```go
// 检查用户是否可以点击删除按钮
ok, _ := enforcer.CheckButton("user:123", "user.delete", "click")
```

#### 菜单权限
```go
// 检查用户是否可以查看菜单
ok, _ := enforcer.CheckMenu("user:123", "system.user", "view")
```

### 3. 在 Pocket 中使用

```go
// 启用权限控制
pocket := core.NewPocketBuilder("config.yml").
    WithEnforcer().  // 启用 Casbin
    MustBuild()

// 使用 Enforcer
enforcer := pocket.Enforcer

// 检查权限
ok, _ := enforcer.CheckAPI("user:123", "/api/v1/users", "GET")
```

### 4. 权限模型

```
[request_definition]
r = sub, obj, act, type

参数说明：
- sub:  主体 (user:123, role:admin)
- obj:  对象 (/api/v1/users, user.delete, system.user)
- act:  操作 (GET, POST, click, view)
- type: 类型 (api, button, menu)
```

### 5. 使用示例

#### 初始化权限
```go
// 添加角色权限
enforcer.AddPolicy("role:admin", "/api/v1/*", "*", auth.PermissionAPI)
enforcer.AddPolicy("role:admin", "user.delete", "click", auth.PermissionButton)
enforcer.AddPolicy("role:admin", "system", "view", auth.PermissionMenu)

// 为用户分配角色
enforcer.AddRoleForUser("user:1", "role:admin")
```

#### API 中间件
```go
// 在路由中使用
router.Use(auth.APIAuthMiddleware(enforcer, logger))
```

#### 获取用户菜单
```go
menus, _ := auth.GetUserMenus(enforcer, "123", allMenus)
```

#### 获取用户按钮
```go
buttons, _ := auth.GetUserButtons(enforcer, "123", pageButtons)
```

## 📚 完整文档

详细设计和使用说明请查看：
- **`docs/CASBIN_DESIGN.md`** - 完整的设计文档，包含所有示例

## 🎯 特点

✅ **三级权限控制** - API、按钮、菜单
✅ **RBAC 模型** - 基于角色的访问控制
✅ **持久化存储** - 使用 GORM Adapter
✅ **优雅的 API** - 简洁易用
✅ **完整的日志** - 记录所有权限检查
✅ **中间件支持** - 开箱即用的 Gin 中间件

## 🚀 快速开始

```go
// 1. 启用 Enforcer
pocket := core.NewPocketBuilder("config.yml").
    WithEnforcer().
    MustBuild()

// 2. 初始化默认权限（自动完成）
// auth.InitDefaultPermissions() 已在 Pocket 初始化时调用

// 3. 使用权限检查
if ok, _ := pocket.Enforcer.CheckAPI("user:1", "/api/v1/users", "GET"); ok {
    // 有权限
}
```

这是一个**生产级别的权限系统**，可以直接用于实际项目！🎉
