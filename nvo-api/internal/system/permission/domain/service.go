package domain

// PermissionService 权限领域服务接口
type PermissionService interface {
	// 获取用户权限
	GetUserMenus(userID uint) ([]Menu, error)
	GetUserButtons(userID uint) ([]string, error)
	GetUserPermissions(userID uint) (*UserPermissions, error)

	// 权限检查
	CheckPermission(userID uint, object, action string) (bool, error)
}

// Menu 菜单项
type Menu struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Icon     string  `json:"icon"`
	Children []*Menu `json:"children,omitempty"`
}

// UserPermissions 用户权限集合
type UserPermissions struct {
	Menus   []Menu   `json:"menus"`
	Buttons []string `json:"buttons"`
	APIs    []string `json:"apis"`
}
