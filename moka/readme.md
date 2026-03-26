# Go 最佳实践方法论

在 Go 的开发过程中，我在不断的思考和反复实践，得出了下面的一些心得，给自己书写了一套可复用的方法论.

对于我来说我觉得最优雅的架构设计如下:
 - 对于基础设施应该采用**全局单例设计**
 - 对于上层业务应该采用**依赖注入设计**

基础设施: MySQL、Redis、OSS ...

上层业务: Handler、Service ...

## 全局单例设计

全局单例的设计, 变量采用小写并通过方法返回, 这样不可被修改, 安全可靠.

同时 Go 编译器会对于简单的函数进行**内联优化**, 无需担心有额外的开销.

### 1 初始化模式选择

#### 1.1 饿汉式 (Eager Initialization)

在包加载时立即初始化, 使用 Go 的 `init()` 初始化.

**适用场景:**
 - 失败即致命, 无降级方案
 - 必定会被使用
 - 无外部依赖 (仅依赖常量或已经初始化的包级变量)
 - 初始化成本低

``` go
var client *Config

func init() {
    var c Config

    if err := config.Viper.UnmarshalKey("key", &c); err != nil {
        panic(err)
    }

    // ...

    client = &c
}

func Client() {
    return client
}
```

#### 1.2 懒汉式 (Lazy Initialization)

延迟到首次使用时初始化, 使用 `sync.Once` 保证线程安全.

**适用场景:**
 - 可能不会被使用 (按需加载)
 - 初始化成本高 (如大对象、网络连接)
 - 依赖运行时参数
 - 需要优雅降级

``` go
var (
    once   sync.Once
    client *Client
)

func Client() *Client {
    once.Do(func() {
        var c Config
        
        if err := config.Viper.UnmarshalKey("key", &c); err != nil {
            panic(err)
        }
        
        // ...
        
        client = &c
    })
    
    return client
}
```

### 2 初始化函数选择

#### 2.1 init() 函数

**适用场景:**
 - 无外部服务依赖
 - 失败即致命, 无需错误处理
   ``` go
   func init() {
        if condition {
            panic("panic")
        }
   }
   ```

**优点:** 自动执行, 代码简洁
**缺点:** 无法控制顺序, 无法返回错误, 难以测试

#### 2.2 Init() 手动初始化

**适用场景:**
 - 需要控制初始化时机
   ```go
   func main() {
       one.Init()
       two.Init()
       three.Init()
       // ...
   }
   ```
 - 需要优雅的错误处理
   ``` go
   if err := module.Init(); err != nil {
        // 降级方案
        // ...
   }
   ```
 - 有外部服务依赖

**优点:** 可控、可测试、可处理错误
**缺点:** 需要手动调用, 可能忘记初始化

### 3 最优雅的决策树

提示: 
 - 自动 `init()` 初始化函数就是名为 init, 手动 `Init()` 可以是名为其他的初始化函数 例: `initModule()`
 - 对于懒汉式的手动 `Init()` 其实在初始化简单的情况下不需要创建一个额外的手动初始化函数, 即省略 `Init()`


**决策流程图**

```mermaid
开始
  │
  ├─ 是否有外部服务依赖？(数据库/Redis/HTTP/文件系统)
  │   ├─ 是 → 手动 Init()
  │   └─ 否 → 继续
  │
  ├─ 初始化可能失败且需要降级处理？
  │   ├─ 是 → 手动 Init()
  │   └─ 否 → 继续
  │
  ├─ 需要控制初始化顺序？(依赖其他模块)
  │   ├─ 是 → 手动 Init()
  │   └─ 否 → 继续
  │
  ├─ 资源可能不会被使用？(可选功能)
  │   ├─ 是 → 懒汉式 sync.Once
  │   └─ 否 → 继续
  │
  ├─ 初始化成本很高？(>100ms 或大内存)
  │   ├─ 是 → 懒汉式 sync.Once
  │   └─ 否 → 饿汉式 init()
```

**快速决策表**

| 问题 | 答案 | 推荐方案 |
|------|------|----------|
| 依赖外部服务？ | 是 | 手动 Init() + 饿汉式 |
| 需要错误降级？ | 是 | 手动 Init() + 饿汉式 |
| 需要控制顺序？ | 是 | 手动 Init() + 饿汉式 |
| 可能不使用？  | 是 | 手动 Init() + 懒汉式 |
| 初始化成本高？ | 是 | 自动 init() + 懒汉式 |
| 以上都否 | - | 自动 init() + 饿汉式 |


