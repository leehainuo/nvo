package auth

import (
	"nvo-api/core"
	"nvo-api/internal/system/auth/api"
	"nvo-api/internal/system/auth/domain"
	"nvo-api/internal/system/auth/service"

	"github.com/gin-gonic/gin"
)

type Module struct {
	pocket  *core.Pocket
	handler *api.AuthHandler
	service domain.AuthService
}

func NewModule(pocket *core.Pocket) *Module {
	authService := service.NewAuthService(pocket)
	authHandler := api.NewAuthHandler(authService)

	return &Module{
		pocket:  pocket,
		handler: authHandler,
		service: authService,
	}
}

func (m *Module) Service() domain.AuthService {
	return m.service
}

func (m *Module) Name() string {
	return "auth"
}

func (m *Module) Models() []any {
	return []any{
		&domain.LoginLog{},
		&domain.PasswordResetCode{},
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", m.handler.Login)
		auth.POST("/refresh", m.handler.RefreshToken)
		auth.GET("/captcha", m.handler.GetCaptcha)
		auth.POST("/forgot-password", m.handler.ForgotPassword)
		auth.POST("/reset-password", m.handler.ResetPassword)
		auth.POST("/logout", m.handler.Logout)
		auth.GET("/me", m.handler.GetCurrentUser)
		auth.GET("/menus", m.handler.GetUserMenus)
		auth.PUT("/password", m.handler.ChangePassword)
	}
}
