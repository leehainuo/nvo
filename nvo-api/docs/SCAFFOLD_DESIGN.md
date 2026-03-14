# NVO 脚手架设计文档

## 架构概述

基于 **Pocket（口袋）** 依赖注入容器的优雅脚手架设计，参考 go-zero 的 ServiceContext 模式，实现了清晰的依赖管理和模块化架构。

## 核心设计理念

### 1. Pocket - 依赖注入容器

`Pocket` 是整个应用的核心，负责管理所有应用级依赖：

```go
type Pocket struct {
    Config    *core.Config        // 配置
    DB        *gorm.DB            // 数据库
    Redis     *redis.Client       // 缓存
    Logger    *zap.Logger         // 日志
    Enforcer  *casbin.Enforcer    // 权限（可选）
    GinEngine *gin.Engine         // Web引擎
}
```

### 2. Builder 模式初始化

使用 Builder 模式优雅地初始化 Pocket，保证依赖按顺序加载：

```go
pocket := core.NewPocketBuilder("config.yml").MustBuild()
```

初始化顺序：
1. 加载配置 (Config)
2. 初始化日志 (Logger)
3. 初始化数据库 (DB)
4. 初始化 Redis
5. 初始化 Gin 引擎

### 3. 模块化业务架构

每个业务模块采用 DDD 分层架构：

```
internal/mods/system/user/
├── domain/          # 领域层（实体、接口定义）
│   └── entity.go    # User 实体 + UserRepository 接口
├── repository/      # 数据访问层（实现 Repository 接口）
│   └── user_repo.go # MySQL 实现
├── service/         # 业务逻辑层
│   └── user_service.go
├── api/             # 接口层（HTTP Handler）
│   └── handler.go
└── module.go        # 模块入口（组装依赖 + 注册路由）
```

## 目录结构

```
nvo-api/
├── cmd/                    # 程序入口
│   └── main.go             # 主程序（初始化 Pocket + 注册模块）
├── core/                   # 核心框架
│   ├── config/             # 配置管理
│   │   └── config.go       # Viper 配置加载
│   ├── db/                 # 数据库初始化
│   │   └── gorm.go         # GORM 连接池配置
│   ├── logger/             # 日志初始化
│   │   └── logger.go       # Zap 日志配置
│   ├── pocket/             # 依赖注入容器
│   │   └── pocket.go       # Pocket 核心
│   └── auth/               # 权限管理（Casbin）
├── internal/               # 业务模块
│   └── mods/
│       ├── system/         # 系统模块
│       │   └── user/       # 用户模块
│       └── custom/         # 自定义模块
├── pkg/                    # 通用工具
│   ├── client/             # 客户端封装
│   │   ├── mysql/
│   │   └── redis/
│   └── util/               # 工具函数
├── config.yml              # 配置文件
└── go.mod
```

## 使用示例

### 1. 主程序入口

```go
// cmd/main.go
func main() {
    // 1. 初始化 Pocket（一行代码完成所有依赖注入）
    pocket := core.NewPocketBuilder("config.yml").MustBuild()
    defer cleanup(pocket)

    // 2. 初始化业务模块（传入 Pocket）
    userModule := user.NewModule(pocket)

    // 3. 注册路由
    api := pocket.GinEngine.Group("/api/v1")
    userModule.RegisterRoutes(api)

    // 4. 启动服务
    addr := fmt.Sprintf("%s:%d", pocket.Config.Server.Host, pocket.Config.Server.Port)
    pocket.GinEngine.Run(addr)
}
```

### 2. 业务模块开发

#### 2.1 定义领域实体和接口

```go
// domain/entity.go
type User struct {
    ID       uint   `gorm:"primarykey"`
    Username string `gorm:"uniqueIndex"`
    Email    string
}

type UserRepository interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
}
```

#### 2.2 实现 Repository

```go
// repository/user_repo.go
type userRepository struct {
    db *gorm.DB  // 从 Pocket 获取
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
    return &userRepository{db: db}
}
```

#### 2.3 实现 Service

```go
// service/user_service.go
type UserService struct {
    repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) Register(username, email, password string) (*domain.User, error) {
    // 业务逻辑
}
```

#### 2.4 实现 Handler

```go
// api/handler.go
type UserHandler struct {
    userService *service.UserService
}

func (h *UserHandler) Register(c *gin.Context) {
    // HTTP 处理
}
```

