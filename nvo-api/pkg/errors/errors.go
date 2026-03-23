package errors

import "fmt"

// Code - 错误码类型
type Code int

// 错误码定义
const (
	// 成功
	SUCCESS Code = 0

	// 业务错误 (3xxx)

)

// Error - 业务错误
type Error struct {
	Raw     error  `json:"-"`       // 原始错误
	Code    Code   `json:"code"`    // 业务错误码
	Message string `json:"message"` // 响应消息
}

// Error - 实现 error 接口
func (err *Error) Error() string {
	if err.Raw != nil {
		return fmt.Sprintf("[%d] %s: %v", err.Code, err.Message, err.Raw)
	}
	return fmt.Sprintf("[%d] %s", err.Code, err.Message)
}

// Unwrap - 支持错误链
func (err *Error) Unwrap() error {
	return err.Raw
}

// New - 创建业务错误
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap - 包装错误
func Wrap(code Code, message string, err error) *Error {
	return &Error{
		Raw:     err,
		Code:    code,
		Message: message,
	}
}
