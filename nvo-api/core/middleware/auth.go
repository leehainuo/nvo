package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"nvo-api/core/log"
	"nvo-api/pkg/util/jwt"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWTAuth JWT 认证中间件（支持白名单）
func JWTAuth(jwtUtil *jwt.JWT, whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否在白名单中
		if isInWhitelist(c.Request.URL.Path, whitelist) {
			log.Debug("path in whitelist, skip jwt auth", zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}

		// 1. 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("missing authorization header", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 2. 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("invalid authorization format", zap.String("header", authHeader))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1002,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. 验证 Token
		claims, err := jwtUtil.ParseToken(tokenString)
		if err != nil {
			log.Warn("invalid token",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1003,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 4. 解析用户 ID
		userID, err := strconv.ParseUint(claims.UserID, 10, 32)
		if err != nil {
			log.Error("invalid user_id in token", zap.String("user_id", claims.UserID))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1004,
				"message": "认证令牌数据异常",
			})
			c.Abort()
			return
		}

		// 5. 将用户信息注入到上下文
		c.Set("user_id", uint(userID))
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("claims", claims)

		log.Debug("jwt auth success",
			zap.Uint("user_id", uint(userID)),
			zap.String("username", claims.Username),
			zap.Strings("roles", claims.Roles))

		c.Next()
	}
}

// CasbinAuth Casbin 权限认证中间件（支持白名单）
func CasbinAuth(enforcer *casbin.SyncedEnforcer, whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否在白名单中
		if isInWhitelist(c.Request.URL.Path, whitelist) {
			log.Debug("path in whitelist, skip casbin auth", zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}

		// 1. 获取用户信息（由 JWT 中间件注入）
		userID, exists := c.Get("user_id")
		if !exists {
			log.Warn("user_id not found in context", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		// 2. 构建 Casbin subject
		subject := fmt.Sprintf("user:%d", userID.(uint))
		path := c.Request.URL.Path
		method := c.Request.Method

		// 3. 检查是否是超级管理员
		roles, err := enforcer.GetRolesForUser(subject)
		if err != nil {
			log.Error("failed to get user roles",
				zap.Error(err),
				zap.String("subject", subject))
		} else {
			// 超级管理员拥有所有权限
			for _, role := range roles {
				if role == "role:1" || role == "admin" {
					log.Debug("super admin access granted",
						zap.String("subject", subject),
						zap.String("path", path),
						zap.String("method", method))
					c.Next()
					return
				}
			}
		}

		// 4. 普通用户权限检查
		ok, err := enforcer.Enforce(subject, path, method, "api")
		if err != nil {
			log.Error("casbin enforce failed",
				zap.Error(err),
				zap.String("subject", subject),
				zap.String("path", path),
				zap.String("method", method))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    2002,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}

		if !ok {
			log.Warn("permission denied",
				zap.String("subject", subject),
				zap.String("path", path),
				zap.String("method", method))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    2001,
				"message": "无权访问此资源",
			})
			c.Abort()
			return
		}

		log.Debug("casbin auth success",
			zap.String("subject", subject),
			zap.String("path", path),
			zap.String("method", method))

		c.Next()
	}
}

// isInWhitelist 检查路径是否在白名单中（公共方法）
func isInWhitelist(path string, whitelist []string) bool {
	for _, pattern := range whitelist {
		// 精确匹配
		if path == pattern {
			return true
		}

		// 前缀匹配（支持通配符 /*）
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}

		// 后缀匹配（支持通配符 *）
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(path, suffix) {
				return true
			}
		}
	}
	return false
}
