package domain

import "time"

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         *UserProfile `json:"user"`
}

// UserProfile 用户信息
type UserProfile struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
	ExpiresIn    int    `json:"expires_in"`
}

// LoginLog 登录日志
type LoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:50;index" json:"username"`
	IP        string    `gorm:"size:50" json:"ip"`
	Location  string    `gorm:"size:100" json:"location"`
	Device    string    `gorm:"size:200" json:"device"`
	Status    string    `gorm:"size:20;index" json:"status"` // success, failed
	Message   string    `gorm:"size:500" json:"message"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// TableName 表名
func (LoginLog) TableName() string {
	return "sys_login_logs"
}

// PasswordResetCode 密码重置验证码
type PasswordResetCode struct {
	ID        uint      `gorm:"primarykey"`
	Email     string    `gorm:"size:100;index"`
	Code      string    `gorm:"size:10"`
	ExpiresAt time.Time `gorm:"index"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// TableName 表名
func (PasswordResetCode) TableName() string {
	return "sys_password_reset_codes"
}
