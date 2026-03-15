package middleware

import (
	"net/http"
	"nvo-api/pkg/util/jwt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

const SuperAdminRole = "role:super_admin"

func Casbin(enforcer *casbin.SyncedEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息（由 JWT 中间件设置）
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		subject := "user:" + userID.(string)

		// 检查是否是超级管理员
		roles, err := enforcer.GetRolesForUser(subject)
		if err == nil {
			for _, role := range roles {
				if role == SuperAdminRole {
					// 超管直接放行
					c.Next()
					return
				}
			}
		}

		// 普通用户进行权限检查
		path := c.Request.URL.Path
		method := c.Request.Method

		ok, err := enforcer.Enforce(subject, path, method, "api")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "permission check failed",
			})
			c.Abort()
			return
		}

		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "permission denied",
			})
			c.Abort()
			return
		}

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
