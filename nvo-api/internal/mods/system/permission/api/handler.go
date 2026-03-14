package api

import (
	"net/http"

	"nvo-api/core/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	enforcer *auth.Enforcer
	logger   *zap.Logger
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(enforcer *auth.Enforcer, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		enforcer: enforcer,
		logger:   logger,
	}
}

// GetUserMenus 获取用户可访问的菜单列表
// @Summary 获取用户菜单
// @Description 返回当前用户有权限访问的菜单列表，前端根据此数据渲染导航菜单
// @Tags 权限
// @Produce json
// @Success 200 {object} MenuResponse
// @Router /api/v1/permissions/menus [get]
func (h *PermissionHandler) GetUserMenus(c *gin.Context) {
	// 从上下文获取用户ID（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	subject := "user:" + userID.(string)

	// 定义所有菜单（通常从数据库读取）
	allMenus := []Menu{
		{Code: "dashboard", Name: "工作台", Icon: "dashboard", Path: "/dashboard", Sort: 1},
		{Code: "system", Name: "系统管理", Icon: "setting", Sort: 100},
		{Code: "system.user", Name: "用户管理", Icon: "user", Path: "/system/user", ParentCode: "system", Sort: 101},
		{Code: "system.role", Name: "角色管理", Icon: "team", Path: "/system/role", ParentCode: "system", Sort: 102},
		{Code: "system.permission", Name: "权限管理", Icon: "lock", Path: "/system/permission", ParentCode: "system", Sort: 103},
	}

	// 过滤用户有权限的菜单
	var accessibleMenus []Menu
	for _, menu := range allMenus {
		ok, err := h.enforcer.CheckMenu(subject, menu.Code, "view")
		if err != nil {
			h.logger.Error("failed to check menu permission",
				zap.String("user", subject),
				zap.String("menu", menu.Code),
				zap.Error(err))
			continue
		}
		if ok {
			accessibleMenus = append(accessibleMenus, menu)
		}
	}

	// 构建树形结构
	menuTree := buildMenuTree(accessibleMenus)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"menus": menuTree,
		},
	})
}

// GetUserButtons 获取用户在指定页面的按钮权限
// @Summary 获取页面按钮权限
// @Description 返回当前用户在指定页面有权限的按钮列表，前端根据此数据控制按钮显示
// @Tags 权限
// @Param page query string true "页面标识" example(user)
// @Produce json
// @Success 200 {object} ButtonResponse
// @Router /api/v1/permissions/buttons [get]
func (h *PermissionHandler) GetUserButtons(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// 获取页面标识
	page := c.Query("page")
	if page == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少 page 参数",
		})
		return
	}

	subject := "user:" + userID.(string)

	// 定义该页面的所有按钮（通常从数据库读取）
	allButtons := getPageButtons(page)

	// 过滤用户有权限的按钮
	var accessibleButtons []Button
	for _, btn := range allButtons {
		ok, err := h.enforcer.CheckButton(subject, btn.Code, "click")
		if err != nil {
			h.logger.Error("failed to check button permission",
				zap.String("user", subject),
				zap.String("button", btn.Code),
				zap.Error(err))
			continue
		}
		if ok {
			accessibleButtons = append(accessibleButtons, btn)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"buttons": accessibleButtons,
		},
	})
}

// GetUserPermissions 获取用户的所有权限信息
// @Summary 获取用户权限信息
// @Description 返回用户的角色和所有权限，用于前端权限判断
// @Tags 权限
// @Produce json
// @Success 200 {object} PermissionResponse
// @Router /api/v1/permissions/user [get]
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	subject := "user:" + userID.(string)

	// 获取用户角色
	roles, err := h.enforcer.GetRolesForUser(subject)
	if err != nil {
		h.logger.Error("failed to get user roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色失败",
		})
		return
	}

	// 获取用户权限
	permissions, err := h.enforcer.GetPermissionsForUser(subject)
	if err != nil {
		h.logger.Error("failed to get user permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user_id":     userID,
			"roles":       roles,
			"permissions": permissions,
		},
	})
}

// Menu 菜单结构
type Menu struct {
	Code       string  `json:"code"`        // 菜单编码
	Name       string  `json:"name"`        // 菜单名称
	Icon       string  `json:"icon"`        // 图标
	Path       string  `json:"path"`        // 路由路径
	ParentCode string  `json:"parent_code"` // 父菜单编码
	Sort       int     `json:"sort"`        // 排序
	Children   []Menu  `json:"children,omitempty"` // 子菜单
}

// Button 按钮结构
type Button struct {
	Code string `json:"code"` // 按钮编码
	Name string `json:"name"` // 按钮名称
	Type string `json:"type"` // 按钮类型: primary, danger, default
	Icon string `json:"icon"` // 图标
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []Menu) []Menu {
	menuMap := make(map[string]*Menu)
	var roots []Menu

	// 第一遍：创建所有菜单的映射
	for i := range menus {
		menuMap[menus[i].Code] = &menus[i]
	}

	// 第二遍：构建树形结构
	for i := range menus {
		if menus[i].ParentCode == "" {
			// 根菜单
			roots = append(roots, menus[i])
		} else {
			// 子菜单
			if parent, ok := menuMap[menus[i].ParentCode]; ok {
				parent.Children = append(parent.Children, menus[i])
			}
		}
	}

	return roots
}

// getPageButtons 获取页面的所有按钮（示例，实际应从数据库读取）
func getPageButtons(page string) []Button {
	buttonMap := map[string][]Button{
		"user": {
			{Code: "user.create", Name: "新建", Type: "primary", Icon: "plus"},
			{Code: "user.edit", Name: "编辑", Type: "default", Icon: "edit"},
			{Code: "user.delete", Name: "删除", Type: "danger", Icon: "delete"},
			{Code: "user.export", Name: "导出", Type: "default", Icon: "download"},
			{Code: "user.import", Name: "导入", Type: "default", Icon: "upload"},
		},
		"role": {
			{Code: "role.create", Name: "新建", Type: "primary", Icon: "plus"},
			{Code: "role.edit", Name: "编辑", Type: "default", Icon: "edit"},
			{Code: "role.delete", Name: "删除", Type: "danger", Icon: "delete"},
			{Code: "role.assign", Name: "分配权限", Type: "primary", Icon: "key"},
		},
	}

	if buttons, ok := buttonMap[page]; ok {
		return buttons
	}
	return []Button{}
}
