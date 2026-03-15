package core

import (
	"fmt"
	"time"

	"nvo-api/core/auth"
	"nvo-api/core/config"
	"nvo-api/core/log"
	"nvo-api/core/middleware"
	"nvo-api/internal"
	"nvo-api/pkg/client/mysql"
	"nvo-api/pkg/client/redis"
	"nvo-api/pkg/util/jwt"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"

	goredis "github.com/redis/go-redis/v9"
)

// Pocket - 口袋 （依赖注入容器）
// 管理所有应用级依赖和业务服务，业务模块通过 Pocket 获取所需资源
type Pocket struct {
	// 基础设施依赖
	Config      *config.Config
	DB          *gorm.DB
	Redis       *goredis.Client
	JWT         *jwt.JWT
	Enforcer    *casbin.SyncedEnforcer
	GinEngine   *gin.Engine
	RateLimiter middleware.RateLimiter

	// 业务服务
	System *internal.SystemService

	// 未来可扩展其他模块：
	// Business *businessDomain.BusinessService // 业务模块服务聚合
	// Platform *platformDomain.PlatformService // 平台模块服务聚合
}

// Close 优雅关闭所有资源
func (p *Pocket) Close() error {
	var errs []error

	// 关闭 Redis
	if p.Redis != nil {
		if err := p.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis: %w", err))
		}
	}

	// 关闭数据库
	if p.DB != nil {
		if sqlDB, err := p.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close database: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// PocketBuilder 构建器模式，用于初始化 Pocket
type PocketBuilder struct {
	configPath     string
	config         *config.Config
	viper          *viper.Viper
	pocket         *Pocket
	enableEnforcer bool
	skipDB         bool
	skipRedis      bool
}

// NewPocketBuilder 创建 Pocket 构建器
func NewPocketBuilder(path string) *PocketBuilder {
	return &PocketBuilder{
		configPath: path,
		pocket:     &Pocket{},
	}
}

// Build 构建并初始化所有依赖
func (p *PocketBuilder) Build() (*Pocket, error) {
	// 1. 加载配置
	if err := p.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. 初始化数据库
	if !p.skipDB {
		if err := p.initDB(); err != nil {
			return nil, fmt.Errorf("failed to init database: %w", err)
		}
	}

	// 3. 初始化 Redis（允许失败，降级到内存）
	if !p.skipRedis {
		if err := p.initRedis(); err != nil {
			log.Warn("Redis init failed, will use memory fallback", zap.Error(err))
		}
	}

	// 4. 初始化 JWT
	if err := p.initJWT(); err != nil {
		return nil, fmt.Errorf("failed to init JWT: %w", err)
	}

	// 5. 初始化权限控制（可选）
	if p.enableEnforcer {
		if err := p.initEnforcer(); err != nil {
			return nil, fmt.Errorf("failed to init enforcer: %w", err)
		}
	}

	// 6. 初始化限流器
	p.initRateLimiter()

	// 7. 初始化 Gin 引擎
	p.initGin()

	log.Info("Pocket initialized successfully")
	return p.pocket, nil
}

// loadConfig 加载配置
func (p *PocketBuilder) loadConfig() error {
	c, v, err := config.LoadConfig(p.configPath)
	if err != nil {
		return err
	}
	p.config = c
	p.viper = v
	p.pocket.Config = c
	return nil
}

// initDB 初始化数据库
func (p *PocketBuilder) initDB() error {
	database, err := mysql.InitFromViper(p.viper, "database")
	if err != nil {
		return err
	}
	p.pocket.DB = database
	log.Info("Database connected successfully")
	return nil
}

// initRedis 初始化 Redis
func (p *PocketBuilder) initRedis() error {
	rdb, err := redis.InitFromViper(p.viper, "redis")
	if err != nil {
		return err
	}
	p.pocket.Redis = rdb
	log.Info("Redis connected successfully")
	return nil
}

// initJWT 初始化 JWT
func (p *PocketBuilder) initJWT() error {
	jwt, err := jwt.InitFromViper(p.viper, "jwt")
	if err != nil {
		return err
	}
	p.pocket.JWT = jwt
	log.Info("JWT initialized successfully")
	return nil
}

// initRateLimiter 初始化限流器
func (p *PocketBuilder) initRateLimiter() {
	rate := p.viper.GetInt("ratelimit.rate")
	capacity := p.viper.GetInt("ratelimit.capacity")
	window := p.viper.GetDuration("ratelimit.window")

	if rate == 0 {
		rate = 10
	}
	if capacity == 0 {
		capacity = 100
	}
	if window == 0 {
		window = time.Minute
	}

	p.pocket.RateLimiter = middleware.NewAutoFallbackRateLimiter(
		p.pocket.Redis,
		rate,
		capacity,
		window,
	)
	log.Info("Rate limiter initialized",
		zap.Int("rate", rate),
		zap.Int("capacity", capacity),
		zap.Duration("window", window))
}

// initGin 初始化 Gin 引擎
func (p *PocketBuilder) initGin() {
	gin.SetMode(p.config.Server.Mode)
	engine := gin.New()
	engine.Use(
		middleware.Logger(),
		middleware.Recovery(true),
	)

	p.pocket.GinEngine = engine
	log.Info("Gin engine initialized")
}

// initEnforcer 初始化权限控制
func (p *PocketBuilder) initEnforcer() error {
	enforcer, err := auth.NewEnforcer(p.pocket.DB)
	if err != nil {
		return err
	}
	p.pocket.Enforcer = enforcer
	log.Info("Enforcer initialized successfully")

	// 初始化默认权限（可选）
	if err := auth.InitDefaultPermissions(enforcer); err != nil {
		log.Warn("failed to init default permissions", zap.Error(err))
	}

	return nil
}

// WithEnforcer 启用 Casbin 权限控制
func (p *PocketBuilder) WithEnforcer() *PocketBuilder {
	p.enableEnforcer = true
	return p
}

// SkipDB 跳过数据库初始化（用于不需要数据库的服务）
func (p *PocketBuilder) SkipDB() *PocketBuilder {
	p.skipDB = true
	return p
}

// SkipRedis 跳过 Redis 初始化（用于不需要缓存的服务）
func (p *PocketBuilder) SkipRedis() *PocketBuilder {
	p.skipRedis = true
	return p
}

// MustBuild 构建 Pocket，失败则 panic
func (p *PocketBuilder) MustBuild() *Pocket {
	pocket, err := p.Build()
	if err != nil {
		panic(err)
	}
	return pocket
}
