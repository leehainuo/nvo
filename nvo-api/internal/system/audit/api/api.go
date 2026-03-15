package api

import (
	"net/http"
	"strconv"
	"nvo-api/internal/system/audit/domain"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service domain.AuditService
}

func NewAuditHandler(service domain.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) Create(c *gin.Context) {
	var req domain.CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功"})
}

func (h *AuditHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	log, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": log})
}

func (h *AuditHandler) GetList(c *gin.Context) {
	var req domain.ListAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.service.GetList(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

func (h *AuditHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *AuditHandler) CleanOldLogs(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if err := h.service.CleanOldLogs(days); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "清理成功"})
}
