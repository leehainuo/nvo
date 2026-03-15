package role

import (
	"nvo-api/core"
	"nvo-api/internal/system/role/api"
	"nvo-api/internal/system/role/domain"
	"nvo-api/internal/system/role/service"

	"github.com/gin-gonic/gin"
)

// Module 角色模块
type Module struct {
	pocket  *core.Pocket
	handler *api.RoleHandler
	service domain.RoleService
}

// NewModule 创建角色模块
func NewModule(pocket *core.Pocket) *Module {
	// 初始化服务
	roleService := service.NewRoleService(pocket)

	// 初始化处理器
	roleHandler := api.NewRoleHandler(roleService)

	return &Module{
		pocket:  pocket,
		handler: roleHandler,
		service: roleService,
	}
}

// Service 返回服务接口
func (m *Module) Service() domain.RoleService {
	return m.service
}

// Name 模块名称
func (m *Module) Name() string {
	return "role"
}

// Models 返回需要迁移的数据模型
func (m *Module) Models() []any {
	return []any{
		&domain.Role{},
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	roles := r.Group("/roles")
	{
		roles.POST("", m.handler.Create)
		roles.GET("", m.handler.List)
		roles.GET("/all", m.handler.GetAll)
		roles.GET("/:id", m.handler.GetByID)
		roles.PUT("/:id", m.handler.Update)
		roles.DELETE("/:id", m.handler.Delete)
		roles.POST("/:id/permissions", m.handler.AssignPermissions)
		roles.GET("/:id/permissions", m.handler.GetPermissions)
	}
}
