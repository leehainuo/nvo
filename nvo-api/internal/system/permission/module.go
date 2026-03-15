package permission

import (
	"nvo-api/core"
	"nvo-api/internal/system/permission/api"
	"nvo-api/internal/system/permission/domain"
	"nvo-api/internal/system/permission/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Module 权限模块
type Module struct {
	pocket  *core.Pocket
	handler *api.PermissionHandler
	service domain.PermissionService
}

// NewModule 创建权限模块
func NewModule(pocket *core.Pocket) *Module {
	// 初始化服务
	permService := service.NewPermissionService(pocket)

	// 初始化处理器
	permHandler := api.NewPermissionHandler(pocket.Enforcer, zap.L())

	return &Module{
		pocket:  pocket,
		handler: permHandler,
		service: permService,
	}
}

// Service 返回服务接口
func (m *Module) Service() domain.PermissionService {
	return m.service
}

// Name 模块名称
func (m *Module) Name() string {
	return "permission"
}

// Models 返回需要迁移的数据模型
// Permission 模块不需要数据库表，返回空切片
func (m *Module) Models() []any {
	return []any{}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	permissions := r.Group("/permissions")
	{
		permissions.GET("/menus", m.handler.GetUserMenus)
		permissions.GET("/buttons", m.handler.GetUserButtons)
		permissions.GET("/user", m.handler.GetUserPermissions)
	}
}
