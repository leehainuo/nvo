package router

import (
	"moka/pkg/config"
	"moka/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	mode := config.Conf.Server.Mode

	switch mode {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
	case "beta":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(true))
	// router.Use(gin.Logger())
	router.Use(middleware.Logger())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"mode":    mode,
			"message": "pong",
		})
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("测试 Recovery 中间件的日志格式")
	})

	return router
}
