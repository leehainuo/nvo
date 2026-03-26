package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"moka/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var j *Config

// Config j 配置
type Config struct {
	Secret          string        `mapstructure:"secret"`
	Issuer          string        `mapstructure:"issuer"`
	AccessExpire    time.Duration `mapstructure:"access_expire"`
	RefreshExpire   time.Duration `mapstructure:"refresh_expire"`
}

// UserClaims 用户访问令牌载荷
type UserClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	TokenID  string   `json:"token_id"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// RefreshClaims 刷新令牌载荷
type RefreshClaims struct {
    UserID  string `json:"user_id"`
    TokenID string `json:"token_id"`
    jwt.RegisteredClaims
}

// TokenPair Token 对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
}

func init() {
	var c Config
	if err := config.Viper.UnmarshalKey("jwt", &c); err != nil {
		panic(fmt.Sprintf("\033[1;31mfailed to unmarshal jwt config: %v\033[0m", err))
	}

	if c.Secret == "" {
		panic("\033[1;31mjwt secret is required\033[0m")
	}

	if c.Issuer == "" {
		c.Issuer = "lihainuo"
	}

	if c.AccessExpire == 0 {
		c.AccessExpire = 2 * time.Hour
	}

	if c.RefreshExpire == 0 {
		c.RefreshExpire = 7 * 24 * time.Hour
	}

	j = &c
}

func generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func JWT() Config {
	return *j
}

func GenerateAccessToken(userID, username string, roles []string) (string, error) {
	now := time.Now()

	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}

	claims := UserClaims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.AccessExpire)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.Secret))
}

func GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}

	claims := RefreshClaims{
        UserID:  userID,
        TokenID: tokenID,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    j.Issuer,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(j.RefreshExpire)),
        },
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.Secret))
}

func GenerateTokenPair(userID, username string, roles[]string) (*TokenPair, error) {
	access, err := GenerateAccessToken(userID, username, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, err := GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	pair := &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(j.AccessExpire.Seconds()),
	}

	return pair, nil
}

func ValidateAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func ExtractTokenID(tokenString string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &UserClaims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims.TokenID, nil
	}

	return "", jwt.ErrTokenInvalidClaims
}