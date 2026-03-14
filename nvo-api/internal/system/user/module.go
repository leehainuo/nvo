package user

import (
	"nvo-api/core"
	"nvo-api/core/log"
	"nvo-api/internal/system/user/api"
	"nvo-api/internal/system/user/domain"
	"nvo-api/internal/system/user/service"
	"nvo-api/internal/system/user/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Module 用户模块
type Module struct {
	pocket  *core.Pocket
	handler *api.UserHandler
}

// NewModule 创建用户模块
func NewModule(pocket *core.Pocket) *Module {
	userRepo := repository.NewUserRepository(pocket.DB)
	userService := service.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)

	if err := pocket.DB.AutoMigrate(&domain.User{}); err != nil {
		log.Error("Failed to migrate user table", zap.Error(err))
	}

	return &Module{
		pocket:  pocket,
		handler: userHandler,
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.RegisterRoutes(r)
}

// Name 模块名称
func (m *Module) Name() string {
	return "user"
}
