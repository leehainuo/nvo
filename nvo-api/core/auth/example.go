package auth

import (
	"github.com/casbin/casbin/v3"
	"gorm.io/gorm"
)

// InitDefaultPermissions 初始化默认权限数据（示例）
func InitDefaultPermissions(enforcer *casbin.SyncedEnforcer) error {
	// ==================== 1. 定义角色 ====================
	// 注意：超级管理员 (role:super_admin) 不需要配置权限，在权限检查时直接放行

	// ==================== 2. API 权限 ====================
	// 用户管理员 - 用户相关 API
	enforcer.AddPolicy("role:user_manager", "/api/v1/users", "GET", string(PermissionAPI))
	enforcer.AddPolicy("role:user_manager", "/api/v1/users", "POST", string(PermissionAPI))
	enforcer.AddPolicy("role:user_manager", "/api/v1/users/:id", "GET", string(PermissionAPI))
	enforcer.AddPolicy("role:user_manager", "/api/v1/users/:id", "PUT", string(PermissionAPI))
	enforcer.AddPolicy("role:user_manager", "/api/v1/users/:id", "DELETE", string(PermissionAPI))

	// 普通用户 - 只能访问自己的信息
	enforcer.AddPolicy("role:user", "/api/v1/profile", "GET", string(PermissionAPI))
	enforcer.AddPolicy("role:user", "/api/v1/profile", "PUT", string(PermissionAPI))

	// ==================== 3. 按钮权限 ====================
	// 命名规范：模块:资源:操作 (例如 sys:user:create)

	// 管理员 - 所有按钮
	enforcer.AddPolicy("role:admin", "sys:user:create", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:user:edit", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:user:delete", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:user:export", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:user:import", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:role:create", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:role:edit", "click", string(PermissionButton))
	enforcer.AddPolicy("role:admin", "sys:role:delete", "click", string(PermissionButton))

	// 用户管理员 - 部分按钮
	enforcer.AddPolicy("role:user_manager", "sys:user:create", "click", string(PermissionButton))
	enforcer.AddPolicy("role:user_manager", "sys:user:edit", "click", string(PermissionButton))
	enforcer.AddPolicy("role:user_manager", "sys:user:export", "click", string(PermissionButton))

	// 普通用户 - 只能查看
	enforcer.AddPolicy("role:user", "sys:user:view", "click", string(PermissionButton))

	// ==================== 4. 菜单权限 ====================
	// 命名规范：使用路径格式，与前端路由保持一致

	// 管理员 - 所有菜单
	enforcer.AddPolicy("role:admin", "/dashboard", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:admin", "/system", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:admin", "/system/user", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:admin", "/system/role", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:admin", "/system/permission", "view", string(PermissionMenu))

	// 用户管理员 - 部分菜单
	enforcer.AddPolicy("role:user_manager", "/dashboard", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:user_manager", "/system", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:user_manager", "/system/user", "view", string(PermissionMenu))

	// 普通用户 - 基础菜单
	enforcer.AddPolicy("role:user", "/dashboard", "view", string(PermissionMenu))
	enforcer.AddPolicy("role:user", "/profile", "view", string(PermissionMenu))

	// ==================== 5. 分配角色给用户（示例） ====================
	// 这部分通常在用户注册或管理员分配时动态添加
	// enforcer.AddRoleForUser("user:1", "role:super_admin")  // 超级管理员，权限检查时直接放行
	// enforcer.AddRoleForUser("user:2", "role:user_manager")
	// enforcer.AddRoleForUser("user:3", "role:user")

	return nil
}

// MigratePermissionTables 迁移权限相关表
func MigratePermissionTables(db *gorm.DB) error {
	// Casbin 会自动创建 casbin_rule 表
	// 这里可以创建自定义的权限管理表

	type Role struct {
		ID          uint   `gorm:"primaryKey"`
		Code        string `gorm:"uniqueIndex;size:50;not null"`
		Name        string `gorm:"size:100;not null"`
		Description string `gorm:"size:255"`
	}

	type Permission struct {
		ID          uint   `gorm:"primaryKey"`
		Type        string `gorm:"size:20;not null"` // api, button, menu
		Object      string `gorm:"size:255;not null"`
		Action      string `gorm:"size:50;not null"`
		Description string `gorm:"size:255"`
	}

	return db.AutoMigrate(&Role{}, &Permission{})
}
