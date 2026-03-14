package auth

import (
	"fmt"
	
	"nvo-api/core/log"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// NewEnforcerWithWildcard 创建支持通配符的权限执行器（高级版本）
// 支持路径通配符匹配，更灵活的权限配置
func NewEnforcerWithWildcard(db *gorm.DB) (*Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 使用 KeyMatch2 支持通配符
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
	m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == "*") && r.type == p.type
	`)

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	log.Info("Casbin enforcer (with wildcard) initialized successfully")

	return &Enforcer{
		enforcer: enforcer,
	}, nil
}

// 使用示例：
//
// 1. 通配符路径匹配
// p, role:admin, /api/v1/*, *, api              // 所有 v1 接口
// p, role:admin, /api/v1/users/*, GET, api      // 所有用户查询
//
// 2. 通配符操作匹配
// p, role:admin, /api/v1/users, *, api          // 用户接口的所有操作
//
// 3. 精确匹配（仍然支持）
// p, role:user, /api/v1/profile, GET, api       // 只能 GET
//
// KeyMatch2 规则：
// /api/v1/*        匹配 /api/v1/users, /api/v1/orders
// /api/v1/users/*  匹配 /api/v1/users/123, /api/v1/users/456
// /api/*/users     匹配 /api/v1/users, /api/v2/users
