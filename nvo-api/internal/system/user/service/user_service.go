package service

import (
	"errors"
	"fmt"

	"nvo-api/internal/system/user/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	repo domain.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register 用户注册
func (s *UserService) Register(username, email, password string) (*domain.User, error) {
	existUser, _ := s.repo.FindByUsername(username)
	if existUser != nil {
		return nil, errors.New("username already exists")
	}

	existEmail, _ := s.repo.FindByEmail(email)
	if existEmail != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Nickname: username,
		Status:   1,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*domain.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*domain.User, error) {
	return s.repo.FindByID(id)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int) ([]*domain.User, int64, error) {
	return s.repo.List(page, pageSize)
}
