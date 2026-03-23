# 代码优化指南

## 🎯 优化目标
让代码更加**优雅**、**美观**、**易于理解**、**易于维护**

---

## 1. 错误处理优化

### 1.1 当前问题
❌ **重复的错误检查代码**
```go
// 每个 Service 方法都重复这段代码
user, err := s.repo.GetByID(id)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, errors.New("用户不存在")
    }
    return nil, err
}
```

### 1.2 优化方案

#### 方案 A：统一错误定义（推荐）
```go
// pkg/errors/business.go
package errors

var (
    // 通用错误
    ErrRecordNotFound = New(40400, "记录不存在")
    ErrInvalidParams  = New(40000, "参数错误")
    
    // 用户相关
    ErrUserNotFound      = New(40401, "用户不存在")
    ErrUserExists        = New(40901, "用户已存在")
    ErrEmailExists       = New(40902, "邮箱已存在")
    ErrInvalidPassword   = New(40101, "密码错误")
    ErrUserDisabled      = New(40301, "账号已被禁用")
    
    // 角色相关
    ErrRoleNotFound = New(40402, "角色不存在")
    ErrRoleExists   = New(40903, "角色已存在")
    
    // 字典相关
    ErrDictTypeNotFound = New(40403, "字典类型不存在")
    ErrDictTypeExists   = New(40904, "字典类型已存在")
    ErrDictDataNotFound = New(40404, "字典数据不存在")
    
    // 菜单相关
    ErrMenuNotFound     = New(40405, "菜单不存在")
    ErrHasChildren      = New(40905, "存在子项，无法删除")
    
    // 部门相关
    ErrDeptNotFound = New(40406, "部门不存在")
)
```

**使用示例：**
```go
// ✅ 优雅的写法
func (s *UserService) GetByID(id uint) (*userDomain.UserResponse, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, pkgErrors.ErrUserNotFound
        }
        return nil, err
    }
    
    roles, _ := s.getRoles(user.ID)
    return user.ToResponse(roles), nil
}
```

#### 方案 B：Repository 层统一处理（更优雅）
```go
// internal/common/repository/base.go
package repository

import (
    "errors"
    "gorm.io/gorm"
    pkgErrors "nvo-api/pkg/errors"
)

// HandleNotFound 统一处理记录不存在错误
func HandleNotFound(err error, customErr error) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return customErr
    }
    return err
}

// 使用示例
func (s *UserService) GetByID(id uint) (*userDomain.UserResponse, error) {
    user, err := s.repo.GetByID(id)
    if err := repository.HandleNotFound(err, pkgErrors.ErrUserNotFound); err != nil {
        return nil, err
    }
    
    roles, _ := s.getRoles(user.ID)
    return user.ToResponse(roles), nil
}
```

---

## 2. 日志记录优化

### 2.1 当前问题
❌ **日志不一致，缺少关键信息**
```go
log.Error("create dict type failed", zap.Error(err))
log.Info("dict type created", zap.String("dict_type", dictType.DictType))
```

### 2.2 优化方案

#### 统一日志格式
```go
// ✅ 优雅的日志记录
func (s *UserService) Create(req *userDomain.CreateUserRequest) (*userDomain.User, error) {
    log.Info("creating user",
        zap.String("username", req.Username),
        zap.String("email", req.Email))
    
    user, err := s.createUser(req)
    if err != nil {
        log.Error("failed to create user",
            zap.String("username", req.Username),
            zap.Error(err))
        return nil, err
    }
    
    log.Info("user created successfully",
        zap.Uint("user_id", user.ID),
        zap.String("username", user.Username))
    
    return user, nil
}
```

#### 日志规范
- **Info**: 正常业务流程（创建、更新、删除成功）
- **Warn**: 业务异常但可恢复（验证失败、权限不足）
- **Error**: 系统错误（数据库错误、外部服务错误）
- **Debug**: 调试信息（仅开发环境）

---

## 3. 验证逻辑优化

