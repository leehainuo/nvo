package router

import (
	"moka/pkg/apperr"
	"moka/pkg/config"
	"moka/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	mode := config.Conf().Server.Mode

	switch mode {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	case "beta":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	apperr.Init(mode != "prod")

	router := gin.New()

	// router.Use(gin.Recovery())
	// router.Use(gin.Logger())
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(true))
	router.Use(middleware.Logger())
	router.Use(middleware.Error())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"mode":    mode,
			"message": "pong",
		})
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("测试 Recovery 中间件的日志格式")
	})

	group := router.Group("/api/v1")
	{
		InitUserRouter(group)
		// InitXxxRouter ...
	}

	return router
}
