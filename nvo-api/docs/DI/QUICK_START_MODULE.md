# 模块开发快速指南

## 🚀 5分钟创建新模块

本指南帮助你快速在脚手架中创建新模块，遵循最佳实践。

---

## 📋 目录结构

每个模块遵循统一的目录结构：

```
internal/system/{module}/
├── domain/
│   ├── {module}.go      # 领域模型 + DTO
│   └── service.go       # 服务接口
├── repository/
│   └── {module}_repository.go
├── service/
│   └── {module}_service.go
├── api/
│   └── {module}_handler.go
└── module.go            # 模块注册
```

---

## 📝 创建步骤

### 场景 1：无依赖模块（最简单）

假设创建一个 **Product（商品）** 模块，不依赖其他业务模块。

#### Step 1: 创建目录结构

```bash
mkdir -p internal/business/product/{domain,repository,service,api}
```

#### Step 2: 定义领域模型

```go
// internal/business/product/domain/product.go
package domain

import (
    "time"
    "gorm.io/gorm"
)

type Product struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    Name        string         `gorm:"size:100;not null" json:"name"`
    Price       float64        `gorm:"type:decimal(10,2)" json:"price"`
    Stock       int            `gorm:"default:0" json:"stock"`
    Status      int8           `gorm:"default:1" json:"status"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
    return "products"
}

type CreateProductRequest struct {
    Name   string  `json:"name" binding:"required,max=100"`
    Price  float64 `json:"price" binding:"required,min=0"`
    Stock  int     `json:"stock" binding:"min=0"`
    Status int8    `json:"status" binding:"oneof=0 1"`
}

type UpdateProductRequest struct {
    Name   string   `json:"name" binding:"max=100"`
    Price  *float64 `json:"price" binding:"omitempty,min=0"`
    Stock  *int     `json:"stock" binding:"omitempty,min=0"`
    Status *int8    `json:"status" binding:"omitempty,oneof=0 1"`
}
```

#### Step 3: 定义服务接口

```go
// internal/business/product/domain/service.go
package domain

type ProductService interface {
    Create(req *CreateProductRequest) (*Product, error)
    GetByID(id uint) (*Product, error)
    Update(id uint, req *UpdateProductRequest) error
    Delete(id uint) error
    GetList() ([]*Product, error)
}
```

#### Step 4: 实现 Repository

```go
// internal/business/product/repository/product_repository.go
package repository

import (
    "nvo-api/internal/business/product/domain"
    "gorm.io/gorm"
)

type ProductRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
    return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *domain.Product) error {
    return r.db.Create(product).Error
}

func (r *ProductRepository) GetByID(id uint) (*domain.Product, error) {
    var product domain.Product
    err := r.db.First(&product, id).Error
    return &product, err
}

func (r *ProductRepository) Update(product *domain.Product) error {
    return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
    return r.db.Delete(&domain.Product{}, id).Error
}

func (r *ProductRepository) GetAll() ([]*domain.Product, error) {
    var products []*domain.Product
    err := r.db.Find(&products).Error
    return products, err
}
```

#### Step 5: 实现 Service

```go
// internal/business/product/service/product_service.go
package service

import (
    "errors"
    "nvo-api/internal/business/product/domain"
    "nvo-api/internal/business/product/repository"
    "gorm.io/gorm"
)

type ProductService struct {
    db   *gorm.DB
    repo *repository.ProductRepository
}

func NewProductService(db *gorm.DB) domain.ProductService {
    return &ProductService{
        db:   db,
        repo: repository.NewProductRepository(db),
    }
}

func (s *ProductService) Create(req *domain.CreateProductRequest) (*domain.Product, error) {
    product := &domain.Product{
        Name:   req.Name,
        Price:  req.Price,
        Stock:  req.Stock,
        Status: req.Status,
    }
    
    if err := s.repo.Create(product); err != nil {
        return nil, err
    }
    return product, nil
}

func (s *ProductService) GetByID(id uint) (*domain.Product, error) {
    product, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("商品不存在")
        }
        return nil, err
    }
    return product, nil
}

func (s *ProductService) Update(id uint, req *domain.UpdateProductRequest) error {
    product, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("商品不存在")
        }
        return err
    }
    
    if req.Name != "" {
        product.Name = req.Name
    }
    if req.Price != nil {
        product.Price = *req.Price
    }
    if req.Stock != nil {
        product.Stock = *req.Stock
    }
    if req.Status != nil {
        product.Status = *req.Status
    }
    
    return s.repo.Update(product)
}

func (s *ProductService) Delete(id uint) error {
    if _, err := s.repo.GetByID(id); err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("商品不存在")
        }
        return err
    }
    return s.repo.Delete(id)
}

func (s *ProductService) GetList() ([]*domain.Product, error) {
    return s.repo.GetAll()
}
```

#### Step 6: 实现 API Handler

```go
// internal/business/product/api/product_handler.go
package api

import (
    "net/http"
    "strconv"
    "nvo-api/internal/business/product/domain"
    "github.com/gin-gonic/gin"
)

