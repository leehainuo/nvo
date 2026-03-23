package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"nvo-api/core"
	"nvo-api/core/log"
	"nvo-api/internal/system/auth/repository"
	"nvo-api/pkg/util/jwt"

	"github.com/casbin/casbin/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	authDomain "nvo-api/internal/system/auth/domain"
	menuDomain "nvo-api/internal/system/menu/domain"
	userDomain "nvo-api/internal/system/user/domain"
)

type AuthService struct {
	db          *gorm.DB
	jwt         *jwt.JWT
	enforcer    *casbin.SyncedEnforcer
	repo        *repository.AuthRepository
	userService userDomain.UserService
	menuService menuDomain.MenuService
}

func NewAuthService(pocket *core.Pocket) authDomain.AuthService {
	return &AuthService{
		db:          pocket.DB,
		jwt:         pocket.JWT,
		enforcer:    pocket.Enforcer,
		repo:        repository.NewAuthRepository(pocket.DB),
		userService: pocket.System.User,
		menuService: pocket.System.Menu,
	}
}

// Login 用户登录
func (s *AuthService) Login(req *authDomain.LoginRequest, ip, device string) (*authDomain.LoginResponse, error) {
	// 1. 验证验证码（暂时跳过，后续实现）
	// TODO: 验证验证码

	// 2. 查询用户
	var user userDomain.User
	err := s.db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		s.recordLoginLog(0, req.Username, ip, device, "failed", "用户名或密码错误")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 3. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordLoginLog(user.ID, req.Username, ip, device, "failed", "用户名或密码错误")
		return nil, errors.New("用户名或密码错误")
	}

	// 4. 检查用户状态
	if user.Status != 1 {
		s.recordLoginLog(user.ID, req.Username, ip, device, "failed", "账号已被禁用")
		return nil, errors.New("账号已被禁用")
	}

	// 5. 获取用户角色
	subject := fmt.Sprintf("user:%d", user.ID)
	roleCodes, _ := s.enforcer.GetRolesForUser(subject)

	// 6. 获取用户权限
	permissions := s.getUserPermissions(user.ID)

	// 7. 生成 Token
	tokenPair, err := s.jwt.GenerateTokenPair(
		fmt.Sprintf("%d", user.ID),
		user.Username,
		roleCodes,
	)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	// 8. 记录登录成功日志
	s.recordLoginLog(user.ID, req.Username, ip, device, "success", "登录成功")

	// 9. 返回登录响应
	return &authDomain.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User: &authDomain.UserProfile{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Email:       user.Email,
			Phone:       user.Phone,
			Roles:       roleCodes,
			Permissions: permissions,
		},
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(userID uint, token string) error {
	// TODO: 将 token 加入黑名单（Redis）
	log.Info("user logout", zap.Uint("user_id", userID))
	return nil
}

// RefreshToken 刷新 Token
func (s *AuthService) RefreshToken(req *authDomain.RefreshTokenRequest) (*authDomain.LoginResponse, error) {
	// 1. 刷新 Token
	tokenPair, err := s.jwt.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("刷新 Token 失败: %w", err)
	}

	// 2. 解析 Token 获取用户信息
	claims, err := s.jwt.ParseToken(tokenPair.AccessToken)
	if err != nil {
		return nil, err
	}

	// 3. 获取用户完整信息
	userProfile, err := s.GetCurrentUser(s.parseUserID(claims.UserID))
	if err != nil {
		return nil, err
	}

	return &authDomain.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         userProfile,
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(userID uint) (*authDomain.UserProfile, error) {
	// 1. 获取用户基本信息
	var user userDomain.User
	err := s.db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 2. 获取用户角色
	subject := fmt.Sprintf("user:%d", user.ID)
	roleCodes, _ := s.enforcer.GetRolesForUser(subject)

	// 3. 获取用户权限
	permissions := s.getUserPermissions(user.ID)

	return &authDomain.UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Phone:       user.Phone,
		Roles:       roleCodes,
		Permissions: permissions,
	}, nil
}

// GetUserMenus 获取用户菜单
func (s *AuthService) GetUserMenus(userID uint) (interface{}, error) {
	// TODO: 根据用户权限过滤菜单
	// 暂时返回所有菜单
	return s.menuService.GetTree()
}

// GetCaptcha 获取验证码
func (s *AuthService) GetCaptcha() (*authDomain.CaptchaResponse, error) {
	// TODO: 实现验证码生成
	// 可以使用 github.com/mojocn/base64Captcha
	return &authDomain.CaptchaResponse{
		CaptchaID:    fmt.Sprintf("captcha_%d", time.Now().UnixNano()),
		CaptchaImage: "data:image/png;base64,iVBORw0KG...",
		ExpiresIn:    300,
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, req *authDomain.ChangePasswordRequest) error {
	// 1. 获取用户
	var user userDomain.User
	err := s.db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新密码
	err = s.db.Model(&user).Update("password", string(hashedPassword)).Error
	if err != nil {
		return err
	}

	log.Info("password changed", zap.Uint("user_id", userID))
	return nil
}

// ForgotPassword 忘记密码 - 发送验证码
func (s *AuthService) ForgotPassword(req *authDomain.ForgotPasswordRequest) error {
	// 1. 检查邮箱是否存在
	var user userDomain.User
	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("邮箱不存在")
		}
		return err
	}

	// 2. 生成 6 位验证码
	code := s.generateVerificationCode()

	// 3. 保存验证码到数据库
	resetCode := &authDomain.PasswordResetCode{
		Email:     req.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}
	if err := s.repo.CreatePasswordResetCode(resetCode); err != nil {
		return err
	}

	// 4. 发送邮件（TODO: 实现邮件发送）
	log.Info("password reset code generated",
		zap.String("email", req.Email),
		zap.String("code", code))

	// TODO: 实际发送邮件
	// s.emailService.Send(req.Email, "密码重置验证码", fmt.Sprintf("您的验证码是: %s，15分钟内有效", code))

	return nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(req *authDomain.ResetPasswordRequest) error {
	// 1. 验证验证码
	resetCode, err := s.repo.GetValidPasswordResetCode(req.Email, req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("验证码无效或已过期")
		}
		return err
	}

	// 2. 获取用户
	var user userDomain.User
	err = s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新密码
		if err := tx.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
			return err
		}

		// 标记验证码为已使用
		if err := s.repo.MarkPasswordResetCodeAsUsed(resetCode.ID); err != nil {
			return err
		}

		log.Info("password reset", zap.String("email", req.Email))
		return nil
	})
}

// recordLoginLog 记录登录日志
func (s *AuthService) recordLoginLog(userID uint, username, ip, device, status, message string) {
	loginLog := &authDomain.LoginLog{
		UserID:   userID,
		Username: username,
		IP:       ip,
		Device:   device,
		Status:   status,
		Message:  message,
	}
	if err := s.repo.CreateLoginLog(loginLog); err != nil {
		log.Error("failed to create login log", zap.Error(err))
	}
}

// getUserPermissions 获取用户权限列表
func (s *AuthService) getUserPermissions(userID uint) []string {
	subject := fmt.Sprintf("user:%d", userID)
	permissions, _ := s.enforcer.GetPermissionsForUser(subject)

	perms := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		if len(perm) >= 3 {
			perms = append(perms, fmt.Sprintf("%s:%s:%s", perm[1], perm[2], perm[3]))
		}
	}

	return perms
}

// generateVerificationCode 生成 6 位验证码
func (s *AuthService) generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// parseUserID 解析用户 ID
func (s *AuthService) parseUserID(userIDStr string) uint {
	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)
	return userID
}
