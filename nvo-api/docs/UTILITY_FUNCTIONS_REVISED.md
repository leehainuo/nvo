# 🛠️ 通用工具函数建议（修订版）

## ✅ 已有的完善功能

### 1. **分页功能** ✅
```go
// pkg/response/response.go
func Page[T any](c *gin.Context, list T, page int, size int, total int64)
```
**状态**：已完整实现，无需重复

### 2. **数据验证** ✅
```go
// Gin 自带的 binding 验证
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=4,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Phone    string `json:"phone" binding:"omitempty,len=11"`
}
```
**状态**：Gin 框架已提供完善的验证功能，基本够用

---

## 🎯 真正需要的工具函数

### 🔴 高优先级

#### 1. **字符串工具** - `pkg/util/stringx/`

**问题**：代码中大量重复的字符串解析
```go
// 在 15+ 个 API Handler 中重复出现
id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
days, _ := strconv.Atoi(c.Query("days"))
```

**建议实现**：
```go
// pkg/util/stringx/stringx.go
package stringx

import "strconv"

// ParseUint 安全解析 uint
func ParseUint(s string) (uint, error) {
    val, err := strconv.ParseUint(s, 10, 32)
    return uint(val), err
}

// MustParseUint 解析 uint，失败返回 0
func MustParseUint(s string) uint {
    val, _ := ParseUint(s)
    return val
}

// ParseInt 安全解析 int
func ParseInt(s string) (int, error) {
    return strconv.Atoi(s)
}

// MustParseInt 解析 int，失败返回 0
func MustParseInt(s string) int {
    val, _ := ParseInt(s)
    return val
}

// IsEmpty 判断字符串是否为空（包括空白字符）
func IsEmpty(s string) bool {
    return len(strings.TrimSpace(s)) == 0
}

// DefaultIfEmpty 如果为空返回默认值
func DefaultIfEmpty(s, defaultValue string) string {
    if IsEmpty(s) {
        return defaultValue
    }
    return s
}

// MaskPhone 手机号掩码 (138****5678)
func MaskPhone(phone string) string {
    if len(phone) != 11 {
        return phone
    }
    return phone[:3] + "****" + phone[7:]
}

// MaskEmail 邮箱掩码 (abc***@gmail.com)
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return email
    }
    username := parts[0]
    if len(username) > 3 {
        username = username[:3] + "***"
    }
    return username + "@" + parts[1]
}
```

**使用效果**：
```go
// 之前：每个 Handler 都要写
id, err := strconv.ParseUint(c.Param("id"), 10, 32)
if err != nil {
    response.Error(c, errors.New("无效的ID"))
    return
}

// 之后：一行搞定
id := stringx.MustParseUint(c.Param("id"))
```

**预计时间**：1 小时  
**代码减少**：~200 行

---

#### 2. **密码工具** - `pkg/util/password/`

**问题**：密码加密逻辑重复
```go
// user/service/service.go 中重复出现
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

**建议实现**：
```go
// pkg/util/password/password.go
package password

import (
    "crypto/rand"
    "encoding/base64"
    "golang.org/x/crypto/bcrypt"
)

