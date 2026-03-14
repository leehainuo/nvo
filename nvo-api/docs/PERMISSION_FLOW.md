# 权限校验流程设计

## 🎯 三级权限的不同校验方式

### 1. API 权限 - 后端强制校验 ✅

**位置**：后端中间件
**目的**：安全防护，防止未授权访问
**方式**：强制拦截

```go
// 在路由中使用中间件
router.Use(auth.APIAuthMiddleware(enforcer, logger))

// 流程
用户请求 → 认证中间件 → API权限中间件 → 业务处理
                ↓              ↓
            设置user_id    检查API权限
                              ↓
                          无权限返回403
```

**特点**：
- ✅ **强制性** - 必须通过才能访问
- ✅ **安全性** - 后端控制，无法绕过
- ✅ **实时性** - 每次请求都检查

---

### 2. 菜单权限 - 前端控制显示 ✅

**位置**：前端导航组件
**目的**：用户体验，只显示可访问的菜单
**方式**：前端获取权限数据后渲染

```javascript
// 前端流程
1. 用户登录成功
   ↓
2. 调用 GET /api/v1/permissions/menus
   ↓
3. 后端返回用户有权限的菜单列表
   ↓
4. 前端根据菜单列表渲染导航

// Vue 示例
<template>
  <el-menu>
    <el-menu-item 
      v-for="menu in userMenus" 
      :key="menu.code"
      :index="menu.path">
      {{ menu.name }}
    </el-menu-item>
  </el-menu>
</template>

<script>
export default {
  data() {
    return {
      userMenus: []
    }
  },
  async mounted() {
    // 获取用户菜单
    const res = await this.$http.get('/api/v1/permissions/menus')
    this.userMenus = res.data.menus
  }
}
</script>
```

**特点**：
- ✅ **用户友好** - 不显示无权限的菜单
- ✅ **减少请求** - 避免用户点击后才发现无权限
- ⚠️ **非强制** - 仅控制显示，后端仍需校验 API

---

### 3. 按钮权限 - 前端控制显示 ✅

**位置**：前端页面组件
**目的**：用户体验，只显示可操作的按钮
**方式**：前端获取权限数据后渲染

```javascript
// 前端流程
1. 进入某个页面（如用户管理）
   ↓
2. 调用 GET /api/v1/permissions/buttons?page=user
   ↓
3. 后端返回该页面用户有权限的按钮
   ↓
4. 前端根据按钮列表渲染操作按钮

// Vue 示例
<template>
  <div>
    <!-- 只显示有权限的按钮 -->
    <el-button 
      v-if="hasButton('user.create')"
      type="primary"
      @click="handleCreate">
      新建
    </el-button>
    
    <el-button 
      v-if="hasButton('user.delete')"
      type="danger"
      @click="handleDelete">
      删除
    </el-button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      buttons: []
    }
  },
  async mounted() {
    // 获取页面按钮权限
    const res = await this.$http.get('/api/v1/permissions/buttons', {
      params: { page: 'user' }
    })
    this.buttons = res.data.buttons
  },
  methods: {
    hasButton(code) {
      return this.buttons.some(btn => btn.code === code)
    }
  }
}
</script>
```

**特点**：
- ✅ **精细控制** - 按钮级别的权限
- ✅ **用户友好** - 不显示无权限的按钮
- ⚠️ **非强制** - 仅控制显示，后端仍需校验 API

---

## 🔒 双重保护机制

### 前端 + 后端双重校验

```
前端（菜单/按钮）          后端（API）
     ↓                        ↓
  控制显示                  强制拦截
     ↓                        ↓
  用户体验                  安全保障
```

**示例：删除用户**

```javascript
// 前端 - 控制按钮显示
<el-button 
  v-if="hasButton('user.delete')"  // ← 前端权限控制
  @click="deleteUser">
  删除
</el-button>

async deleteUser(id) {
  // 即使按钮显示了，后端仍会校验
  await this.$http.delete(`/api/v1/users/${id}`)
}
```

```go
// 后端 - API 强制校验
router.DELETE("/users/:id", 
    auth.APIAuthMiddleware(enforcer, logger),  // ← 后端强制校验
    handler.DeleteUser)

func (h *UserHandler) DeleteUser(c *gin.Context) {
    // 可选：二次检查按钮权限（更严格）
    userID := c.GetString("user_id")
    ok, _ := h.enforcer.CheckButton("user:"+userID, "user.delete", "click")
    if !ok {
        c.JSON(403, gin.H{"message": "无删除权限"})
        return
    }
    
    // 执行删除
    // ...
}
```

---

## 📊 完整流程图

```
┌─────────────────────────────────────────────────────────────┐
│                        用户登录                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              前端获取权限数据（一次性）                       │
│  GET /api/v1/permissions/menus    → 渲染导航菜单             │
│  GET /api/v1/permissions/buttons  → 控制按钮显示             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                   用户操作（点击按钮）                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              前端发起 API 请求                                │
│         POST /api/v1/users                                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              后端 API 权限中间件校验                          │
│  enforcer.CheckAPI(user, "/api/v1/users", "POST")           │
│         ↓ 有权限                    ↓ 无权限                 │
│    执行业务逻辑                  返回 403                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 最佳实践

### 1. 菜单权限
```go
// ✅ 推荐：前端控制显示
GET /api/v1/permissions/menus
→ 前端只渲染有权限的菜单

// ❌ 不推荐：后端拦截路由
// 用户体验差，点击后才发现无权限
```

### 2. 按钮权限
```go
// ✅ 推荐：前端控制显示 + 后端 API 校验
前端：v-if="hasButton('user.delete')"
后端：APIAuthMiddleware 检查 API 权限

// ⚠️ 可选：后端二次检查按钮权限（更严格）
if !enforcer.CheckButton(user, "user.delete", "click") {
    return 403
}
```

### 3. API 权限
```go
// ✅ 必须：后端强制校验
router.Use(auth.APIAuthMiddleware(enforcer, logger))

// ❌ 错误：只在前端校验
// 前端校验可以被绕过，不安全
```

---

## 📝 总结

| 权限类型 | 校验位置 | 校验方式 | 目的 |
|---------|---------|---------|------|
| **API** | 后端中间件 | 强制拦截 | 安全保障 |
| **菜单** | 前端组件 | 控制显示 | 用户体验 |
| **按钮** | 前端组件 | 控制显示 | 用户体验 |

**核心原则：**
- 🔒 **安全靠后端** - API 权限必须后端校验
- 🎨 **体验靠前端** - 菜单/按钮权限前端控制显示
- 🛡️ **双重保护** - 前端控制 + 后端校验

你的理解完全正确！🎉
