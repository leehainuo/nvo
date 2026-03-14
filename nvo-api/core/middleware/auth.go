package middleware

import (
	"net/http"
	"nvo-api/core/auth"
	"nvo-api/pkg/util/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func Casbin(enforcer *auth.Enforcer) gin.HandlerFunc {
	return func (c *gin.Context) {
		c.Next()
	}
}

func Jwt(jwt *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authoruzation header",
			})
			c.Abort()
			return 
		}

		// 解析
		token := strings.SplitN(authorization, " ", 2)
		if len(token) != 2 || token[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
			})
			c.Abort()
			return
		}

		// 验证
		claims, err := jwt.ParseToken(token[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}
		
		// 注入
		c.Set("claims", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("claims", claims)

		c.Next()
	}
}