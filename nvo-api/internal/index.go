package internal

import (
	"github.com/gin-gonic/gin"

	auditDomain "nvo-api/internal/system/audit/domain"
	authDomain  "nvo-api/internal/system/auth/domain"
	deptDomain  "nvo-api/internal/system/dept/domain"
	dictDomain  "nvo-api/internal/system/dict/domain"
	menuDomain  "nvo-api/internal/system/menu/domain"
	permDomain  "nvo-api/internal/system/permission/domain"
	roleDomain  "nvo-api/internal/system/role/domain"
	userDomain  "nvo-api/internal/system/user/domain"
)

// Module 模块接口
// 所有业务模块都应该实现此接口，以实现模块化、可插拔的架构设计
type Module interface {
	Name() string                      // Name 返回模块名称
	Models() []any                     // Models 返回模块模型
	RegisterRoutes(r *gin.RouterGroup) // RegisterRoutes 注册模块路由
}

// SystemService 系统模块服务聚合（企业级中台架构）
// 将所有系统相关的服务统一管理，提供更优雅的访问方式
type SystemService struct {
	User       userDomain.UserService       // 用户服务
	Role       roleDomain.RoleService       // 角色服务
	Permission permDomain.PermissionService // 权限服务
	Menu       menuDomain.MenuService       // 菜单服务
	Dept       deptDomain.DeptService       // 部门服务
	Audit      auditDomain.AuditService     // 审计服务
	Auth       authDomain.AuthService       // 认证服务
	Dict       dictDomain.DictService       // 字典服务
}

// NewSystemService 创建系统服务聚合
func NewSystemService(
	userService userDomain.UserService,
	roleService roleDomain.RoleService,
	permService permDomain.PermissionService,
	menuService menuDomain.MenuService,
	deptService deptDomain.DeptService,
	auditService auditDomain.AuditService,
	authService authDomain.AuthService,
	dictService dictDomain.DictService,
) *SystemService {
	return &SystemService{
		User:       userService,
		Role:       roleService,
		Permission: permService,
		Menu:       menuService,
		Dept:       deptService,
		Audit:      auditService,
		Auth:       authService,
		Dict:       dictService,
	}
}
