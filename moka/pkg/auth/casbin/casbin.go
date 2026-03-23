package casbin

import (
	"fmt"
	"moka/pkg/client/mysql"
	"moka/pkg/client/redis"
	"moka/pkg/util/log"
	"slices"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"go.uber.org/zap"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	rediswatcher "github.com/casbin/redis-watcher/v2"
)

var (
	Enforcer *casbin.SyncedCachedEnforcer
	watcher  persist.Watcher
)

func Init() error {
	adapter, err := gormadapter.NewAdapterByDB(mysql.Client)
	if err != nil {
		return fmt.Errorf("failed to create casbin adapter: %w", err)
	}

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

	enforcer, err := casbin.NewSyncedCachedEnforcer(m, adapter)
	if err != nil {
		return fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}

	enforcer.SetExpireTime(10 * time.Minute)

	if err := setupRedisWatcher(enforcer); err != nil {
		return fmt.Errorf("failed to setup redis watcher: %w", err)
	}

	Enforcer = enforcer

	return nil
}

func Close() error {
	if Enforcer == nil {
		return nil
	}

	Enforcer.StopAutoLoadPolicy()

	if watcher != nil {
		watcher.Close()
	}

	return nil
}

func CheckAPI(path, method, userID string) (bool, error) {
	sub := "user:" + userID
	return Enforcer.Enforce(sub, path, method, "api")
}

func CheckMenu(menu, userID string) (bool, error) {
	sub := "user:" + userID
	return Enforcer.Enforce(sub, menu, "view", "menu")
}

func CheckButton(button, userID string) (bool, error) {
	sub := "user:" + userID
	return Enforcer.Enforce(sub, button, "click", "button")
}

func IsSuperAdmin(userID string) (bool, error) {
	sub := "user:" + userID

	roles, err := Enforcer.GetRolesForUser(sub)
	if err != nil {
		return false, err
	}

	if slices.Contains(roles, "role:super_admin") {
		return true, nil
	}

	return false, nil
}

func AddRoleForUser(role, userID string) (bool, error) {
	sub := "user:" + userID

	role = "role:" + role

	ok, err := Enforcer.AddRoleForUser(sub, role)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func DeleteRoleForUser(role, userID string) (bool, error) {
	sub := "user:" + userID

	role = "role:" + role

	ok, err := Enforcer.DeleteRoleForUser(sub, role)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func AddPolicy(sub, obj, act, permType string) (bool, error) {
	ok, err := Enforcer.AddPolicy(sub, obj, act, permType)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func RemovePolicy(sub, obj, act, permType string) (bool, error) {
	ok, err := Enforcer.RemovePolicy(sub, obj, act, permType)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func ReloadPolicy() error {
	if err := Enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to reload policy: %w", err)
	}

	return nil
}

func GetMenus(userID string) ([]string, error) {
	sub := "user:" + userID

	roles, err := Enforcer.GetRolesForUser(sub)
	if err != nil {
		return []string{}, err
	}

	all := append([]string{sub}, roles...)

	menus := make(map[string]bool)
	for _, item := range all {
		policies, err := Enforcer.GetFilteredPolicy(0, item)
		if err != nil {
			return []string{}, err
		}
		for _, policy := range policies {
			if len(policy) >= 4 && policy[3] == "menu" {
				menus[policy[1]] = true
			}
		}
	}

	res := make([]string, 0, len(menus))
	for menu := range menus {
		res = append(res, menu)
	}

	return res, nil
}

func GetButtons(userID string) ([]string, error) {
	sub := "user:" + userID

	roles, err := Enforcer.GetRolesForUser(sub)
	if err != nil {
		return []string{}, err
	}

	all := append([]string{sub}, roles...)

	buttons := make(map[string]bool)
	for _, item := range all {
		policies, err := Enforcer.GetFilteredPolicy(0, item)
		if err != nil {
			return []string{}, err
		}
		for _, policy := range policies {
			if len(policy) >= 4 && policy[3] == "button" {
				buttons[policy[1]] = true
			}
		}
	}

	res := make([]string, 0, len(buttons))
	for button := range buttons {
		res = append(res, button)
	}

	return res, nil
}

func setupRedisWatcher(enforcer *casbin.SyncedCachedEnforcer) error {
	r := redis.Client
	if r == nil {
		log.Warn("Redis client not available, running in standalone mode")
		return nil
	}

	w, err := rediswatcher.NewWatcher(r.Options().Addr, rediswatcher.WatcherOptions{
		Channel: "/casbin/policy",
	})
	if err != nil {
		return fmt.Errorf("failed to create redis watcher: %w", err)
	}

	if err := enforcer.SetWatcher(w); err != nil {
		return fmt.Errorf("failed to set watcher: %w", err)
	}

	if err := w.SetUpdateCallback(func(msg string) {
		log.Info("Policy updated from other instance, reloading...", zap.String("msg", msg))
		if err := enforcer.LoadPolicy(); err != nil {
			log.Error("Failed to reload policy", zap.Error(err))
		}
	}); err != nil {
		return fmt.Errorf("failed to set update callback: %w", err)
	}

	watcher = w
	log.Info("Redis watcher initialized successfully, multi-instance support enabled")
	return nil
}
