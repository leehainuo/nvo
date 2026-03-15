package auth

import (
	"fmt"

	"nvo-api/core/log"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionAPI    PermissionType = "api"    // API 接口权限
	PermissionButton PermissionType = "button" // 按钮权限
	PermissionMenu   PermissionType = "menu"   // 菜单权限
)

// NewEnforcer 创建权限执行器
func NewEnforcer(db *gorm.DB) (*casbin.SyncedEnforcer, error) {
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

	return enforcer, nil
}
