package repository

import (
	"nvo-api/internal/system/auth/domain"
	"time"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateLoginLog 创建登录日志
func (r *AuthRepository) CreateLoginLog(log *domain.LoginLog) error {
	return r.db.Create(log).Error
}

// CreatePasswordResetCode 创建密码重置验证码
func (r *AuthRepository) CreatePasswordResetCode(code *domain.PasswordResetCode) error {
	return r.db.Create(code).Error
}

// GetValidPasswordResetCode 获取有效的密码重置验证码
func (r *AuthRepository) GetValidPasswordResetCode(email, code string) (*domain.PasswordResetCode, error) {
	var resetCode domain.PasswordResetCode
	err := r.db.Where("email = ? AND code = ? AND used = ? AND expires_at > ?",
		email, code, false, time.Now()).
		First(&resetCode).Error
	return &resetCode, err
}

// MarkPasswordResetCodeAsUsed 标记验证码为已使用
func (r *AuthRepository) MarkPasswordResetCodeAsUsed(id uint) error {
	return r.db.Model(&domain.PasswordResetCode{}).
		Where("id = ?", id).
		Update("used", true).Error
}

// DeleteExpiredPasswordResetCodes 删除过期的验证码
func (r *AuthRepository) DeleteExpiredPasswordResetCodes() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&domain.PasswordResetCode{}).Error
}
