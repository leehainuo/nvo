package middleware

import (
	"moka/pkg/auth/casbin"
	"moka/pkg/util/jwt"
	"moka/pkg/util/log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func JWTAuth(whitelist ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查白名单 是否放行
		if isInWhitelist(c.Request.URL.Path, whitelist) {
			log.Info("path in whitelist, skip jwt auth")
			c.Next()
			return
		}

		// 获取 Authorization 头
		header := c.GetHeader("Authorization")
		if header == "" {
			log.Warn("missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("invalid authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "认证令牌格式错误",
			})
		}

		tokenString := parts[1]
		
		// 验证 Token
		claims, err := jwt.ValidateAccessToken(tokenString)
		if err != nil {
			log.Warn("invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 注入信息
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("token_id", claims.TokenID)
		c.Set("claims", claims)

		c.Next()
	}
}

func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		path   := c.Request.URL.Path
		method := c.Request.Method

		// 放行
		is, err := casbin.IsSuperAdmin(userID)
		if err != nil {
	    	c.JSON(http.StatusInternalServerError, gin.H{
        		"code":    http.StatusInternalServerError,
        		"message": "权限检查失败",
    		})
    		c.Abort()
    		return
		}

		if is {
			c.Next()
			return
		}

		// 鉴权
		ok, err := casbin.CheckAPI(path, method, userID)
		if err != nil || !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isInWhitelist(path string, whitelist []string) bool {
    for _, pattern := range whitelist {
        // 精确匹配
        if path == pattern {
            return true
        }
 
        // 前缀匹配（支持通配符 /*）
        if prefix, ok := strings.CutSuffix(pattern, "/*"); ok {
            if strings.HasPrefix(path, prefix) {
                return true
            }
        }
 
        // 后缀匹配（支持通配符 *）
        if suffix, ok := strings.CutPrefix(pattern, "*"); ok {
            if strings.HasSuffix(path, suffix) {
                return true
            }
        }
    }
    return false
}