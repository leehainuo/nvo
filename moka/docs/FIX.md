基于你的 Casbin 模型和企业级 RBAC 需求，我将提供一个优雅的实现方案，覆盖 **API、菜单、按钮** 三个维度的细粒度权限控制。

基于你的 Casbin 模型，我将提供一个企业级 RBAC 的完整实现方案，优雅地实现 **API、菜单、按钮** 三维度的细粒度权限控制。

## 一、核心设计思路

你的 Casbin 模型已经很好地支持了多维度权限：
- `sub`: 主体（用户/角色，如 `user:1001` 或 `role:admin`）
- `obj`: 资源对象（API路径、菜单ID、按钮ID）
- `act`: 操作（GET/POST/DELETE 或 view/edit/delete）
- `type`: 权限类型（api/menu/button）

## 二、数据库设计

### 1. 权限资源表设计

```sql
-- 资源表（统一管理 API、菜单、按钮）
CREATE TABLE `permissions` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `type` varchar(20) NOT NULL COMMENT '权限类型: api/menu/button',
  `code` varchar(100) NOT NULL COMMENT '权限编码: order:view',
  `name` varchar(100) NOT NULL COMMENT '权限名称',
  `resource` varchar(255) NOT NULL COMMENT '资源标识: /api/orders 或 menu:order 或 btn:order:export',
  `action` varchar(50) NOT NULL COMMENT '操作: GET/POST/view/edit/delete',
  `parent_id` bigint DEFAULT 0 COMMENT '父级ID（菜单层级）',
  `sort` int DEFAULT 0,
  `status` tinyint DEFAULT 1 COMMENT '状态: 1启用 0禁用',
  `created_at` datetime,
  `updated_at` datetime,
  UNIQUE KEY `uk_resource_action_type` (`resource`, `action`, `type`)
) COMMENT='权限资源表';

-- 角色表
CREATE TABLE `roles` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `code` varchar(50) NOT NULL UNIQUE COMMENT '角色编码: admin/finance_manager',
  `name` varchar(100) NOT NULL COMMENT '角色名称',
  `parent_id` bigint DEFAULT 0 COMMENT '父角色ID（支持角色继承）',
  `level` int DEFAULT 1 COMMENT '角色层级',
  `status` tinyint DEFAULT 1,
  `created_at` datetime,
  `updated_at` datetime
) COMMENT='角色表';

-- 用户角色关联表
CREATE TABLE `user_roles` (
  `user_id` bigint NOT NULL,
  `role_id` bigint NOT NULL,
  `expire_at` datetime DEFAULT NULL COMMENT '过期时间（临时角色）',
  PRIMARY KEY (`user_id`, `role_id`)
) COMMENT='用户角色关联表';

-- 数据权限表（可选，用于数据范围控制）
CREATE TABLE `data_scopes` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `role_id` bigint NOT NULL,
  `scope_type` varchar(20) NOT NULL COMMENT '数据范围: all/dept/self/custom',
  `scope_value` text COMMENT '自定义范围（JSON）',
  UNIQUE KEY `uk_role_scope` (`role_id`)
) COMMENT='数据权限表';
```

## 三、Casbin 策略示例

### 权限策略格式

```
# 格式: p, sub, obj, act, type

# API 权限
p, role:admin, /api/orders, GET, api
p, role:admin, /api/orders/:id, POST, api
p, role:finance, /api/reports/*, GET, api

# 菜单权限
p, role:admin, menu:system, view, menu
p, role:admin, menu:system:user, view, menu
p, role:operator, menu:order, view, menu

# 按钮权限
p, role:admin, btn:order:export, click, button
p, role:admin, btn:order:delete, click, button
p, role:operator, btn:order:view, click, button

# 角色继承
g, user:1001, role:admin
g, user:1002, role:finance
g, role:finance, role:operator  # 角色继承：finance 继承 operator 的权限
```

## 四、Moka 项目实现方案

### 1. 增强 Casbin 初始化