### 3.1 当前问题
❌ **验证逻辑分散在 Service 层**
```go
func (s *UserService) Create(req *userDomain.CreateUserRequest) (*userDomain.User, error) {
    // 检查用户名
    exists, err := s.repo.ExistsByUsername(req.Username)
    if exists {
        return nil, errors.New("用户名已存在")
    }
    
    // 检查邮箱
    if req.Email != "" {
        exists, err := s.repo.ExistsByEmail(req.Email)
        if exists {
            return nil, errors.New("邮箱已存在")
        }
    }
    // ...
}
```

### 3.2 优化方案

#### 提取验证方法
```go
// service/service.go
func (s *UserService) validateCreate(req *userDomain.CreateUserRequest) error {
    // 检查用户名
    if exists, _ := s.repo.ExistsByUsername(req.Username); exists {
        return pkgErrors.ErrUserExists
    }
    
    // 检查邮箱
    if req.Email != "" {
        if exists, _ := s.repo.ExistsByEmail(req.Email); exists {
            return pkgErrors.ErrEmailExists
        }
    }
    
    return nil
}

func (s *UserService) Create(req *userDomain.CreateUserRequest) (*userDomain.User, error) {
    // ✅ 清晰的验证步骤
    if err := s.validateCreate(req); err != nil {
        return nil, err
    }
    
    // 业务逻辑
    return s.createUser(req)
}
```

---

## 4. 代码复用优化

### 4.1 当前问题
❌ **相似的代码模式重复**
```go
// 多个 Service 都有类似的 getRoles 逻辑
subject := fmt.Sprintf("user:%d", user.ID)
roles, _ := s.enforcer.GetRolesForUser(subject)
```

### 4.2 优化方案

#### 提取公共方法
```go
// service/service.go
func (s *UserService) getUserRoles(userID uint) []string {
    subject := fmt.Sprintf("user:%d", userID)
    roles, _ := s.enforcer.GetRolesForUser(subject)
    return roles
}

// 使用
func (s *UserService) GetByID(id uint) (*userDomain.UserResponse, error) {
    user, err := s.repo.GetByID(id)
    if err := repository.HandleNotFound(err, pkgErrors.ErrUserNotFound); err != nil {
        return nil, err
    }
    
    return user.ToResponse(s.getUserRoles(user.ID)), nil
}
```

---

## 5. 函数命名优化

### 5.1 当前问题
❌ **命名不够语义化**
```go
func (s *UserService) batchGetUserRoles(users []*userDomain.User) map[uint][]string
```

### 5.2 优化方案
```go
// ✅ 更清晰的命名
func (s *UserService) loadUserRolesMap(users []*userDomain.User) map[uint][]string
func (s *UserService) buildRolesMap(users []*userDomain.User) map[uint][]string
```

**命名规范：**
- `Get` - 获取单个资源
- `List` - 获取列表
- `Load` - 加载关联数据
- `Build` - 构建/组装数据
- `Validate` - 验证
- `Handle` - 处理

---

## 6. 注释优化

### 6.1 当前问题
❌ **缺少必要的注释或注释过于简单**
```go
func (s *UserService) List(req *userDomain.ListUserRequest) ([]*userDomain.UserResponse, int64, error) {
    users, total, err := s.repo.List(req)
    // ...
}
```

### 6.2 优化方案
```go
// List 获取用户列表
// 该方法会批量加载用户的角色信息，避免 N+1 查询问题
// 参数:
//   - req: 列表查询请求，包含分页和过滤条件
// 返回:
//   - []*UserResponse: 用户响应列表
//   - int64: 总记录数
//   - error: 错误信息
func (s *UserService) List(req *userDomain.ListUserRequest) ([]*userDomain.UserResponse, int64, error) {
    users, total, err := s.repo.List(req)
    if err != nil {
        return nil, 0, err
    }
    
    if len(users) == 0 {
        return []*userDomain.UserResponse{}, 0, nil
    }
    
    // 批量加载角色，避免 N+1 查询
    rolesMap := s.buildRolesMap(users)
    
    return userDomain.ToResponseList(users, rolesMap), total, nil
}
```

---

## 7. 常量提取优化

