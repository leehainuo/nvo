# 架构设计规范

## 1. 模块化设计原则

### 1.1 目录结构
每个业务模块应遵循以下目录结构：
```
module/
├── domain/          # 领域层
│   ├── entity.go    # 实体定义
│   └── service.go   # 服务接口
├── repository/      # 数据访问层
│   └── repository.go
├── service/         # 业务逻辑层
│   └── service.go
├── api/            # API 处理层
│   └── api.go
└── module.go       # 模块注册
```

### 1.2 领域模型设计

#### 实体（Entity）
- 使用 GORM 标签定义数据库映射
- 包含业务字段和时间戳
- 实现 `TableName()` 方法指定表名

#### 请求对象（Request）
- `CreateXxxRequest` - 创建请求
- `UpdateXxxRequest` - 更新请求
- `ListXxxRequest` - 列表查询请求
- 使用 `binding` 标签进行参数验证

#### 响应对象（Response）
- `XxxResponse` - 单个实体响应
- 包含需要返回给前端的字段
- **不包含敏感信息**（如密码）

## 2. 响应转换模式（ToResponse Pattern）

### 2.1 设计理念
**职责分离**：将实体到响应的转换逻辑归属于领域模型，而非业务逻辑层。

### 2.2 实现规范

#### 单个实体转换
```go
// ToResponse 将实体转换为响应对象
func (e *Entity) ToResponse() *EntityResponse {
    return &EntityResponse{
        ID:        e.ID,
        Field1:    e.Field1,
        Field2:    e.Field2,
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}
```

#### 批量转换（无额外依赖）
```go
// ToEntityResponseList 批量转换实体列表
func ToEntityResponseList(entities []*Entity) []*EntityResponse {
    responses := make([]*EntityResponse, 0, len(entities))
    for _, entity := range entities {
        responses = append(responses, entity.ToResponse())
    }
    return responses
}
```

#### 批量转换（有额外依赖）
```go
// ToEntityResponseList 批量转换实体列表（带关联数据）
func ToEntityResponseList(entities []*Entity, relatedDataMap map[uint][]string) []*EntityResponse {
    responses := make([]*EntityResponse, 0, len(entities))
    for _, entity := range entities {
        responses = append(responses, entity.ToResponse(relatedDataMap[entity.ID]))
    }
    return responses
}
```

### 2.3 Service 层使用示例

**优雅的写法** ✅
```go
func (s *Service) GetByID(id uint) (*EntityResponse, error) {
    entity, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    return entity.ToResponse(), nil
}

func (s *Service) List(req *ListRequest) ([]*EntityResponse, int64, error) {
    entities, total, err := s.repo.List(req)
    if err != nil {
        return nil, 0, err
    }
    return ToEntityResponseList(entities), total, nil
}
```

**冗余的写法** ❌
```go
func (s *Service) GetByID(id uint) (*EntityResponse, error) {
    entity, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    // 手动构建响应 - 代码重复，难以维护
    return &EntityResponse{
        ID:        entity.ID,
        Field1:    entity.Field1,
        Field2:    entity.Field2,
        CreatedAt: entity.CreatedAt,
        UpdatedAt: entity.UpdatedAt,
    }, nil
}
```

## 3. 优势总结

### 3.1 代码简洁性
- Service 层代码减少 70%+
- 单行转换替代 10+ 行手动构建

### 3.2 可维护性
- 转换逻辑集中管理
- 字段变更只需修改一处
- 减少人为错误

### 3.3 可测试性
- 转换逻辑可独立测试
- Mock 更加简单

### 3.4 一致性
- 所有模块遵循相同模式
- 新人易于理解和上手

## 4. 命名规范

### 4.1 方法命名
- 单个转换：`ToResponse()`
- 批量转换：`To{Entity}ResponseList()`
- 带参数转换：`ToResponse(extraData)`

### 4.2 包级函数
- 批量转换函数放在 `domain` 包中
- 函数名以 `To` 开头，清晰表达意图

## 5. 适用场景

### 5.1 必须使用
- ✅ 所有需要返回给前端的数据
- ✅ 包含关联数据的响应
- ✅ 需要过滤敏感字段的场景

### 5.2 可选使用
- 内部服务间调用（可直接使用实体）
- 简单的 CRUD 操作（但建议统一使用）

## 6. 示例对比

### Before（不推荐）
```go
// Service 层充斥着重复的转换代码
responses := make([]*UserResponse, 0, len(users))
for _, user := range users {
    responses = append(responses, &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        // ... 10+ 行重复代码
    })
}
```

### After（推荐）
```go
// Service 层简洁清晰
return ToUserResponseList(users, rolesMap), total, nil
```

---

**遵循这些规范，确保整个脚手架的工程质量和架构一致性！**
