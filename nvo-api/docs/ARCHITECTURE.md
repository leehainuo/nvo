# 架构设计：完全解耦的配置管理

## 设计理念

**核心思想**：让每个客户端包自己负责读取配置，避免 `core/config` 对 `pkg/client/*` 的反向依赖。

## 架构图

```
┌─────────────────────────────────────────────────────────┐
│                    core/config                          │
│  - 只管理核心配置 (Server, Logger, Auth)                │
│  - 返回 viper 实例给其他包使用                          │
│  - 不依赖任何客户端包 ✅                                 │
└─────────────────────────────────────────────────────────┘
                           │
                           │ 返回 *viper.Viper
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   core/pocket                           │
│  - 接收 viper 实例                                       │
│  - 调用各客户端的 InitFromViper()                       │
└─────────────────────────────────────────────────────────┘
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
        ▼                                     ▼
┌──────────────────┐                 ┌──────────────────┐
│ pkg/client/mysql │                 │ pkg/client/redis │
│ - Config 结构    │                 │ - Config 结构    │
│ - InitDB()       │                 │ - InitRedis()    │
│ - InitFromViper()│                 │ - InitFromViper()│
└──────────────────┘                 └──────────────────┘
```

## 关键设计点

### 1. 配置结构解耦

**之前（有依赖）：**
```go
// core/config/config.go
import (
    "nvo-api/pkg/client/mysql"  // ❌ 反向依赖
    "nvo-api/pkg/client/redis"  // ❌ 反向依赖
)

type Config struct {
    Database mysql.Config  // 依赖客户端包
    Redis    redis.Config  // 依赖客户端包
}
```

**现在（完全解耦）：**
```go
// core/config/config.go
type Config struct {
    Server ServerConfig  // ✅ 只包含核心配置
    Logger LoggerConfig
    Auth   AuthConfig
}

func LoadConfig(path string) (*Config, *viper.Viper, error) {
    // 返回 viper 实例，让各包自己读取
}
```

### 2. 客户端包自治

每个客户端包都有：

**配置结构 (config.go)：**
```go
// pkg/client/mysql/config.go
package mysql

type Config struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Database string `mapstructure:"database"`
    // ...
}
```

**初始化方法 (mysql.go)：**
```go
// pkg/client/mysql/mysql.go
package mysql

// 方式1：直接传入配置
func InitDB(c Config) (*gorm.DB, error) { ... }

// 方式2：从 Viper 读取配置（推荐）
func InitFromViper(v *viper.Viper, key string) (*gorm.DB, error) {
    var cfg Config
    v.UnmarshalKey(key, &cfg)
    return InitDB(cfg)
}
```

### 3. Pocket 使用方式

```go
// core/pocket/pocket.go
func (b *PocketBuilder) loadConfig() error {
    cfg, v, err := core.LoadConfig(b.configPath)
    b.config = cfg
    b.viper = v  // 保存 viper 实例
    return nil
}

func (b *PocketBuilder) initDB() error {
    // 让 mysql 包自己从 viper 读取配置
    db, err := mysql.InitFromViper(b.viper, "database")
    b.pocket.DB = db
    return nil
}

func (b *PocketBuilder) initRedis() error {
    // 让 redis 包自己从 viper 读取配置
    rdb, err := redis.InitFromViper(b.viper, "redis")
    b.pocket.Redis = rdb
    return nil
}
```

## 优势对比

| 特性 | 之前的设计 | 现在的设计 |
|------|-----------|-----------|
| **依赖方向** | core/config → pkg/client/* ❌ | core/config ← pkg/client/* ✅ |
| **配置管理** | 集中在 core/config | 各包自己管理 |
| **扩展性** | 新增客户端需修改 core/config | 新增客户端无需修改 core |
| **独立性** | 客户端包依赖 core | 客户端包完全独立 |
| **复用性** | 客户端包难以复用 | 客户端包可独立复用 |

## 目录结构

```
nvo-api/
├── core/
│   ├── config/
│   │   └── config.go          # 核心配置（Server, Logger, Auth）
│   ├── logger/
│   │   └── logger.go          # 日志初始化
│   └── pocket/
│       └── pocket.go          # 依赖注入容器
│
├── pkg/client/
│   ├── mysql/
│   │   ├── config.go          # MySQL 配置结构
│   │   └── mysql.go           # MySQL 初始化（含 InitFromViper）
│   └── redis/
│       ├── config.go          # Redis 配置结构
│       └── redis.go           # Redis 初始化（含 InitFromViper）
│
└── config.yml                 # 配置文件
```

## 配置文件示例

```yaml
# 核心配置（由 core/config 管理）
server:
  name: nvo-api
  host: 0.0.0.0
  port: 8080
  mode: debug

logger:
  level: info
  format: console
  output_path: stdout

# 客户端配置（由各客户端包自己读取）
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
  password: ""
  db: 0
  pool_size: 10
```

## 使用示例

### 在 Pocket 中使用

```go
pocket := core.NewPocketBuilder("config.yml").MustBuild()
// 自动完成：
// 1. 加载核心配置
// 2. 初始化 Logger
// 3. MySQL 包从 viper 读取 "database" 配置
// 4. Redis 包从 viper 读取 "redis" 配置
```

### 独立使用客户端包

```go
// 客户端包可以独立使用，不依赖 core
import (
    "github.com/spf13/viper"
    "your-project/pkg/client/mysql"
)

v := viper.New()
v.SetConfigFile("config.yml")
v.ReadInConfig()

// 方式1：直接使用配置
cfg := mysql.Config{
    Host: "localhost",
    Port: 3306,
}
db, _ := mysql.InitDB(cfg)

// 方式2：从 viper 读取
db, _ := mysql.InitFromViper(v, "database")
```

## 扩展新客户端

添加新客户端（如 MongoDB）只需：

1. **创建客户端包**
```bash
mkdir -p pkg/client/mongodb
```

2. **定义配置**
```go
// pkg/client/mongodb/config.go
package mongodb

type Config struct {
    URI      string `mapstructure:"uri"`
    Database string `mapstructure:"database"`
}
```

3. **实现初始化**
```go
// pkg/client/mongodb/mongodb.go
package mongodb

func InitFromViper(v *viper.Viper, key string) (*mongo.Client, error) {
    var cfg Config
    v.UnmarshalKey(key, &cfg)
    // 初始化逻辑
}
```

4. **在 Pocket 中使用**
```go
func (b *PocketBuilder) initMongoDB() error {
    client, err := mongodb.InitFromViper(b.viper, "mongodb")
    b.pocket.MongoDB = client
    return nil
}
```

**无需修改 `core/config`！** ✅

## 总结

这种设计实现了：

1. ✅ **完全解耦**：core 不依赖 pkg，依赖方向正确
2. ✅ **高内聚**：每个客户端包管理自己的配置和初始化
3. ✅ **易扩展**：新增客户端无需修改核心代码
4. ✅ **可复用**：客户端包可以在任何项目中独立使用
5. ✅ **类型安全**：保持强类型，不使用 interface{}

**这是一个优雅且可维护的架构设计！**