type ProductHandler struct {
    service domain.ProductService
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(c *gin.Context) {
    var req domain.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    product, err := h.service.Create(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    product, err := h.service.GetByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) Update(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    var req domain.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.service.Update(uint(id), &req); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *ProductHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    if err := h.service.Delete(uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *ProductHandler) GetList(c *gin.Context) {
    products, err := h.service.GetList()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": products})
}
```

#### Step 7: 创建模块注册

```go
// internal/business/product/module.go
package product

import (
    "nvo-api/core"
    "nvo-api/internal/business/product/api"
    "nvo-api/internal/business/product/domain"
    "nvo-api/internal/business/product/service"
    "github.com/gin-gonic/gin"
)

type Module struct {
    pocket  *core.Pocket
    service domain.ProductService
    handler *api.ProductHandler
}

func NewModule(pocket *core.Pocket) *Module {
    productService := service.NewProductService(pocket.DB)
    productHandler := api.NewProductHandler(productService)
    
    return &Module{
        pocket:  pocket,
        service: productService,
        handler: productHandler,
    }
}

func (m *Module) Service() domain.ProductService {
    return m.service
}

func (m *Module) Name() string {
    return "product"
}

func (m *Module) Models() []any {
    return []any{
        &domain.Product{},
    }
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    products := r.Group("/products")
    {
        products.POST("", m.handler.Create)
        products.GET("", m.handler.GetList)
        products.GET("/:id", m.handler.GetByID)
        products.PUT("/:id", m.handler.Update)
        products.DELETE("/:id", m.handler.Delete)
    }
}
```

#### Step 8: 注册到系统

```go
// internal/business/index.go
package business

import (
    "nvo-api/core"
    "nvo-api/internal"
    "nvo-api/internal/business/product"
    "github.com/gin-gonic/gin"
)

func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 初始化模块
    productModule := product.NewModule(p)
    
    // 聚合到 BusinessService
    p.Business = internal.NewBusinessService(
        productModule.Service(),
    )
    
    // 数据库迁移
    modules := []internal.Module{productModule}
    migrateModels(p.DB, modules)
    
    // 注册路由
    for _, module := range modules {
        module.RegisterRoutes(r)
    }
}
```

---

### 场景 2：有依赖模块（单向依赖）

假设创建 **Order（订单）** 模块，依赖 **Product** 和 **User**。

#### 关键差异

**Service 构造函数**：

```go
// internal/business/order/service/order_service.go
package service

import (
    productDomain "nvo-api/internal/business/product/domain"
    userDomain "nvo-api/internal/system/user/domain"
)

type OrderService struct {
    db             *gorm.DB
    repo           *repository.OrderRepository
    productService productDomain.ProductService  // ✅ 依赖接口
    userService    userDomain.UserService        // ✅ 依赖接口
}

func NewOrderService(
    db *gorm.DB,
    productService productDomain.ProductService,
    userService userDomain.UserService,
) domain.OrderService {
    return &OrderService{
        db:             db,
        repo:           repository.NewOrderRepository(db),
        productService: productService,
        userService:    userService,
    }
}
```

**Module 初始化**：

```go
// internal/business/order/module.go
func NewModule(pocket *core.Pocket) *Module {
    orderService := service.NewOrderService(
        pocket.DB,
        pocket.Business.Product,  // ✅ 从聚合服务获取
        pocket.System.User,       // ✅ 跨域依赖
    )
    
    return &Module{service: orderService}
}
```

**系统注册**：

```go
// internal/business/index.go
func RegisterModules(r *gin.RouterGroup, p *core.Pocket) {
    // 阶段 1：无依赖模块
    productModule := product.NewModule(p)
    
    // 阶段 2：聚合
    p.Business = internal.NewBusinessService(
        productModule.Service(),
        nil,  // orderService 稍后注入
    )
    
    // 阶段 3：有依赖模块
    orderModule := order.NewModule(p)  // ✅ 此时依赖已可用
    p.Business.Order = orderModule.Service()
}
```

---

## 🎯 检查清单

创建新模块时，确保：

- [ ] 目录结构符合规范
- [ ] 接口定义在 `domain/service.go`
- [ ] Service 依赖接口，不依赖实现
- [ ] 构造函数显式声明所有依赖
- [ ] Module 实现 `internal.Module` 接口
- [ ] 正确注册到对应的聚合服务
- [ ] 路由注册正确
- [ ] 数据库模型迁移正确

---

## 💡 常用模板

### DTO 模板

```go
type CreateXxxRequest struct {
    Field1 string `json:"field1" binding:"required,max=100"`
    Field2 int    `json:"field2" binding:"min=0"`
}

type UpdateXxxRequest struct {
    Field1 string `json:"field1" binding:"max=100"`
    Field2 *int   `json:"field2" binding:"omitempty,min=0"`
}

type XxxResponse struct {
    ID        uint      `json:"id"`
    Field1    string    `json:"field1"`
    Field2    int       `json:"field2"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Repository 模板

```go
type XxxRepository struct {
    db *gorm.DB
}

func NewXxxRepository(db *gorm.DB) *XxxRepository {
    return &XxxRepository{db: db}
}

func (r *XxxRepository) Create(xxx *domain.Xxx) error {
    return r.db.Create(xxx).Error
}

func (r *XxxRepository) GetByID(id uint) (*domain.Xxx, error) {
    var xxx domain.Xxx
    err := r.db.First(&xxx, id).Error
    return &xxx, err
}
```

### Handler 模板

```go
type XxxHandler struct {
    service domain.XxxService
}

func NewXxxHandler(service domain.XxxService) *XxxHandler {
    return &XxxHandler{service: service}
}

func (h *XxxHandler) Create(c *gin.Context) {
    var req domain.CreateXxxRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    result, err := h.service.Create(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

---

## 📚 相关文档

- [依赖管理最佳实践](./DEPENDENCY_MANAGEMENT.md)
- [架构设计文档](./ARCHITECTURE.md)

---

**快速开始，遵循规范，构建企业级应用！** 🚀
