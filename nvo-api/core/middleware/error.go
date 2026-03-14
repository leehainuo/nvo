package middleware

import (
	"net/http"
	"nvo-api/core/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Error - 统一错误处理中间件
func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 记录日志
			log.Error("unhandled error",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(err),
			)

			// 返回未知错误
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "internal server error",
				})
			}
		}
	}
}
