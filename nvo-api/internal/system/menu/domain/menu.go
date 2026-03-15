package domain

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ParentID  uint           `gorm:"index;default:0" json:"parent_id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Path      string         `gorm:"size:200" json:"path"`
	Component string         `gorm:"size:200" json:"component"`
	Icon      string         `gorm:"size:50" json:"icon"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Type      int8           `gorm:"default:1" json:"type"`
	Visible   int8           `gorm:"default:1" json:"visible"`
	Status    int8           `gorm:"default:1" json:"status"`
	Perms     string         `gorm:"size:100" json:"perms"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Menu) TableName() string {
	return "sys_menus"
}

type CreateMenuRequest struct {
	ParentID  uint   `json:"parent_id"`
	Name      string `json:"name" binding:"required,max=50"`
	Path      string `json:"path" binding:"max=200"`
	Component string `json:"component" binding:"max=200"`
	Icon      string `json:"icon" binding:"max=50"`
	Sort      int    `json:"sort"`
	Type      int8   `json:"type" binding:"required,oneof=1 2"`
	Visible   int8   `json:"visible" binding:"oneof=0 1"`
	Status    int8   `json:"status" binding:"oneof=0 1"`
	Perms     string `json:"perms" binding:"max=100"`
}

type UpdateMenuRequest struct {
	ParentID  *uint  `json:"parent_id"`
	Name      string `json:"name" binding:"max=50"`
	Path      string `json:"path" binding:"max=200"`
	Component string `json:"component" binding:"max=200"`
	Icon      string `json:"icon" binding:"max=50"`
	Sort      *int   `json:"sort"`
	Type      *int8  `json:"type" binding:"omitempty,oneof=1 2"`
	Visible   *int8  `json:"visible" binding:"omitempty,oneof=0 1"`
	Status    *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Perms     string `json:"perms" binding:"max=100"`
}

type MenuTree struct {
	*Menu
	Children []*MenuTree `json:"children,omitempty"`
}
