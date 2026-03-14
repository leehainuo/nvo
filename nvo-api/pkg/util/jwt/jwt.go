package jwt

import (
	"fmt"

	"github.com/spf13/viper"
)

type JWT struct {
	Config Config
}

// InitJWT - 创建JWT实例
func InitJWT(c Config) *JWT {
	return &JWT{
		Config: c,
	}
}

// InitFromViper - 从 Viper 初始化
func InitFromViper(v *viper.Viper, key string) (*JWT, error) {
	var c Config
	if err := v.UnmarshalKey(key, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jwt config: %w", err)
	}
	return InitJWT(c), nil
}

// GenerateTokenPair - 生成 Token 对
func (j *JWT) GenerateTokenPair(userID, username string, roles []string) (*TokenPair, error) {
	// TODO: 实现 token 生成逻辑
	return nil, nil
}

// ParseToken - 解析 Token
func (j *JWT) ParseToken(token string) (*UserClaims, error) {
	// TODO: 实现 token 解析逻辑
	return nil, nil
}

// RefreshToken - 刷新 Token
func (j *JWT) RefreshToken(token string) (*TokenPair, error) {
	// TODO: 实现 token 刷新逻辑
	return nil, nil
}