#### 2.5 组装模块

```go
// module.go
type Module struct {
    pocket  *core.Pocket
    handler *api.UserHandler
}

func NewModule(pocket *core.Pocket) *Module {
    // 依赖注入：从 Pocket 获取 DB，逐层组装
    userRepo := repository.NewUserRepository(pocket.DB)
    userService := service.NewUserService(userRepo)
    userHandler := api.NewUserHandler(userService)

    // 自动迁移表结构
    pocket.DB.AutoMigrate(&domain.User{})

    return &Module{
        pocket:  pocket,
        handler: userHandler,
    }
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    m.handler.RegisterRoutes(r)
}
```

## 核心优势

### 1. 依赖管理清晰
- **集中管理**：所有依赖通过 Pocket 统一管理
- **按需获取**：业务模块只需要什么依赖，就从 Pocket 取什么
- **易于测试**：可以轻松 Mock Pocket 中的依赖

### 2. 模块高度解耦
- 每个模块独立完整（domain/repository/service/api）
- 模块间无直接依赖，通过 Pocket 间接通信
- 可以轻松拆分为微服务

### 3. 初始化流程优雅
```go
// 一行代码完成所有初始化
pocket := core.NewPocketBuilder("config.yml").MustBuild()

// 可选：添加额外组件
pocket := core.NewPocketBuilder("config.yml")
    .WithEnforcer("model.conf", "policy.csv")
    .MustBuild()
```

### 4. 扩展性强
- 新增模块：只需实现 `Module` 接口
- 新增依赖：在 Pocket 中添加字段 + Builder 中添加初始化方法
- 新增中间件：在 Pocket 的 `initGin()` 中统一注册

## 配置文件

```yaml
server:
  name: nvo-api
  host: 0.0.0.0
  port: 8080
  mode: debug

database:
  driver: mysql
  host: localhost
  port: 3306
  database: nvo_db
  username: root
  password: password
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: localhost
  port: 6379
  db: 0

logger:
  level: info
  format: console
  output_path: stdout
```

## API 示例

启动服务后，可以访问以下接口：

```bash
# 用户注册
POST /api/v1/users/register
{
  "username": "test",
  "email": "test@example.com",
  "password": "123456"
}

# 用户登录
POST /api/v1/users/login
{
  "username": "test",
  "password": "123456"
}

# 获取用户信息
GET /api/v1/users/:id

# 用户列表
GET /api/v1/users?page=1&page_size=10
```

## 快速开始

1. **安装依赖**
```bash
go mod tidy
```

2. **配置数据库**
编辑 `config.yml`，配置数据库连接信息

3. **运行服务**
```bash
go run cmd/main.go
```

4. **测试接口**
```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'
```

## 最佳实践

### 1. 模块开发流程
1. 定义领域实体和接口（domain）
2. 实现数据访问层（repository）
3. 实现业务逻辑层（service）
4. 实现接口层（api/handler）
5. 组装模块（module.go）
6. 在 main.go 中注册模块

### 2. 依赖注入原则
- 业务模块只依赖 Pocket，不依赖具体实现
- Repository 依赖 DB 接口，不依赖 GORM
- Service 依赖 Repository 接口，不依赖具体实现
- Handler 依赖 Service，不依赖 Repository

### 3. 错误处理
- 使用 `fmt.Errorf` 包装错误，保留错误链
- 在 Service 层处理业务错误
- 在 Handler 层转换为 HTTP 响应

## 对比 go-zero

| 特性 | go-zero | NVO Pocket |
|------|---------|------------|
| 依赖管理 | ServiceContext | Pocket |
| 初始化方式 | 手动组装 | Builder 模式 |
| 配置管理 | config.Config | Viper + Config |
| 模块结构 | logic/svc/handler | domain/repository/service/api |
| 路由注册 | 自动生成 | 手动注册（更灵活） |

## 总结

这个脚手架设计的核心思想是：

1. **Pocket 作为依赖中心**：所有依赖统一管理，业务模块按需获取
2. **Builder 模式初始化**：优雅地完成复杂的初始化流程
3. **DDD 分层架构**：清晰的职责划分，易于维护和测试
4. **模块化设计**：高内聚低耦合，支持平滑演进到微服务

这种设计既保持了 go-zero 的简洁性，又提供了更灵活的扩展能力。
