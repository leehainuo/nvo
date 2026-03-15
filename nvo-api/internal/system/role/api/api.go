package api

import (
	"errors"
	"strconv"

	"nvo-api/core/log"
	"nvo-api/internal/system/role/domain"
	"nvo-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	service domain.RoleService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(service domain.RoleService) *RoleHandler {
	return &RoleHandler{
		service: service,
	}
}

// Create 创建角色
// @Summary 创建角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param body body domain.CreateRoleRequest true "创建角色请求"
// @Success 200 {object} response.Response
// @Router /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req domain.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	role, err := h.service.Create(&req)
	if err != nil {
		log.Error("create role failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, role)
}

// GetByID 获取角色详情
// @Summary 获取角色详情
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的角色ID"))
		return
	}

	role, err := h.service.GetByID(uint(id))
	if err != nil {
		log.Error("get role failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, role)
}

// Update 更新角色
// @Summary 更新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param body body domain.UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的角色ID"))
		return
	}

	var req domain.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(uint(id), &req); err != nil {
		log.Error("update role failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除角色
// @Summary 删除角色
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的角色ID"))
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		log.Error("delete role failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取角色列表
// @Summary 获取角色列表
// @Tags 角色管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param name query string false "角色名称"
// @Param code query string false "角色编码"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var req domain.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	roles, total, err := h.service.List(&req)
	if err != nil {
		log.Error("list roles failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Page(c, roles, req.Page, req.PageSize, total)
}

// AssignPermissions 为角色分配权限
// @Summary 分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param body body domain.AssignPermissionsRequest true "权限列表"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的角色ID"))
		return
	}

	var req domain.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.AssignPermissions(uint(id), &req); err != nil {
		log.Error("assign permissions failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "权限分配成功"})
}

// GetPermissions 获取角色权限
// @Summary 获取角色权限
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/v1/roles/{id}/permissions [get]
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的角色ID"))
		return
	}

	permissions, err := h.service.GetPermissions(uint(id))
	if err != nil {
		log.Error("get permissions failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"permissions": permissions})
}

// GetAll 获取所有角色
// @Summary 获取所有角色
// @Tags 角色管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/roles/all [get]
func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.service.GetAll()
	if err != nil {
		log.Error("get all roles failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}
