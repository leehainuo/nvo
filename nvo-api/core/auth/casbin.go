package auth

import (
	"fmt"
	
	"nvo-api/core/log"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionAPI    PermissionType = "api"    // API 接口权限
	PermissionButton PermissionType = "button" // 按钮权限
	PermissionMenu   PermissionType = "menu"   // 菜单权限
)

// Enforcer Casbin 权限执行器
type Enforcer struct {
	enforcer *casbin.SyncedEnforcer
}

// NewEnforcer 创建权限执行器
func NewEnforcer(db *gorm.DB) (*Enforcer, error) {
	// 使用 GORM Adapter 持久化权限策略
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 定义 RBAC 模型
	m := model.NewModel()
	m.LoadModelFromText(`
	[request_definition]
	r = sub, obj, act, type

	[policy_definition]
	p = sub, obj, act, type

	[role_definition]
	g = _, _

	[policy_effect]
	e = some(where (p.eft == allow))

	[matchers]
	m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.type == p.type
	`)

	// 创建 Enforcer
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	log.Info("Casbin enforcer initialized successfully")

	return &Enforcer{
		enforcer: enforcer,
	}, nil
}

// CheckAPI 检查 API 权限
// 示例: CheckAPI("user:123", "/api/v1/users", "GET")
func (e *Enforcer) CheckAPI(subject, path, method string) (bool, error) {
	ok, err := e.enforcer.Enforce(subject, path, method, string(PermissionAPI))
	if err != nil {
		log.Error("failed to check API permission",
			zap.String("subject", subject),
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return false, err
	}
	return ok, nil
}

// CheckButton 检查按钮权限
// 示例: CheckButton("user:123", "user.delete", "click")
func (e *Enforcer) CheckButton(subject, buttonCode, action string) (bool, error) {
	ok, err := e.enforcer.Enforce(subject, buttonCode, action, string(PermissionButton))
	if err != nil {
		log.Error("failed to check button permission",
			zap.String("subject", subject),
			zap.String("button", buttonCode),
			zap.String("action", action),
			zap.Error(err))
		return false, err
	}
	return ok, nil
}

// CheckMenu 检查菜单权限
// 示例: CheckMenu("user:123", "system.user", "view")
func (e *Enforcer) CheckMenu(subject, menuCode, action string) (bool, error) {
	ok, err := e.enforcer.Enforce(subject, menuCode, action, string(PermissionMenu))
	if err != nil {
		log.Error("failed to check menu permission",
			zap.String("subject", subject),
			zap.String("menu", menuCode),
			zap.String("action", action),
			zap.Error(err))
		return false, err
	}
	return ok, nil
}

// AddRoleForUser 为用户添加角色
func (e *Enforcer) AddRoleForUser(user, role string) (bool, error) {
	ok, err := e.enforcer.AddRoleForUser(user, role)
	if err != nil {
		log.Error("failed to add role for user",
			zap.String("user", user),
			zap.String("role", role),
			zap.Error(err))
		return false, err
	}
	log.Info("role added for user",
		zap.String("user", user),
		zap.String("role", role))
	return ok, nil
}

// AddPolicy 添加权限策略
func (e *Enforcer) AddPolicy(subject, object, action string, permType PermissionType) (bool, error) {
	ok, err := e.enforcer.AddPolicy(subject, object, action, string(permType))
	if err != nil {
		log.Error("failed to add policy",
			zap.String("subject", subject),
			zap.String("object", object),
			zap.String("action", action),
			zap.String("type", string(permType)),
			zap.Error(err))
		return false, err
	}
	log.Info("policy added",
		zap.String("subject", subject),
		zap.String("object", object),
		zap.String("action", action),
		zap.String("type", string(permType)))
	return ok, nil
}

// RemovePolicy 删除权限策略
func (e *Enforcer) RemovePolicy(subject, object, action string, permType PermissionType) (bool, error) {
	ok, err := e.enforcer.RemovePolicy(subject, object, action, string(permType))
	if err != nil {
		log.Error("failed to remove policy",
			zap.String("subject", subject),
			zap.String("object", object),
			zap.String("action", action),
			zap.String("type", string(permType)),
			zap.Error(err))
		return false, err
	}
	return ok, nil
}

// GetRolesForUser 获取用户的所有角色
func (e *Enforcer) GetRolesForUser(user string) ([]string, error) {
	roles, err := e.enforcer.GetRolesForUser(user)
	if err != nil {
		log.Error("failed to get roles for user",
			zap.String("user", user),
			zap.Error(err))
		return nil, err
	}
	return roles, nil
}

// GetPermissionsForUser 获取用户的所有权限
func (e *Enforcer) GetPermissionsForUser(user string) ([][]string, error) {
	perms, err := e.enforcer.GetPermissionsForUser(user)
	if err != nil {
		log.Error("failed to get permissions for user",
			zap.String("user", user),
			zap.Error(err))
		return nil, err
	}
	return perms, nil
}
