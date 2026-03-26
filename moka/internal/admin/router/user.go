package router

import (
	"moka/internal/admin/handler"
	"moka/internal/admin/service"
	"moka/pkg/auth/casbin"
	"moka/pkg/client/mysql"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(group *gin.RouterGroup) {
	// 依赖注入:
	userService := service.NewUserService(mysql.Client(), casbin.Enforcer())
	userHandler := handler.NewUserHandler(userService)

	router := group.Group("/user")
	{
		router.GET("", userHandler.Demo)
	}
}