// Hash 加密密码
func Hash(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Verify 验证密码
func Verify(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

// Generate 生成随机密码
func Generate(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// Strength 密码强度等级
type Strength int

const (
    Weak Strength = iota
    Medium
    Strong
)

// CheckStrength 检查密码强度
func CheckStrength(password string) Strength {
    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )
    
    for _, char := range password {
        switch {
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case 'a' <= char && char <= 'z':
            hasLower = true
        case '0' <= char && char <= '9':
            hasNumber = true
        default:
            hasSpecial = true
        }
    }
    
    strength := 0
    if hasUpper { strength++ }
    if hasLower { strength++ }
    if hasNumber { strength++ }
    if hasSpecial { strength++ }
    
    if len(password) < 8 || strength < 2 {
        return Weak
    }
    if len(password) >= 12 && strength >= 3 {
        return Strong
    }
    return Medium
}
```

**使用效果**：
```go
// 之前
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    return nil, fmt.Errorf("密码加密失败: %w", err)
}

// 之后
hashedPassword, err := password.Hash(req.Password)
if err != nil {
    return nil, fmt.Errorf("密码加密失败: %w", err)
}

// 验证也更简洁
if password.Verify(user.Password, req.Password) {
    // 密码正确
}
```

**预计时间**：30 分钟  
**代码减少**：~50 行

---

#### 3. **时间工具** - `pkg/util/timex/`

**建议实现**：
```go
// pkg/util/timex/timex.go
package timex

import (
    "fmt"
    "time"
)

// 常用时间格式
const (
    DateFormat     = "2006-01-02"
    TimeFormat     = "15:04:05"
    DateTimeFormat = "2006-01-02 15:04:05"
)

// Today 今天零点
func Today() time.Time {
    now := time.Now()
    return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
    return t.Format(DateFormat)
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
    return t.Format(DateTimeFormat)
}

// ParseDate 解析日期
func ParseDate(s string) (time.Time, error) {
    return time.Parse(DateFormat, s)
}

// ParseDateTime 解析日期时间
func ParseDateTime(s string) (time.Time, error) {
    return time.Parse(DateTimeFormat, s)
}

// TimeAgo 多久之前（人性化显示）
func TimeAgo(t time.Time) string {
    duration := time.Since(t)
    
    if duration < time.Minute {
        return "刚刚"
    }
    if duration < time.Hour {
        return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
    }
    if duration < 24*time.Hour {
        return fmt.Sprintf("%d小时前", int(duration.Hours()))
    }
    if duration < 30*24*time.Hour {
        return fmt.Sprintf("%d天前", int(duration.Hours()/24))
    }
    if duration < 365*24*time.Hour {
        return fmt.Sprintf("%d个月前", int(duration.Hours()/24/30))
    }
    return fmt.Sprintf("%d年前", int(duration.Hours()/24/365))
}

// IsExpired 是否过期
func IsExpired(t time.Time) bool {
    return t.Before(time.Now())
}
```

**预计时间**：1 小时

---

### 🟡 中优先级

#### 4. **切片工具** - `pkg/util/slicex/`

**泛型工具函数**：
```go
// pkg/util/slicex/slicex.go
package slicex

// Contains 判断切片是否包含元素
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

// Unique 去重
func Unique[T comparable](slice []T) []T {
    seen := make(map[T]bool)
    result := make([]T, 0)
    for _, item := range slice {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }
    return result
}

// Filter 过滤
func Filter[T any](slice []T, predicate func(T) bool) []T {
    result := make([]T, 0)
    for _, item := range slice {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

// Map 映射
func Map[T any, R any](slice []T, mapper func(T) R) []R {
    result := make([]R, len(slice))
    for i, item := range slice {
        result[i] = mapper(item)
    }
    return result
}
```

**预计时间**：1 小时

---

#### 5. **JSON 工具** - `pkg/util/jsonx/`

```go
// pkg/util/jsonx/jsonx.go
package jsonx

import "encoding/json"

// Marshal 序列化（忽略错误）
func Marshal(v any) string {
    bytes, _ := json.Marshal(v)
    return string(bytes)
}

// MarshalIndent 格式化序列化
func MarshalIndent(v any) string {
    bytes, _ := json.MarshalIndent(v, "", "  ")
    return string(bytes)
}

// Unmarshal 反序列化
func Unmarshal(data string, v any) error {
    return json.Unmarshal([]byte(data), v)
}

// Clone 深拷贝
func Clone[T any](src T) (T, error) {
    var dst T
    bytes, err := json.Marshal(src)
    if err != nil {
        return dst, err
    }
    err = json.Unmarshal(bytes, &dst)
    return dst, err
}
```

**预计时间**：30 分钟

---

### 🟢 低优先级

#### 6. **随机工具** - `pkg/util/random/`

```go
// pkg/util/random/random.go
package random

import (
    "crypto/rand"
    "math/big"
)

const (
    Numbers = "0123456789"
    Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    All     = Numbers + Letters
)

// String 生成随机字符串
func String(length int, charset string) string {
    result := make([]byte, length)
    for i := range result {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        result[i] = charset[num.Int64()]
    }
    return string(result)
}

// Number 生成随机数字字符串
func Number(length int) string {
    return String(length, Numbers)
}

// Code 生成验证码
func Code(length int) string {
    return Number(length)
}
```

**预计时间**：30 分钟

---

## 📊 修订后的优先级

| 工具包 | 优先级 | 原因 | 预计时间 | 影响范围 |
|--------|--------|------|----------|---------|
| stringx | 🔴 高 | 15+ 处重复代码 | 1h | 所有 API Handler |
| password | 🔴 高 | 密码处理重复 | 30min | User Service |
| timex | 🔴 高 | 常用时间操作 | 1h | 多个模块 |
| slicex | 🟡 中 | 泛型工具 | 1h | 提升开发效率 |
| jsonx | 🟡 中 | JSON 操作 | 30min | 可选 |
| random | 🟢 低 | 随机生成 | 30min | 验证码等 |

---

## 🎯 实施建议

### 第一步：核心工具（本周）

1. **stringx** - 解决大量重复代码
2. **password** - 统一密码处理
3. **timex** - 常用时间操作

**总计**：2.5 小时，立即提升代码质量

### 第二步：辅助工具（按需）

4. **slicex** - 泛型工具函数
5. **jsonx** - JSON 操作
6. **random** - 随机生成

---

## 💡 总结

**已有功能**：
- ✅ 分页响应（response.Page）
- ✅ 数据验证（Gin binding）

**真正需要的**：
- 🔴 **stringx** - 解决 15+ 处重复的字符串解析
- 🔴 **password** - 统一密码加密/验证
- 🔴 **timex** - 常用时间操作

**建议**：优先实现这 3 个核心工具包，只需 2.5 小时，就能显著提升代码质量！
