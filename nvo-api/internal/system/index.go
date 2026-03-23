package system

import (
	"nvo-api/core"
	"nvo-api/core/log"
	"nvo-api/internal"
	"nvo-api/internal/system/audit"
	"nvo-api/internal/system/auth"
	"nvo-api/internal/system/dept"
	"nvo-api/internal/system/dict"
	"nvo-api/internal/system/menu"
	"nvo-api/internal/system/permission"
	"nvo-api/internal/system/role"
	"nvo-api/internal/system/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterModules 注册所有系统模块
// 三阶段初始化：user.NewModule 需要访问 p.System.Role，所以需要先创建 SystemService
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
	log.Info("Registering system modules...")

	// 阶段 1：初始化无依赖模块
	permModule  := permission.NewModule(p)
	roleModule  := role.NewModule(p)
	menuModule  := menu.NewModule(p)
	deptModule  := dept.NewModule(p)
	auditModule := audit.NewModule(p)
	dictModule  := dict.NewModule(p)

	// 阶段 2：初始化 SystemService
	// UserService 依赖 RoleService 在阶段3注入
	p.System = internal.NewSystemService(
		nil,
		roleModule.Service(),
		permModule.Service(),
		menuModule.Service(),
		deptModule.Service(),
		auditModule.Service(),
		nil,
		dictModule.Service(),
	)

	// 阶段 3：初始化单向依赖模块
	userModule := user.NewModule(p)
	p.System.User = userModule.Service()

	// 阶段 4：初始化认证模块（依赖 User 和 Menu）
	authModule := auth.NewModule(p)
	p.System.Auth = authModule.Service()

	// 阶段 5：数据库迁移和路由注册
	modules := []internal.Module{
		permModule,
		roleModule,
		userModule,
		menuModule,
		deptModule,
		auditModule,
		authModule,
		dictModule,
	}

	if err := migrateModels(p.DB, modules); err != nil {
		log.Fatal("Database migration failed", zap.Error(err))
	}

	for _, module := range modules {
		module.RegisterRoutes(r)
	}

	log.Info("✓ System modules registered successfully (8 modules)")
}

// migrateModels 收集并迁移所有模块的数据模型
func migrateModels(db *gorm.DB, modules []internal.Module) error {
	var allModels []any

	// 收集所有模型
	for _, module := range modules {
		models := module.Models()
		if len(models) > 0 {
			log.Info("Collecting models from module",
				zap.String("module", module.Name()),
				zap.Int("count", len(models)))
			allModels = append(allModels, models...)
		}
	}

	// 统一迁移
	if len(allModels) > 0 {
		log.Info("Starting database migration", zap.Int("total_models", len(allModels)))
		if err := db.AutoMigrate(allModels...); err != nil {
			return err
		}
		log.Info("Database migration completed successfully")
	} else {
		log.Info("No models to migrate")
	}

	return nil
}
