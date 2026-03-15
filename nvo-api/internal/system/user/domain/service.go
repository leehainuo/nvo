package domain

// UserService 用户领域服务接口
type UserService interface {
	// 基础 CRUD
	Create(req *CreateUserRequest) (*User, error)
	GetByID(id uint) (*UserResponse, error)
	Update(id uint, req *UpdateUserRequest) error
	Delete(id uint) error
	List(req *ListUserRequest) ([]*UserResponse, int64, error)

	// 业务方法
	ChangePassword(id uint, oldPassword, newPassword string) error

	// 跨模块业务方法（可扩展）
	GetUserWithRoles(id uint) (*UserWithRoles, error)
	AssignRoles(userID uint, roleIDs []uint) error
}

// UserWithRoles 用户及其角色（聚合根）
type UserWithRoles struct {
	*UserResponse
	RoleDetails []*RoleDetail `json:"role_details"`
}

// RoleDetail 角色详情
type RoleDetail struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
