package api

import (
	"errors"
	"strings"

	"nvo-api/core/log"
	"nvo-api/internal/system/auth/domain"
	"nvo-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "登录请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	// 获取客户端信息
	ip := c.ClientIP()
	device := c.GetHeader("User-Agent")

	loginResp, err := h.service.Login(&req, ip, device)
	if err != nil {
		log.Error("login failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, loginResp)
}

// Logout 用户登出
// @Summary 用户登出
// @Tags 认证管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := h.getUserID(c)
	token := h.getToken(c)

	if err := h.service.Logout(userID, token); err != nil {
		log.Error("logout failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "登出成功"})
}

// RefreshToken 刷新 Token
// @Summary 刷新 Token
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body domain.RefreshTokenRequest true "刷新 Token 请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	loginResp, err := h.service.RefreshToken(&req)
	if err != nil {
		log.Error("refresh token failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, loginResp)
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags 认证管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := h.getUserID(c)

	userProfile, err := h.service.GetCurrentUser(userID)
	if err != nil {
		log.Error("get current user failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, userProfile)
}

// GetUserMenus 获取用户菜单
// @Summary 获取用户菜单
// @Tags 认证管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/menus [get]
func (h *AuthHandler) GetUserMenus(c *gin.Context) {
	userID := h.getUserID(c)

	menus, err := h.service.GetUserMenus(userID)
	if err != nil {
		log.Error("get user menus failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, menus)
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Tags 认证管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/auth/captcha [get]
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	captcha, err := h.service.GetCaptcha()
	if err != nil {
		log.Error("get captcha failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, captcha)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body domain.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := h.getUserID(c)

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.ChangePassword(userID, &req); err != nil {
		log.Error("change password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功，请重新登录"})
}

// ForgotPassword 忘记密码 - 发送验证码
// @Summary 忘记密码 - 发送验证码
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body domain.ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.ForgotPassword(&req); err != nil {
		log.Error("forgot password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "验证码已发送到邮箱"})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body domain.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.ResetPassword(&req); err != nil {
		log.Error("reset password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// getUserID 从上下文获取用户 ID
func (h *AuthHandler) getUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// getToken 从请求头获取 Token
func (h *AuthHandler) getToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
