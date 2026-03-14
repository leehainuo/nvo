package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取
		id := c.GetHeader("X-Request-ID")

		// 如果没有 则自动生成
		if id == "" {
			id = uuid.New().String()
		}

		// 设置到上下文和响应头
		c.Set("X-Request-ID", id)
		c.Header("X-Request-ID", id)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get("X-Request-ID"); exists {
		return id.(string)
	}
	return ""
}