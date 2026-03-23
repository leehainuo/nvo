package domain

import (
	"time"

	"gorm.io/gorm"
)

// User 用户领域模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:50" json:"nickname"`
	Email     string         `gorm:"uniqueIndex;size:100" json:"email"`
	Phone     string         `gorm:"size:20" json:"phone"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Status    int8           `gorm:"default:1;comment:状态 1-正常 0-禁用" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "sys_users"
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	RoleIDs  []uint `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Avatar   string `json:"avatar"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	RoleIDs  []uint `json:"role_ids"`
}

// ListUserRequest 用户列表请求
type ListUserRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Username string `form:"username"`
	Nickname string `form:"nickname"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    int8      `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse 将 User 转换为 UserResponse
func (u *User) ToResponse(roles []string) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Phone:     u.Phone,
		Avatar:    u.Avatar,
		Status:    u.Status,
		Roles:     roles,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToResponseList 批量转换 User 列表为 UserResponse 列表
func ToResponseList(users []*User, rolesMap map[uint][]string) []*UserResponse {
	responses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, user.ToResponse(rolesMap[user.ID]))
	}
	return responses
}
