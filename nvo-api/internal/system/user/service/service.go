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
// 显式声明所需依赖，清晰明了
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

	return &userDomain.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Status:    user.Status,
		Roles:     roles,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
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

	// 转换为响应格式
	responses := make([]*userDomain.UserResponse, 0, len(users))
	for _, user := range users {
		subject := fmt.Sprintf("user:%d", user.ID)
		roles, _ := s.enforcer.GetRolesForUser(subject)

		responses = append(responses, &userDomain.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Phone:     user.Phone,
			Avatar:    user.Avatar,
			Status:    user.Status,
			Roles:     roles,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return responses, total, nil
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

// GetUserWithRoles 获取用户及其角色详情（跨模块调用）
func (s *UserService) GetUserWithRoles(id uint) (*userDomain.UserWithRoles, error) {
	// 1. 获取用户基本信息
	userResp, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 2. ✅ 直接调用 RoleService（精确依赖）
	if s.roleService == nil {
		return &userDomain.UserWithRoles{
			UserResponse: userResp,
			RoleDetails:  []*userDomain.RoleDetail{},
		}, nil
	}

	roles, err := s.roleService.GetRolesByUserID(id)
	if err != nil {
		log.Warn("failed to get user roles", zap.Error(err))
		return &userDomain.UserWithRoles{
			UserResponse: userResp,
			RoleDetails:  []*userDomain.RoleDetail{},
		}, nil
	}

	// 3. 转换为角色详情
	roleDetails := make([]*userDomain.RoleDetail, 0, len(roles))
	for _, role := range roles {
		roleDetails = append(roleDetails, &userDomain.RoleDetail{
			ID:          role.ID,
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return &userDomain.UserWithRoles{
		UserResponse: userResp,
		RoleDetails:  roleDetails,
	}, nil
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
