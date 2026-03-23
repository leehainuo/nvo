package api

import (
	"errors"
	"strconv"

	"nvo-api/core/log"
	"nvo-api/internal/system/dict/domain"
	"nvo-api/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictHandler struct {
	service domain.DictService
}

func NewDictHandler(service domain.DictService) *DictHandler {
	return &DictHandler{
		service: service,
	}
}

// CreateDictType 创建字典类型
func (h *DictHandler) CreateDictType(c *gin.Context) {
	var req domain.CreateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	dictType, err := h.service.CreateDictType(&req)
	if err != nil {
		log.Error("create dict type failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, dictType)
}

// GetDictTypeByID 获取字典类型详情
func (h *DictHandler) GetDictTypeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	dictType, err := h.service.GetDictTypeByID(uint(id))
	if err != nil {
		log.Error("get dict type failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, dictType)
}

// UpdateDictType 更新字典类型
func (h *DictHandler) UpdateDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	var req domain.UpdateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateDictType(uint(id), &req); err != nil {
		log.Error("update dict type failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// DeleteDictType 删除字典类型
func (h *DictHandler) DeleteDictType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	if err := h.service.DeleteDictType(uint(id)); err != nil {
		log.Error("delete dict type failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ListDictTypes 获取字典类型列表
func (h *DictHandler) ListDictTypes(c *gin.Context) {
	var req domain.ListDictTypeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	dictTypes, total, err := h.service.ListDictTypes(&req)
	if err != nil {
		log.Error("list dict types failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Page(c, dictTypes, req.Page, req.PageSize, total)
}

// CreateDictData 创建字典数据
func (h *DictHandler) CreateDictData(c *gin.Context) {
	var req domain.CreateDictDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	dictData, err := h.service.CreateDictData(&req)
	if err != nil {
		log.Error("create dict data failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, dictData)
}

// GetDictDataByID 获取字典数据详情
func (h *DictHandler) GetDictDataByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	dictData, err := h.service.GetDictDataByID(uint(id))
	if err != nil {
		log.Error("get dict data failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, dictData)
}

// UpdateDictData 更新字典数据
func (h *DictHandler) UpdateDictData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	var req domain.UpdateDictDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	if err := h.service.UpdateDictData(uint(id), &req); err != nil {
		log.Error("update dict data failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// DeleteDictData 删除字典数据
func (h *DictHandler) DeleteDictData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errors.New("无效的ID"))
		return
	}

	if err := h.service.DeleteDictData(uint(id)); err != nil {
		log.Error("delete dict data failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ListDictData 获取字典数据列表
func (h *DictHandler) ListDictData(c *gin.Context) {
	var req domain.ListDictDataRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.New("参数错误: "+err.Error()))
		return
	}

	dictDataList, total, err := h.service.ListDictData(&req)
	if err != nil {
		log.Error("list dict data failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Page(c, dictDataList, req.Page, req.PageSize, total)
}

// GetDictDataByType 根据字典类型获取字典数据
func (h *DictHandler) GetDictDataByType(c *gin.Context) {
	dictType := c.Param("type")
	if dictType == "" {
		response.Error(c, errors.New("字典类型不能为空"))
		return
	}

	dictDataList, err := h.service.GetDictDataByType(dictType)
	if err != nil {
		log.Error("get dict data by type failed", zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, dictDataList)
}
