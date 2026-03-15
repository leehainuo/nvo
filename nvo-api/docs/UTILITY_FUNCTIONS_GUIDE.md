# 🛠️ 通用工具函数建议

## 📊 当前状态分析

### ✅ 已有工具

**pkg/response/** - 响应工具
- ✅ Success - 成功响应
- ✅ Page - 分页响应
- ✅ Error - 错误响应

**pkg/errors/** - 错误处理
- ✅ 自定义错误类型
- ✅ 错误包装

**pkg/util/jwt/** - JWT 工具
- ⚠️ 未实现（TODO）

**pkg/util/storage/** - 存储工具
- ⚠️ 空文件

---

## 🎯 推荐添加的通用工具函数

### 🔴 高优先级（必须）

#### 1. **字符串工具** - `pkg/util/stringx/`

**问题**：代码中大量重复的字符串操作
```go
// 当前代码中重复出现
strconv.ParseUint(c.Param("id"), 10, 32)  // 出现 15+ 次
strings.Split(code, ":")                   // 多次出现
```

**建议实现**：
```go
// pkg/util/stringx/stringx.go
package stringx

import (
    "strconv"
    "strings"
    "unicode"
)

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

// IsNotEmpty 判断字符串是否非空
func IsNotEmpty(s string) bool {
    return !IsEmpty(s)
}

// DefaultIfEmpty 如果为空返回默认值
func DefaultIfEmpty(s, defaultValue string) string {
    if IsEmpty(s) {
        return defaultValue
    }
    return s
}

// Truncate 截断字符串（支持中文）
func Truncate(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen]) + "..."
}

// CamelToSnake 驼峰转蛇形
func CamelToSnake(s string) string {
    var result []rune
    for i, r := range s {
        if unicode.IsUpper(r) {
            if i > 0 {
                result = append(result, '_')
            }
            result = append(result, unicode.ToLower(r))
        } else {
            result = append(result, r)
        }
    }
    return string(result)
}

// SnakeToCamel 蛇形转驼峰
func SnakeToCamel(s string) string {
    parts := strings.Split(s, "_")
    for i := range parts {
        if len(parts[i]) > 0 {
            parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
        }
    }
    return strings.Join(parts, "")
}

// Contains 判断字符串是否包含子串（忽略大小写）
func ContainsIgnoreCase(s, substr string) bool {
    return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Mask 掩码处理（如手机号、邮箱）
func Mask(s string, start, end int, mask rune) string {
    runes := []rune(s)
    if start < 0 || end > len(runes) || start >= end {
        return s
    }
    for i := start; i < end; i++ {
        runes[i] = mask
    }
    return string(runes)
}

// MaskPhone 手机号掩码 (138****5678)
func MaskPhone(phone string) string {
    if len(phone) != 11 {
        return phone
    }
    return Mask(phone, 3, 7, '*')
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

**使用示例**：
```go
// 之前
id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

// 之后
id := stringx.MustParseUint(c.Param("id"))
```

---

#### 2. **密码工具** - `pkg/util/password/`

**问题**：密码加密逻辑重复
```go
// 当前代码中重复出现
bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
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

// Strength 检查密码强度
type StrengthLevel int

const (
    Weak StrengthLevel = iota
    Medium
    Strong
)

func CheckStrength(password string) StrengthLevel {
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
    if hasUpper {
        strength++
    }
    if hasLower {
        strength++
    }
    if hasNumber {
        strength++
    }
    if hasSpecial {
        strength++
    }
    
    if len(password) < 8 || strength < 2 {
        return Weak
    }
    if len(password) >= 12 && strength >= 3 {
        return Strong
    }
    return Medium
}
```

**使用示例**：
```go
// 之前
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

// 之后
hashedPassword, err := password.Hash(req.Password)
```

---

#### 3. **时间工具** - `pkg/util/timex/`

**建议实现**：
```go
// pkg/util/timex/timex.go
package timex

import (
    "time"
)

// 常用时间格式
const (
    DateFormat     = "2006-01-02"
    TimeFormat     = "15:04:05"
    DateTimeFormat = "2006-01-02 15:04:05"
    ISO8601Format  = "2006-01-02T15:04:05Z07:00"
)

// Now 当前时间
func Now() time.Time {
    return time.Now()
}

// Today 今天零点
func Today() time.Time {
    now := time.Now()
    return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow 明天零点
func Tomorrow() time.Time {
    return Today().AddDate(0, 0, 1)
}

// Yesterday 昨天零点
func Yesterday() time.Time {
    return Today().AddDate(0, 0, -1)
}

// StartOfWeek 本周开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
    weekday := int(t.Weekday())
    if weekday == 0 {
        weekday = 7
    }
    return t.AddDate(0, 0, -weekday+1)
}

// EndOfWeek 本周结束时间（周日）
func EndOfWeek(t time.Time) time.Time {
    return StartOfWeek(t).AddDate(0, 0, 7).Add(-time.Second)
}

// StartOfMonth 本月开始时间
func StartOfMonth(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 本月结束时间
func EndOfMonth(t time.Time) time.Time {
    return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Second)
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

// DaysBetween 计算两个日期之间的天数
func DaysBetween(start, end time.Time) int {
    return int(end.Sub(start).Hours() / 24)
}

// IsToday 是否是今天
func IsToday(t time.Time) bool {
    now := time.Now()
    return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsExpired 是否过期
func IsExpired(t time.Time) bool {
    return t.Before(time.Now())
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
```

---

#### 4. **验证工具** - `pkg/util/validator/`

**建议实现**：
```go
// pkg/util/validator/validator.go
package validator

import (
    "regexp"
    "unicode/utf8"
)

var (
    emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
    idCardRegex   = regexp.MustCompile(`^\d{17}[\dXx]$`)
    usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{4,20}$`)
    urlRegex      = regexp.MustCompile(`^https?://[^\s]+$`)
)

// IsEmail 验证邮箱
func IsEmail(email string) bool {
    return emailRegex.MatchString(email)
}

// IsPhone 验证手机号（中国）
func IsPhone(phone string) bool {
    return phoneRegex.MatchString(phone)
}

// IsIDCard 验证身份证号（中国）
func IsIDCard(idCard string) bool {
    return idCardRegex.MatchString(idCard)
}

// IsUsername 验证用户名（4-20位字母数字下划线）
func IsUsername(username string) bool {
    return usernameRegex.MatchString(username)
}

// IsURL 验证 URL
func IsURL(url string) bool {
    return urlRegex.MatchString(url)
}

// IsIP 验证 IP 地址
func IsIP(ip string) bool {
    ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
    if !ipRegex.MatchString(ip) {
        return false
    }
    
    parts := regexp.MustCompile(`\.`).Split(ip, -1)
    for _, part := range parts {
        var num int
        fmt.Sscanf(part, "%d", &num)
        if num < 0 || num > 255 {
            return false
        }
    }
    return true
}

// InRange 验证数字范围
func InRange(value, min, max int) bool {
    return value >= min && value <= max
}

// MinLength 验证最小长度
func MinLength(s string, min int) bool {
    return utf8.RuneCountInString(s) >= min
}

// MaxLength 验证最大长度
func MaxLength(s string, max int) bool {
    return utf8.RuneCountInString(s) <= max
}

// LengthBetween 验证长度范围
func LengthBetween(s string, min, max int) bool {
    length := utf8.RuneCountInString(s)
    return length >= min && length <= max
}

// IsChineseName 验证中文姓名
func IsChineseName(name string) bool {
    chineseRegex := regexp.MustCompile(`^[\p{Han}]{2,10}$`)
    return chineseRegex.MatchString(name)
}
```

---

### 🟡 中优先级（建议）

#### 5. **切片工具** - `pkg/util/slicex/`

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

// Chunk 分块
func Chunk[T any](slice []T, size int) [][]T {
    var chunks [][]T
    for i := 0; i < len(slice); i += size {
        end := i + size
        if end > len(slice) {
            end = len(slice)
        }
        chunks = append(chunks, slice[i:end])
    }
    return chunks
}

// Difference 差集
func Difference[T comparable](slice1, slice2 []T) []T {
    set := make(map[T]bool)
    for _, item := range slice2 {
        set[item] = true
    }
    
    result := make([]T, 0)
    for _, item := range slice1 {
        if !set[item] {
            result = append(result, item)
        }
    }
    return result
}

// Intersection 交集
func Intersection[T comparable](slice1, slice2 []T) []T {
    set := make(map[T]bool)
    for _, item := range slice2 {
        set[item] = true
    }
    
    result := make([]T, 0)
    for _, item := range slice1 {
        if set[item] {
            result = append(result, item)
        }
    }
    return Unique(result)
}
```

---

#### 6. **JSON 工具** - `pkg/util/jsonx/`

```go
// pkg/util/jsonx/jsonx.go
package jsonx

import (
    "encoding/json"
)

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

// ToMap 转换为 map
func ToMap(v any) (map[string]any, error) {
    bytes, err := json.Marshal(v)
    if err != nil {
        return nil, err
    }
    
    var result map[string]any
    err = json.Unmarshal(bytes, &result)
    return result, err
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

---

#### 7. **分页工具** - `pkg/util/pagination/`

```go
// pkg/util/pagination/pagination.go
package pagination

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// Params 分页参数
type Params struct {
    Page     int    `form:"page" json:"page"`
    PageSize int    `form:"page_size" json:"page_size"`
    OrderBy  string `form:"order_by" json:"order_by"`
    Sort     string `form:"sort" json:"sort"` // asc/desc
}

// Result 分页结果
type Result struct {
    List     any   `json:"list"`
    Total    int64 `json:"total"`
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
    Pages    int   `json:"pages"`
}

// Parse 从请求中解析分页参数
func Parse(c *gin.Context) *Params {
    params := &Params{
        Page:     1,
        PageSize: 10,
        Sort:     "desc",
    }
    c.ShouldBindQuery(params)
    
    if params.Page < 1 {
        params.Page = 1
    }
    if params.PageSize < 1 || params.PageSize > 100 {
        params.PageSize = 10
    }
    
    return params
}

// Paginate GORM 分页
func Paginate(params *Params) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        offset := (params.Page - 1) * params.PageSize
        
        query := db.Offset(offset).Limit(params.PageSize)
        
        if params.OrderBy != "" {
            order := params.OrderBy
            if params.Sort == "asc" {
                order += " asc"
            } else {
                order += " desc"
            }
            query = query.Order(order)
        }
        
        return query
    }
}

// NewResult 创建分页结果
func NewResult(list any, total int64, params *Params) *Result {
    pages := int(total) / params.PageSize
    if int(total)%params.PageSize > 0 {
        pages++
    }
    
    return &Result{
        List:     list,
        Total:    total,
        Page:     params.Page,
        PageSize: params.PageSize,
        Pages:    pages,
    }
}
```

---

#### 8. **随机工具** - `pkg/util/random/`

```go
// pkg/util/random/random.go
package random

import (
    "crypto/rand"
    "math/big"
)

const (
    Numbers    = "0123456789"
    LowerChars = "abcdefghijklmnopqrstuvwxyz"
    UpperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    AllChars   = Numbers + LowerChars + UpperChars
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

// Int 生成随机整数 [min, max]
func Int(min, max int) int {
    if min >= max {
        return min
    }
    num, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
    return int(num.Int64()) + min
}
```

---

### 🟢 低优先级（可选）

#### 9. **文件工具** - `pkg/util/filex/`
#### 10. **加密工具** - `pkg/util/crypto/`
#### 11. **ID 生成器** - `pkg/util/idgen/`（雪花算法、UUID）
#### 12. **HTTP 客户端** - `pkg/util/httpx/`

---

## 📋 实施建议

### 立即实施（本周）

1. **字符串工具** - 解决大量重复代码
2. **密码工具** - 统一密码处理
3. **验证工具** - 增强数据验证

### 短期实施（本月）

4. **时间工具** - 常用时间操作
5. **切片工具** - 泛型工具函数
6. **分页工具** - 统一分页逻辑

### 长期规划

7. **JSON 工具**
8. **随机工具**
9. **其他工具**

---

## 🎯 使用效果对比

### 之前（重复代码）

```go
// API Handler 中大量重复
id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

// Service 中重复
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

### 之后（优雅简洁）

```go
// API Handler
id := stringx.MustParseUint(c.Param("id"))

// Service
hashedPassword, err := password.Hash(req.Password)
if password.Verify(user.Password, req.Password) {
    // 密码正确
}
```

---

## 📊 优先级总结

| 工具包 | 优先级 | 原因 | 预计时间 |
|--------|--------|------|----------|
| stringx | 🔴 高 | 大量重复代码 | 1h |
| password | 🔴 高 | 安全性重要 | 30min |
| validator | 🔴 高 | 数据验证必须 | 1h |
| timex | 🟡 中 | 常用功能 | 1h |
| slicex | 🟡 中 | 提升开发效率 | 1h |
| pagination | 🟡 中 | 统一分页 | 1h |
| jsonx | 🟢 低 | 可选 | 30min |
| random | 🟢 低 | 可选 | 30min |

---

**总结**：优先实现高优先级的 3 个工具包，可以立即提升代码质量和开发效率！
