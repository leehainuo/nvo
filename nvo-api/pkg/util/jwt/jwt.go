package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	now := time.Now()

	// 生成 Access Token
	accessClaims := UserClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.Config.AccessExpire)),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.Config.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	result := &TokenPair{
		AccessToken: accessTokenString,
		ExpiresIn:   int64(j.Config.AccessExpire.Seconds()),
	}

	// 如果启用 Refresh Token
	if j.Config.EnableRefresh {
		refreshClaims := UserClaims{
			UserID:   userID,
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    j.Config.Issuer,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(j.Config.RefreshExpire)),
			},
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		refreshTokenString, err := refreshToken.SignedString([]byte(j.Config.Secret))
		if err != nil {
			return nil, fmt.Errorf("failed to sign refresh token: %w", err)
		}
		result.RefreshToken = refreshTokenString
	}

	return result, nil
}

// ParseToken - 解析 Token
func (j *JWT) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken - 刷新 Token
func (j *JWT) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// 解析 Refresh Token
	claims, err := j.ParseToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 生成新的 Token 对
	return j.GenerateTokenPair(claims.UserID, claims.Username, claims.Roles)
}
