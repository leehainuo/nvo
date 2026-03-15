package domain

// RoleService 角色领域服务接口
type RoleService interface {
	// 基础 CRUD
	Create(req *CreateRoleRequest) (*Role, error)
	GetByID(id uint) (*Role, error)
	Update(id uint, req *UpdateRoleRequest) error
	Delete(id uint) error
	List(req *ListRoleRequest) ([]*Role, int64, error)
	GetAll() ([]*Role, error)

	// 权限管理
	AssignPermissions(id uint, req *AssignPermissionsRequest) error
	GetPermissions(id uint) ([][]string, error)

	// 跨模块业务方法（可扩展）
	GetRolesByUserID(userID uint) ([]*Role, error)
}
