package domain

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(req *LoginRequest, ip, device string) (*LoginResponse, error)

	// Logout 用户登出
	Logout(userID uint, token string) error

	// RefreshToken 刷新 Token
	RefreshToken(req *RefreshTokenRequest) (*LoginResponse, error)

	// GetCurrentUser 获取当前用户信息
	GetCurrentUser(userID uint) (*UserProfile, error)

	// GetUserMenus 获取用户菜单
	GetUserMenus(userID uint) (interface{}, error)

	// GetCaptcha 获取验证码
	GetCaptcha() (*CaptchaResponse, error)

	// ChangePassword 修改密码
	ChangePassword(userID uint, req *ChangePasswordRequest) error

	// ForgotPassword 忘记密码 - 发送验证码
	ForgotPassword(req *ForgotPasswordRequest) error

	// ResetPassword 重置密码
	ResetPassword(req *ResetPasswordRequest) error
}