**四象限分类**

```go
        自动 init()
            │
            │
懒汉式 ──────┼───── 饿汉式
            │
            │
        手动 Init()
```

象限 1: 自动 init() + 饿汉式
 - **场景:** 常量初始化、配置加载
 - **示例:** 正则表达式、JWT配置加载
```go
var (
    EmailRegex *regexp.Regexp
    PhoneRegex *regexp.Regexp
)

func init() {
    EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    PhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`) 
}

func ValidateEmail(email string) bool {
    return EmailRegex.MatchString(email)
}
```
```go
// pkg/util/jwt/jwt.go
var JWT *Config

type Config struct {
    Secret    string
    Issuer    string
    ExpiresIn int64
}

func init() {
    var c Config
    if err := config.Viper.UnmarshalKey("jwt", &c); err != nil {
        panic(fmt.Sprintf("failed to load jwt config: %v", err))
    }
    if c.Secret == "" {
        panic("jwt secret is required")
    }
    JWT = &c
}

func GnerateToken() {
    // maybe use JWT.Secret
    // ...
}

// Other func
// ...
```


象限 2: 自动 init() + 懒汉式
 - **场景:** 昂贵但必定执行的初始化
 - **示例:** 大型正则表达式编译
```go
var (
    once  sync.Once
    regex *regexp.Regexp
)

func init() {
    // ...
}

func Regex() *regexp.Regexp {
    once.Do(func() {
        regex = compile()
    })
    return regex
}
```

象限 3: 手动 Init() + 饿汉式
 - **场景:** 外部服务、有依赖链
 - **示例:** MySQL、Redis、Casbin
```go
var client *gorm.DB

func Init() error {
    
    // ...
    // maybe `return err`

    return nil
}

```

象限 4: 手动 Init() + 懒汉式
 - **场景:** 按需加载、可选服务
 - **示例:** 邮件服务、文件上传
```go
var (
    err    error
    once   sync.Once
    mailer *Mailer
)

func Mailer() (*Mailer, error) {
    once.Do(func() {
        mailer, err = initMailer()
    })
    return mailer, err
}
```

### 4 多实例扩展

为了之支持测试和特殊场景, 基础设计应提供工厂函数创建独立实例.

**示例:**
```go
var client *redis.Client

func Init() error {
    // ...
    return nil
}

func Client() *redis.Client {
    return client
}

// 多实例扩展
type Options func (*Config)

func New(opts ...Options) (*redis.Client, error) {
    c := &Config{}

    for _, opt := range opts {
        opt(c)
    }

    // vaild ...
    // maybe err ...

    client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
	})

    // ...
    
    return client, nil
}

// Options func
func WithHost(host string) Options { /* ... */}
func WithPort(port int) Options { /* ... */}
// ...
```


## 依赖注入设计

构建好基础设施了, 接下来看一下上层业务的最佳实践

### 1 编写Service层
```go
type XxxService struct {
    db *gorm.DB
}

func NewXxxService(db *gorm.DB) *XxxService {
    return &XxxService{
        db: db
    }
}

func (x *XxxService) One() {
    // ...
}
```

### 2 编写Handler层
```go
type XxxHandler struct {
    xxxService *XxxService
}

func NewXxxHandler(xxxService *XxxService) *XxxHandler {
    return &XxxHandler{
        xxxService: xxxService
    }
}

func (x *XxxHandler) One(c *gin.Context) {
    // ...
}
```

### 3 编写Router层并初始化依赖注入
```go
func InitXxxRouter(group *gin.RouterGroup) {
    // 依赖注入
    xxxService := NewXxxService(mysql.Client())
    xxxHandler := NewXxxHandler(xxxService)

    router := group.Group("/xxx")
    {
        router.GET("/one", xxxHandler.One)
        // ...
    }
}
```

### 4 编写总Router层初始化
```go
func Init() *gin.Engine {
    // ...

    router := gin.New()
    
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

    // ...
    
    group := router.Group("/api/v1")
    {
        InitXxxRouter(group)
        // ...
    }

    return router
}
```

## 测试实践


## 总结
