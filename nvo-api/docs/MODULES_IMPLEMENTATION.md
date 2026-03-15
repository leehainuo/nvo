# 6个核心模块实现方案

## 📋 模块清单

### ✅ 已实现模块（3个）
1. **User** - 用户模块
2. **Role** - 角色模块  
3. **Permission** - 权限模块

### 🚀 新增模块（3个）
4. **Menu** - 菜单模块（支持树形结构）
5. **Dept** - 部门模块（支持树形结构）
6. **Audit** - 审计日志模块

---

## 🏗️ 架构设计

### 三级架构
```
Pocket（基础设施层）
  ↓
SystemService（业务域聚合层）
  ├─ User, Role, Permission
  ├─ Menu, Dept
  └─ Audit
  ↓
具体服务层（UserService, MenuService...）
```

### 模块结构（统一规范）
```
internal/system/{module}/
├── domain/
│   ├── {module}.go      # 领域模型 + DTO
│   └── service.go       # 服务接口
├── repository/
│   └── {module}_repository.go  # 数据访问层
├── service/
│   └── {module}_service.go     # 业务逻辑层
├── api/
│   └── {module}_handler.go     # HTTP处理层
└── module.go            # 模块注册
```

---

## 📝 实现状态

### Menu 模块
- [x] Domain 模型（支持树形结构）
- [x] Service 接口
- [ ] Repository 实现
- [ ] Service 实现
- [ ] API Handler
- [ ] Module 注册

### Dept 模块
- [x] Domain 模型（支持树形结构）
- [x] Service 接口
- [ ] Repository 实现
- [ ] Service 实现
- [ ] API Handler
- [ ] Module 注册

### Audit 模块
- [x] Domain 模型（操作日志）
- [x] Service 接口
- [ ] Repository 实现
- [ ] Service 实现
- [ ] API Handler
- [ ] Module 注册

---

## 🎯 下一步

需要我继续实现以下内容吗？

1. **Repository 层** - 数据访问（CRUD + 树形查询）
2. **Service 层** - 业务逻辑（包含树形结构构建）
3. **API Handler** - HTTP 接口
4. **Module 注册** - 模块初始化和路由注册
5. **更新 system/index.go** - 注册新模块到 SystemService

请告诉我：
- 是否需要完整实现所有细节？
- 还是先实现核心功能，其他功能后续补充？
