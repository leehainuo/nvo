package service

import (
	"fmt"

	"nvo-api/core"
	"nvo-api/internal/system/permission/domain"

	"github.com/casbin/casbin/v3"
)

// PermissionService 权限业务逻辑层
type PermissionService struct {
	pocket   *core.Pocket
	enforcer *casbin.SyncedEnforcer
}

// NewPermissionService 创建权限服务
func NewPermissionService(pocket *core.Pocket) domain.PermissionService {
	return &PermissionService{
		pocket:   pocket,
		enforcer: pocket.Enforcer,
	}
}

// CheckPermission 检查用户是否有权限
func (s *PermissionService) CheckPermission(userID uint, object, action string) (bool, error) {
	subject := fmt.Sprintf("user:%d", userID)
	ok, err := s.enforcer.Enforce(subject, object, action)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// GetUserMenus 获取用户菜单权限
func (s *PermissionService) GetUserMenus(userID uint) ([]domain.Menu, error) {
	subject := fmt.Sprintf("user:%d", userID)

	// 获取用户的角色
	roles, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, err
	}

	// 获取菜单权限
	menuMap := make(map[string]domain.Menu)

	// 获取用户直接菜单权限
	userPerms, _ := s.enforcer.GetPermissionsForUser(subject)
	for _, perm := range userPerms {
		if len(perm) >= 4 && perm[3] == "menu" {
			menuMap[perm[1]] = domain.Menu{
				ID:   perm[1],
				Name: perm[1],
				Path: perm[1],
			}
		}
	}

	// 获取角色菜单权限
	for _, role := range roles {
		rolePerms, _ := s.enforcer.GetPermissionsForUser(role)
		for _, perm := range rolePerms {
			if len(perm) >= 4 && perm[3] == "menu" {
				menuMap[perm[1]] = domain.Menu{
					ID:   perm[1],
					Name: perm[1],
					Path: perm[1],
				}
			}
		}
	}

	// 转换为切片
	menus := make([]domain.Menu, 0, len(menuMap))
	for _, menu := range menuMap {
		menus = append(menus, menu)
	}

	return menus, nil
}

// GetUserButtons 获取用户按钮权限
func (s *PermissionService) GetUserButtons(userID uint) ([]string, error) {
	subject := fmt.Sprintf("user:%d", userID)

	// 获取用户的角色
	roles, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, err
	}

	// 获取按钮权限
	buttonMap := make(map[string]bool)

	// 获取用户直接按钮权限
	userPerms, _ := s.enforcer.GetPermissionsForUser(subject)
	for _, perm := range userPerms {
		if len(perm) >= 4 && perm[3] == "button" {
			buttonMap[perm[1]+":"+perm[2]] = true
		}
	}

	// 获取角色按钮权限
	for _, role := range roles {
		rolePerms, _ := s.enforcer.GetPermissionsForUser(role)
		for _, perm := range rolePerms {
			if len(perm) >= 4 && perm[3] == "button" {
				buttonMap[perm[1]+":"+perm[2]] = true
			}
		}
	}

	// 转换为切片
	buttons := make([]string, 0, len(buttonMap))
	for button := range buttonMap {
		buttons = append(buttons, button)
	}

	return buttons, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(userID uint) (*domain.UserPermissions, error) {
	menus, err := s.GetUserMenus(userID)
	if err != nil {
		return nil, err
	}

	buttons, err := s.GetUserButtons(userID)
	if err != nil {
		return nil, err
	}

	// 获取API权限
	subject := fmt.Sprintf("user:%d", userID)
	roles, _ := s.enforcer.GetRolesForUser(subject)

	apiMap := make(map[string]bool)

	// 获取用户直接API权限
	userPerms, _ := s.enforcer.GetPermissionsForUser(subject)
	for _, perm := range userPerms {
		if len(perm) >= 4 && perm[3] == "api" {
			apiMap[perm[1]+":"+perm[2]] = true
		}
	}

	// 获取角色API权限
	for _, role := range roles {
		rolePerms, _ := s.enforcer.GetPermissionsForUser(role)
		for _, perm := range rolePerms {
			if len(perm) >= 4 && perm[3] == "api" {
				apiMap[perm[1]+":"+perm[2]] = true
			}
		}
	}

	apis := make([]string, 0, len(apiMap))
	for api := range apiMap {
		apis = append(apis, api)
	}

	return &domain.UserPermissions{
		Menus:   menus,
		Buttons: buttons,
		APIs:    apis,
	}, nil
}
