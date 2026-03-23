package api

import (
	"errors"
	"strconv"

	"nvo-api/core/log"
	"nvo-api/internal/system/user/domain"
	"nvo-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	service domain.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Create 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param body body domain.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	user, err := h.service.Create(&req)
	if err != nil {
		log.Error("create user failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// GetByID 获取用户详情
// @Summary 获取用户详情
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的用户ID"))
		return
	}

	user, err := h.service.GetByID(uint(id))
	if err != nil {
		log.Error("get user failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// Update 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param body body domain.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的用户ID"))
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.Update(uint(id), &req); err != nil {
		log.Error("update user failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的用户ID"))
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		log.Error("delete user failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param username query string false "用户名"
// @Param nickname query string false "昵称"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var req domain.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	users, total, err := h.service.List(&req)
	if err != nil {
		log.Error("list users failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Page(c, users, req.Page, req.PageSize, total)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param body body object true "修改密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的用户ID"))
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.ChangePassword(uint(id), req.OldPassword, req.NewPassword); err != nil {
		log.Error("change password failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}
