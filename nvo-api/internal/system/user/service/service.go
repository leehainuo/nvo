package service

import (
	"errors"
	"fmt"

	"nvo-api/core/log"
	"nvo-api/internal/system/user/repository"

	"github.com/casbin/casbin/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	roleDomain "nvo-api/internal/system/role/domain"
	userDomain "nvo-api/internal/system/user/domain"
)

// UserService 用户业务逻辑层
type UserService struct {
	db          *gorm.DB
	enforcer    *casbin.SyncedEnforcer
	repo        *repository.UserRepository
	roleService roleDomain.RoleService
}

// NewUserService 创建用户服务
func NewUserService(
	db *gorm.DB,
	enforcer *casbin.SyncedEnforcer,
	roleService roleDomain.RoleService,
) userDomain.UserService {
	return &UserService{
		db:          db,
		enforcer:    enforcer,
		repo:        repository.NewUserRepository(db),
		roleService: roleService,
	}
}

// Create 创建用户
func (s *UserService) Create(req *userDomain.CreateUserRequest) (*userDomain.User, error) {
	// 检查用户名是否存在
	exists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if req.Email != "" {
		exists, err := s.repo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户对象
	user := &userDomain.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		if err := s.repo.Create(user); err != nil {
			return err
		}

		// 分配角色
		if len(req.RoleIDs) > 0 {
			subject := fmt.Sprintf("user:%d", user.ID)
			for _, roleID := range req.RoleIDs {
				roleCode := fmt.Sprintf("role:%d", roleID)
				if _, err := s.enforcer.AddRoleForUser(subject, roleCode); err != nil {
					return fmt.Errorf("分配角色失败: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("user created", zap.String("username", user.Username), zap.Uint("id", user.ID))
	return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id uint) (*userDomain.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 获取用户角色
	subject := fmt.Sprintf("user:%d", user.ID)
	roles, _ := s.enforcer.GetRolesForUser(subject)

	return user.ToResponse(roles), nil
}

// Update 更新用户
func (s *UserService) Update(id uint, req *userDomain.UpdateUserRequest) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.ExistsByEmail(req.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("邮箱已被使用")
		}
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新用户信息
		if req.Nickname != "" {
			user.Nickname = req.Nickname
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.Phone != "" {
			user.Phone = req.Phone
		}
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}

		if err := s.repo.Update(user); err != nil {
			return err
		}

		// 更新角色
		if req.RoleIDs != nil {
			subject := fmt.Sprintf("user:%d", user.ID)

			// 删除旧角色
			if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
				return fmt.Errorf("删除旧角色失败: %w", err)
			}

			// 添加新角色
			for _, roleID := range req.RoleIDs {
				roleCode := fmt.Sprintf("role:%d", roleID)
				if _, err := s.enforcer.AddRoleForUser(subject, roleCode); err != nil {
					return fmt.Errorf("分配角色失败: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("user updated", zap.Uint("id", id))
	return nil
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
	// 检查用户是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除用户
		if err := s.repo.Delete(id); err != nil {
			return err
		}

		// 删除用户的所有角色和权限
		subject := fmt.Sprintf("user:%d", id)
		if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
			return fmt.Errorf("删除用户角色失败: %w", err)
		}
		if _, err := s.enforcer.DeletePermissionsForUser(subject); err != nil {
			return fmt.Errorf("删除用户权限失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("user deleted", zap.Uint("id", id))
	return nil
}

// List 获取用户列表
func (s *UserService) List(req *userDomain.ListUserRequest) ([]*userDomain.UserResponse, int64, error) {
	users, total, err := s.repo.List(req)
	if err != nil {
		return nil, 0, err
	}

	if len(users) == 0 {
		return []*userDomain.UserResponse{}, 0, nil
	}

	// ✅ 批量查询所有用户的角色（解决 N+1 问题）
	rolesMap := s.batchGetUserRoles(users)

	// 转换为响应格式
	return userDomain.ToResponseList(users, rolesMap), total, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.Password = string(hashedPassword)
	if err := s.repo.Update(user); err != nil {
		return err
	}

	log.Info("password changed", zap.Uint("user_id", id))
	return nil
}

// batchGetUserRoles 批量获取用户角色
func (s *UserService) batchGetUserRoles(users []*userDomain.User) map[uint][]string {
	rolesMap := make(map[uint][]string)

	if s.enforcer == nil {
		return rolesMap
	}

	// 构建所有用户的 subject 列表
	userSubjects := make([]string, len(users))
	subjectToID := make(map[string]uint)
	for i, user := range users {
		subject := fmt.Sprintf("user:%d", user.ID)
		userSubjects[i] = subject
		subjectToID[subject] = user.ID
		rolesMap[user.ID] = []string{} // 初始化空数组
	}

	// 批量查询 casbin_rule 表（一次查询获取所有用户的角色）
	type CasbinRule struct {
		V0 string // user:1
		V1 string // role:1
	}
	var rules []CasbinRule
	err := s.db.Table("casbin_rule").
		Select("v0, v1").
		Where("ptype = ? AND v0 IN ?", "g", userSubjects).
		Find(&rules).Error

	if err != nil {
		log.Warn("batch get user roles failed", zap.Error(err))
		return rolesMap
	}

	// 构建角色映射
	for _, rule := range rules {
		if userID, ok := subjectToID[rule.V0]; ok {
			rolesMap[userID] = append(rolesMap[userID], rule.V1)
		}
	}

	return rolesMap
}

// AssignRoles 为用户分配角色
func (s *UserService) AssignRoles(userID uint, roleIDs []uint) error {
	// 检查 Enforcer 是否可用
	if s.enforcer == nil {
		return errors.New("权限系统不可用")
	}

	// 检查用户是否存在
	_, err := s.repo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	subject := fmt.Sprintf("user:%d", userID)

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧角色
		if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
			return fmt.Errorf("删除旧角色失败: %w", err)
		}

		// 添加新角色
		for _, roleID := range roleIDs {
			roleCode := fmt.Sprintf("role:%d", roleID)
			if _, err := s.enforcer.AddRoleForUser(subject, roleCode); err != nil {
				return fmt.Errorf("分配角色失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("roles assigned", zap.Uint("user_id", userID), zap.Int("count", len(roleIDs)))
	return nil
}
