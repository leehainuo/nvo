package domain

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色领域模型
type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Code        string         `gorm:"uniqueIndex;size:50;not null;comment:角色编码" json:"code"`
	Name        string         `gorm:"size:100;not null;comment:角色名称" json:"name"`
	Description string         `gorm:"size:255;comment:角色描述" json:"description"`
	Sort        int            `gorm:"default:0;comment:排序" json:"sort"`
	Status      int8           `gorm:"default:1;comment:状态 1-正常 0-禁用" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "sys_roles"
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
	Status      *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// ListRoleRequest 角色列表请求
type ListRoleRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	Permissions []Permission `json:"permissions" binding:"required"`
}

// Permission 权限项
type Permission struct {
	Object string `json:"object" binding:"required"`
	Action string `json:"action" binding:"required"`
	Type   string `json:"type" binding:"required,oneof=api button menu"`
}