### 7.1 当前问题
❌ **魔法数字和字符串**
```go
if user.Status != 1 {
    return errors.New("账号已被禁用")
}

subject := fmt.Sprintf("user:%d", userID)
```

### 7.2 优化方案
```go
// internal/common/constants/status.go
package constants

const (
    StatusEnabled  int8 = 1
    StatusDisabled int8 = 0
)

const (
    SubjectPrefixUser = "user"
    SubjectPrefixRole = "role"
)

// 使用
if user.Status != constants.StatusEnabled {
    return pkgErrors.ErrUserDisabled
}

subject := fmt.Sprintf("%s:%d", constants.SubjectPrefixUser, userID)
```

---

## 8. 结构体初始化优化

### 8.1 当前问题
❌ **字段赋值分散**
```go
dictType := &dictDomain.DictType{
    DictName: req.DictName,
    DictType: req.DictType,
    Status:   req.Status,
    Remark:   req.Remark,
}
```

### 8.2 优化方案
```go
// ✅ 使用构造函数
// domain/dict.go
func NewDictType(req *CreateDictTypeRequest) *DictType {
    return &DictType{
        DictName: req.DictName,
        DictType: req.DictType,
        Status:   req.Status,
        Remark:   req.Remark,
    }
}

// service/service.go
func (s *DictService) CreateDictType(req *dictDomain.CreateDictTypeRequest) (*dictDomain.DictType, error) {
    if err := s.validateDictType(req); err != nil {
        return nil, err
    }
    
    dictType := dictDomain.NewDictType(req)
    if err := s.repo.CreateDictType(dictType); err != nil {
        log.Error("failed to create dict type", zap.Error(err))
        return nil, err
    }
    
    log.Info("dict type created", zap.String("type", dictType.DictType))
    return dictType, nil
}
```

---

## 9. 链式调用优化

### 9.1 优化方案
```go
// ✅ 使用链式方法提高可读性
type UserQueryBuilder struct {
    repo *UserRepository
    conditions map[string]interface{}
}

func (b *UserQueryBuilder) WithUsername(username string) *UserQueryBuilder {
    if username != "" {
        b.conditions["username"] = username
    }
    return b
}

func (b *UserQueryBuilder) WithStatus(status *int8) *UserQueryBuilder {
    if status != nil {
        b.conditions["status"] = *status
    }
    return b
}

func (b *UserQueryBuilder) Build() *gorm.DB {
    query := b.repo.db.Model(&User{})
    for key, val := range b.conditions {
        query = query.Where(key+" = ?", val)
    }
    return query
}

// 使用
users := NewUserQueryBuilder(repo).
    WithUsername("admin").
    WithStatus(&enabled).
    Build().
    Find(&users)
```

---

## 10. 测试友好性优化

### 10.1 依赖注入
```go
// ✅ 使用接口，便于 Mock
type UserRepository interface {
    GetByID(id uint) (*User, error)
    Create(user *User) error
    // ...
}

type UserService struct {
    repo UserRepository  // 接口类型，便于测试
    enforcer *casbin.SyncedEnforcer
}
```

---

## 📊 优化优先级

| 优先级 | 优化项 | 影响范围 | 工作量 |
|--------|--------|----------|--------|
| 🔴 高 | 统一错误定义 | 全局 | 中 |
| 🔴 高 | ToResponse 模式 | 全局 | 已完成 |
| 🟡 中 | 日志规范化 | 全局 | 低 |
| 🟡 中 | 提取验证方法 | Service 层 | 中 |
| 🟢 低 | 常量提取 | 全局 | 低 |
| 🟢 低 | 注释完善 | 全局 | 低 |

---

## 🎯 实施建议

### 阶段一：基础优化（立即实施）
1. ✅ 统一错误定义
2. ✅ 提取常量
3. ✅ 规范日志

### 阶段二：结构优化（逐步实施）
1. 提取验证方法
2. 提取公共方法
3. 完善注释

### 阶段三：高级优化（可选）
1. 构造函数模式
2. 链式调用
3. 查询构建器

---

**遵循这些优化建议，代码将更加优雅、易读、易维护！**
