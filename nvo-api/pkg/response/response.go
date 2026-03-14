package response

import (
	"net/http"
	"nvo-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response - 统一响应结构
type Response struct {
	Code    errors.Code `json:"code"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
}

// PageData - 分页数据
type PageData struct {
	List  any   `json:"list"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// Success - 成功响应
func Success[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// SuccessWithMessage - 成功响应 (自定义消息)
func SuccessWithMessage[T any](c *gin.Context, message string, data T) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Page - 分页响应
func Page[T any](c *gin.Context, list T, page int, size int, total int64) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data: PageData{
			List:  list,
			Page:  page,
			Size:  size,
			Total: total,
		},
	})
}

// Error - 错误响应
func Error(c *gin.Context, err error) {
	// 判断错误类型
	if err, ok := err.(*errors.Error); ok {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    err.Code,
			Message: err.Message,
		})
		return
	}

	// 未知错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}

// ErrorWithStatus - 错误响应 (指定 HTTP 状态码)
func ErrorWithStatus(c *gin.Context, httpStatus int, err error) {
	// 判断是否为业务错误
	if err, ok := err.(*errors.Error); ok {
		c.JSON(httpStatus, Response{
			Code:    err.Code,
			Message: err.Message,
		})
		return
	}

	// 普通错误，使用 HTTP 状态码作为业务错误码
	c.JSON(httpStatus, Response{
		Code:    errors.Code(httpStatus),
		Message: err.Error(),
	})
}
