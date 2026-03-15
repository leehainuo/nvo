package domain

import (
	"time"

	"gorm.io/gorm"
)

type Dept struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ParentID  uint           `gorm:"index;default:0" json:"parent_id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Leader    string         `gorm:"size:50" json:"leader"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Email     string         `gorm:"size:100" json:"email"`
	Status    int8           `gorm:"default:1" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Dept) TableName() string {
	return "sys_depts"
}

type CreateDeptRequest struct {
	ParentID uint   `json:"parent_id"`
	Name     string `json:"name" binding:"required,max=50"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"omitempty,email"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateDeptRequest struct {
	ParentID *uint  `json:"parent_id"`
	Name     string `json:"name" binding:"max=50"`
	Sort     *int   `json:"sort"`
	Leader   string `json:"leader" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"omitempty,email"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type DeptTree struct {
	*Dept
	Children []*DeptTree `json:"children,omitempty"`
}
