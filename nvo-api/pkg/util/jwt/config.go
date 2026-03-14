package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config - JWT配置
type Config struct {
	Secret          string        `mapstructure:"secret"`
	Issuer          string        `mapstructure:"issuer"`
	AccessExpire    time.Duration `mapstructure:"access_expire"`
	RefreshExpire   time.Duration `mapstructure:"refresh_expire"`
	EnableRefresh   bool          `mapstructure:"enable_refresh"`
}

// UserClaims - 用户声明
type UserClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair - Token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
}