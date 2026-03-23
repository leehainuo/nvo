package user

import (
	"nvo-api/core"
	"nvo-api/internal/system/user/api"
	"nvo-api/internal/system/user/domain"
	"nvo-api/internal/system/user/service"

	"github.com/gin-gonic/gin"
)

// Module 用户模块
type Module struct {
	pocket  *core.Pocket
	handler *api.UserHandler
	service domain.UserService
}

// NewModule 创建用户模块
func NewModule(pocket *core.Pocket) *Module {
	// 初始化服务
	userService := service.NewUserService(
		pocket.DB,
		pocket.Enforcer,
		pocket.System.Role,
	)

	// 初始化处理器
	userHandler := api.NewUserHandler(userService)

	return &Module{
		pocket:  pocket,
		handler: userHandler,
		service: userService,
	}
}

// Service 返回服务接口（供注册到 Pocket）
func (m *Module) Service() domain.UserService {
	return m.service
}

// Name 模块名称
func (m *Module) Name() string {
	return "user"
}

// Models 返回需要迁移的数据模型
func (m *Module) Models() []any {
	return []any{
		&domain.User{},
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", m.handler.Create)
		users.GET("", m.handler.List)
		users.GET("/:id", m.handler.GetByID)
		users.PUT("/:id", m.handler.Update)
		users.DELETE("/:id", m.handler.Delete)
		users.PUT("/:id/password", m.handler.ChangePassword)
	}
}
