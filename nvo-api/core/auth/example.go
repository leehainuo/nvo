package auth

import (
	"gorm.io/gorm"
)

// InitDefaultPermissions 初始化默认权限数据（示例）
func InitDefaultPermissions(enforcer *Enforcer) error {
	// ==================== 1. 定义角色 ====================
	
	// ==================== 2. API 权限 ====================
	// 管理员 - 所有 API 权限
	enforcer.AddPolicy("role:admin", "/api/v1/*", "*", PermissionAPI)
	
	// 用户管理员 - 用户相关 API
	enforcer.AddPolicy("role:user_manager", "/api/v1/users", "GET", PermissionAPI)
	enforcer.AddPolicy("role:user_manager", "/api/v1/users", "POST", PermissionAPI)
	enforcer.AddPolicy("role:user_manager", "/api/v1/users/:id", "PUT", PermissionAPI)
	enforcer.AddPolicy("role:user_manager", "/api/v1/users/:id", "DELETE", PermissionAPI)
	
	// 普通用户 - 只能访问自己的信息
	enforcer.AddPolicy("role:user", "/api/v1/profile", "GET", PermissionAPI)
	enforcer.AddPolicy("role:user", "/api/v1/profile", "PUT", PermissionAPI)
	
	// ==================== 3. 按钮权限 ====================
	// 管理员 - 所有按钮
	enforcer.AddPolicy("role:admin", "user.create", "click", PermissionButton)
	enforcer.AddPolicy("role:admin", "user.edit", "click", PermissionButton)
	enforcer.AddPolicy("role:admin", "user.delete", "click", PermissionButton)
	enforcer.AddPolicy("role:admin", "user.export", "click", PermissionButton)
	enforcer.AddPolicy("role:admin", "user.import", "click", PermissionButton)
	
	// 用户管理员 - 部分按钮
	enforcer.AddPolicy("role:user_manager", "user.create", "click", PermissionButton)
	enforcer.AddPolicy("role:user_manager", "user.edit", "click", PermissionButton)
	enforcer.AddPolicy("role:user_manager", "user.export", "click", PermissionButton)
	
	// 普通用户 - 只能查看
	enforcer.AddPolicy("role:user", "user.view", "click", PermissionButton)
	
	// ==================== 4. 菜单权限 ====================
	// 管理员 - 所有菜单
	enforcer.AddPolicy("role:admin", "dashboard", "view", PermissionMenu)
	enforcer.AddPolicy("role:admin", "system", "view", PermissionMenu)
	enforcer.AddPolicy("role:admin", "system.user", "view", PermissionMenu)
	enforcer.AddPolicy("role:admin", "system.role", "view", PermissionMenu)
	enforcer.AddPolicy("role:admin", "system.permission", "view", PermissionMenu)
	
	// 用户管理员 - 部分菜单
	enforcer.AddPolicy("role:user_manager", "dashboard", "view", PermissionMenu)
	enforcer.AddPolicy("role:user_manager", "system", "view", PermissionMenu)
	enforcer.AddPolicy("role:user_manager", "system.user", "view", PermissionMenu)
	
	// 普通用户 - 基础菜单
	enforcer.AddPolicy("role:user", "dashboard", "view", PermissionMenu)
	enforcer.AddPolicy("role:user", "profile", "view", PermissionMenu)
	
	// ==================== 5. 分配角色给用户（示例） ====================
	// 这部分通常在用户注册或管理员分配时动态添加
	// enforcer.AddRoleForUser("user:1", "role:admin")
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
		ID          uint           `gorm:"primaryKey"`
		Type        string         `gorm:"size:20;not null"` // api, button, menu
		Object      string         `gorm:"size:255;not null"`
		Action      string         `gorm:"size:50;not null"`
		Description string         `gorm:"size:255"`
	}
	
	return db.AutoMigrate(&Role{}, &Permission{})
}