```go
// moka/pkg/auth/casbin.go
package auth

import (
	"fmt"
	"moka/pkg/util/log"
	
	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type PermissionType string

const (
	PermissionAPI    PermissionType = "api"
	PermissionMenu   PermissionType = "menu"
	PermissionButton PermissionType = "button"
)

type Enforcer struct {
	*casbin.SyncedEnforcer
}

func NewEnforcer(db *gorm.DB) (*Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	m := model.NewModel()
	m.LoadModelFromText(`
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
	`)

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	log.Info("Casbin enforcer initialized successfully")

	return &Enforcer{SyncedEnforcer: enforcer}, nil
}

// CheckPermission 统一权限校验
func (e *Enforcer) CheckPermission(userID string, resource, action string, permType PermissionType) (bool, error) {
	subject := fmt.Sprintf("user:%s", userID)
	return e.Enforce(subject, resource, action, string(permType))
}

// CheckAPI 校验 API 权限
func (e *Enforcer) CheckAPI(userID, path, method string) (bool, error) {
	return e.CheckPermission(userID, path, method, PermissionAPI)
}

// CheckMenu 校验菜单权限
func (e *Enforcer) CheckMenu(userID, menuCode string) (bool, error) {
	return e.CheckPermission(userID, menuCode, "view", PermissionMenu)
}

// CheckButton 校验按钮权限
func (e *Enforcer) CheckButton(userID, buttonCode string) (bool, error) {
	return e.CheckPermission(userID, buttonCode, "click", PermissionButton)
}

// GetUserMenus 获取用户可访问的菜单列表
func (e *Enforcer) GetUserMenus(userID string) ([]string, error) {
	subject := fmt.Sprintf("user:%s", userID)
	
	// 获取所有策略
	policies := e.GetFilteredPolicy(0, subject)
	
	menus := make([]string, 0)
	for _, policy := range policies {
		if len(policy) >= 4 && policy[3] == string(PermissionMenu) {
			menus = append(menus, policy[1]) // obj
		}
	}
	
	return menus, nil
}

// GetUserButtons 获取用户可访问的按钮列表
func (e *Enforcer) GetUserButtons(userID string) ([]string, error) {
	subject := fmt.Sprintf("user:%s", userID)
	
	policies := e.GetFilteredPolicy(0, subject)
	
	buttons := make([]string, 0)
	for _, policy := range policies {
		if len(policy) >= 4 && policy[3] == string(PermissionButton) {
			buttons = append(buttons, policy[1])
		}
	}
	
	return buttons, nil
}

// IsSuperAdmin 检查是否是超级管理员
func (e *Enforcer) IsSuperAdmin(userID string) bool {
	subject := fmt.Sprintf("user:%s", userID)
	roles, _ := e.GetRolesForUser(subject)
	
	for _, role := range roles {
		if role == "role:admin" || role == "role:super_admin" {
			return true
		}
	}
	return false
}
```

### 2. 增强权限中间件

```go
// moka/pkg/middleware/casbin.go
package middleware

import (
	"fmt"
	"moka/pkg/auth"
	"moka/pkg/util/log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CasbinAuth Casbin 权限中间件（支持白名单）
func CasbinAuth(enforcer *auth.Enforcer, whitelist ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单检查
		if isInWhitelist(c.Request.URL.Path, whitelist) {
			log.Info("path in whitelist, skip casbin auth", zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}

		// 获取用户信息（由 JWT 中间件注入）
		userID, exists := c.Get("user_id")
		if !exists {
			log.Warn("user_id not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		userIDStr := fmt.Sprintf("%v", userID)
		path := c.Request.URL.Path
		method := c.Request.Method

		// 超级管理员放行
		if enforcer.IsSuperAdmin(userIDStr) {
			log.Info("super admin access granted", 
				zap.String("user_id", userIDStr),
				zap.String("path", path))
			c.Next()
			return
		}

		// API 权限校验
		ok, err := enforcer.CheckAPI(userIDStr, path, method)
		if err != nil {
			log.Error("casbin enforce failed", 
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("path", path))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}

		if !ok {
			log.Warn("permission denied",
				zap.String("user_id", userIDStr),
				zap.String("path", path),
				zap.String("method", method))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "无权访问此资源",
			})
			c.Abort()
			return
		}

		log.Info("casbin auth success",
			zap.String("user_id", userIDStr),
			zap.String("path", path))

		c.Next()
	}
}

// RequireButton 按钮权限装饰器
func RequireButton(enforcer *auth.Enforcer, buttonCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		userIDStr := fmt.Sprintf("%v", userID)

		// 超级管理员放行
		if enforcer.IsSuperAdmin(userIDStr) {
			c.Next()
			return
		}

		ok, err := enforcer.CheckButton(userIDStr, buttonCode)
		if err != nil || !ok {
			log.Warn("button permission denied",
				zap.String("user_id", userIDStr),
				zap.String("button", buttonCode))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "无权限操作此按钮",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

### 3. 权限管理 Handler

```go
// moka/internal/handler/permission.go
package handler

import (
	"moka/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	enforcer *auth.Enforcer
}

func NewPermissionHandler(enforcer *auth.Enforcer) *PermissionHandler {
	return &PermissionHandler{enforcer: enforcer}
}

// GetUserMenus 获取用户菜单权限
func (h *PermissionHandler) GetUserMenus(c *gin.Context) {
	userID := c.GetString("user_id")
	
	menus, err := h.enforcer.GetUserMenus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取菜单权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    menus,
	})
}

// GetUserButtons 获取用户按钮权限
func (h *PermissionHandler) GetUserButtons(c *gin.Context) {
	userID := c.GetString("user_id")
	
	buttons, err := h.enforcer.GetUserButtons(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取按钮权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    buttons,
	})
}

// CheckPermission 检查权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req struct {
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
		Type     string `json:"type" binding:"required,oneof=api menu button"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "参数错误",
		})
		return
	}

	userID := c.GetString("user_id")
	
	ok, err := h.enforcer.CheckPermission(userID, req.Resource, req.Action, auth.PermissionType(req.Type))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "权限检查失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"has_permission": ok,
		},
	})
}
```

