package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"nvo-api/core"
	"nvo-api/core/log"
	"nvo-api/internal/system/role/domain"
	"nvo-api/internal/system/role/repository"

	"github.com/casbin/casbin/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleService 角色业务逻辑层
type RoleService struct {
	pocket   *core.Pocket
	repo     *repository.RoleRepository
	enforcer *casbin.SyncedEnforcer
	db       *gorm.DB
}

// NewRoleService 创建角色服务
func NewRoleService(pocket *core.Pocket) domain.RoleService {
	return &RoleService{
		pocket:   pocket,
		enforcer: pocket.Enforcer,
		db:       pocket.DB,
		repo:     repository.NewRoleRepository(pocket.DB),
	}
}

// Create 创建角色
func (s *RoleService) Create(req *domain.CreateRoleRequest) (*domain.Role, error) {
	// 检查角色编码是否存在
	exists, err := s.repo.ExistsByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("角色编码已存在")
	}

	// 创建角色
	role := &domain.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      1,
	}

	if err := s.repo.Create(role); err != nil {
		return nil, err
	}

	log.Info("role created", zap.String("code", role.Code), zap.Uint("id", role.ID))
	return role, nil
}

// GetByID 根据 ID 获取角色
func (s *RoleService) GetByID(id uint) (*domain.Role, error) {
	role, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}
	return role, nil
}

// Update 更新角色
func (s *RoleService) Update(id uint, req *domain.UpdateRoleRequest) error {
	role, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 更新字段
	role.Name = req.Name
	role.Description = req.Description
	role.Sort = req.Sort
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := s.repo.Update(role); err != nil {
		return err
	}

	log.Info("role updated", zap.Uint("id", id))
	return nil
}

// Delete 删除角色
func (s *RoleService) Delete(id uint) error {
	// 检查角色是否存在
	role, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色
		if err := s.repo.Delete(id); err != nil {
			return err
		}

		// 删除角色的所有权限
		roleSubject := fmt.Sprintf("role:%d", id)
		if _, err := s.enforcer.DeletePermissionsForUser(roleSubject); err != nil {
			return fmt.Errorf("删除角色权限失败: %w", err)
		}

		// 删除所有用户的该角色
		if _, err := s.enforcer.DeleteRole(roleSubject); err != nil {
			return fmt.Errorf("删除角色关联失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("role deleted", zap.Uint("id", id), zap.String("code", role.Code))
	return nil
}

// List 获取角色列表
func (s *RoleService) List(req *domain.ListRoleRequest) ([]*domain.Role, int64, error) {
	return s.repo.List(req)
}

// AssignPermissions 为角色分配权限
func (s *RoleService) AssignPermissions(id uint, req *domain.AssignPermissionsRequest) error {
	// 检查角色是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	roleSubject := fmt.Sprintf("role:%d", id)

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色的所有权限
		if _, err := s.enforcer.DeletePermissionsForUser(roleSubject); err != nil {
			return fmt.Errorf("删除旧权限失败: %w", err)
		}

		// 添加新权限
		for _, perm := range req.Permissions {
			if _, err := s.enforcer.AddPolicy(roleSubject, perm.Object, perm.Action, perm.Type); err != nil {
				return fmt.Errorf("添加权限失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("permissions assigned", zap.Uint("role_id", id), zap.Int("count", len(req.Permissions)))
	return nil
}

// GetPermissions 获取角色的权限列表
func (s *RoleService) GetPermissions(id uint) ([][]string, error) {
	// 检查角色是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	roleSubject := fmt.Sprintf("role:%d", id)
	permissions, err := s.enforcer.GetPermissionsForUser(roleSubject)
	if err != nil {
		return nil, fmt.Errorf("获取权限失败: %w", err)
	}

	return permissions, nil
}

// GetAll 获取所有启用的角色
func (s *RoleService) GetAll() ([]*domain.Role, error) {
	return s.repo.GetAll()
}

// GetRolesByUserID 根据用户ID获取角色列表（跨模块调用）
func (s *RoleService) GetRolesByUserID(userID uint) ([]*domain.Role, error) {
	subject := fmt.Sprintf("user:%d", userID)
	roleCodes, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}

	if len(roleCodes) == 0 {
		return []*domain.Role{}, nil
	}

	// 解析角色ID
	roleIDs := make([]uint, 0, len(roleCodes))
	for _, code := range roleCodes {
		// code 格式: "role:1"
		parts := strings.Split(code, ":")
		if len(parts) == 2 {
			if id, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				roleIDs = append(roleIDs, uint(id))
			}
		}
	}

	if len(roleIDs) == 0 {
		return []*domain.Role{}, nil
	}

	// 批量查询角色
	var roles []*domain.Role
	if err := s.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}
