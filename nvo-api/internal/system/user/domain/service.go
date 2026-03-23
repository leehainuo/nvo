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

	AssignRoles(userID uint, roleIDs []uint) error
}