### 4. 路由配置示例

```go
// moka/cmd/api/main.go 或路由配置文件
func setupRoutes(r *gin.Engine, enforcer *auth.Enforcer) {
	// 公开路由
	public := r.Group("/api")
	{
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/refresh", authHandler.RefreshToken)
	}

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(middleware.JWTAuth("/api/auth/*"))
	api.Use(middleware.CasbinAuth(enforcer, "/api/user/profile"))
	{
		// 用户权限查询
		api.GET("/permissions/menus", permHandler.GetUserMenus)
		api.GET("/permissions/buttons", permHandler.GetUserButtons)
		api.POST("/permissions/check", permHandler.CheckPermission)

		// 订单管理（需要按钮权限）
		orders := api.Group("/orders")
		{
			orders.GET("", orderHandler.List)
			orders.GET("/:id", orderHandler.Get)
			orders.POST("", 
				middleware.RequireButton(enforcer, "btn:order:create"),
				orderHandler.Create)
			orders.DELETE("/:id",
				middleware.RequireButton(enforcer, "btn:order:delete"),
				orderHandler.Delete)
			orders.POST("/:id/export",
				middleware.RequireButton(enforcer, "btn:order:export"),
				orderHandler.Export)
		}
	}
}
```

## 五、前端集成示例

### 1. 获取用户权限

```typescript
// 登录后获取权限
async function loadUserPermissions() {
  const [menus, buttons] = await Promise.all([
    api.get('/api/permissions/menus'),
    api.get('/api/permissions/buttons')
  ]);
  
  // 存储到状态管理
  store.commit('setMenus', menus);
  store.commit('setButtons', buttons);
}
```

### 2. 菜单权限控制

```vue
<template>
  <el-menu>
    <el-menu-item v-if="hasMenu('menu:order')" index="/orders">
      订单管理
    </el-menu-item>
    <el-menu-item v-if="hasMenu('menu:system')" index="/system">
      系统管理
    </el-menu-item>
  </el-menu>
</template>

<script>
export default {
  methods: {
    hasMenu(menuCode) {
      return this.$store.state.menus.includes(menuCode);
    }
  }
}
</script>
```

### 3. 按钮权限控制

```vue
<template>
  <div>
    <el-button v-if="hasButton('btn:order:create')" @click="create">
      新建订单
    </el-button>
    <el-button v-if="hasButton('btn:order:export')" @click="exportData">
      导出
    </el-button>
    <el-button v-if="hasButton('btn:order:delete')" @click="deleteOrder">
      删除
    </el-button>
  </div>
</template>

<script>
export default {
  methods: {
    hasButton(buttonCode) {
      return this.$store.state.buttons.includes(buttonCode);
    }
  }
}
</script>
```

## 六、初始化权限数据示例

```go
// 初始化脚本
func InitPermissions(enforcer *auth.Enforcer) error {
	// 1. 添加角色
	enforcer.AddGroupingPolicy("role:finance", "role:operator") // 角色继承
	
	// 2. 添加 API 权限
	enforcer.AddPolicy("role:admin", "/api/orders", "GET", "api")
	enforcer.AddPolicy("role:admin", "/api/orders", "POST", "api")
	enforcer.AddPolicy("role:admin", "/api/orders/:id", "DELETE", "api")
	enforcer.AddPolicy("role:operator", "/api/orders", "GET", "api")
	
	// 3. 添加菜单权限
	enforcer.AddPolicy("role:admin", "menu:system", "view", "menu")
	enforcer.AddPolicy("role:admin", "menu:system:user", "view", "menu")
	enforcer.AddPolicy("role:operator", "menu:order", "view", "menu")
	
	// 4. 添加按钮权限
	enforcer.AddPolicy("role:admin", "btn:order:create", "click", "button")
	enforcer.AddPolicy("role:admin", "btn:order:delete", "click", "button")
	enforcer.AddPolicy("role:admin", "btn:order:export", "click", "button")
	enforcer.AddPolicy("role:operator", "btn:order:view", "click", "button")
	
	// 5. 分配用户角色
	enforcer.AddGroupingPolicy("user:1001", "role:admin")
	enforcer.AddGroupingPolicy("user:1002", "role:finance")
	
	return enforcer.SavePolicy()
}
```

## 七、核心优势

1. **统一模型**：API、菜单、按钮使用同一套 Casbin 模型，通过 `type` 字段区分
2. **细粒度控制**：支持路径级 API 权限、菜单层级、按钮级操作
3. **角色继承**：通过 `g` 关系实现角色层级和权限继承
4. **前后端分离**：后端提供权限数据，前端动态渲染，后端强制校验
5. **可扩展**：可轻松扩展数据权限、时间限制等高级特性

这套方案在 **Ask 模式**下提供给你参考。如需直接实现到代码中，请切换到 **Code 模式**。