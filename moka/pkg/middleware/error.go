package middleware

import (
	"errors"

	"moka/pkg/apperr"

	"github.com/gin-gonic/gin"
)

// response is the unified error response envelope.
// Flat structure, no nesting — matches frontend expectations directly.
type response struct {
	RequestID string   `json:"request_id"`
	Code      string   `json:"code"`
	Msg       string   `json:"msg"`
	Err       string   `json:"err,omitempty"`   // unwrapped cause; debug only
	Stack     []string `json:"stack,omitempty"` // call frames; debug only
}

// httpStatus maps business error codes to HTTP status codes.
// Kept in the HTTP layer — the apperr package has no knowledge of HTTP.
var httpStatus = map[apperr.Code]int{
	apperr.Unauthenticated:    401,
	apperr.InvalidCredentials: 401,
	apperr.TokenExpired:       401,
	apperr.PermissionDenied:   403,
	apperr.NotFound:           404,
	apperr.AlreadyExists:      409,
	apperr.Conflict:           409,
	apperr.InvalidInput:       400,
	apperr.InvalidParam:       400,
	apperr.MissingParam:       400,
	apperr.RateLimit:          429,
	apperr.QuotaExhausted:     429,
	apperr.Timeout:            504,
}

func statusFor(code apperr.Code) int {
	if s, ok := httpStatus[code]; ok {
		return s
	}
	return 500
}

// Error is the global error-handling middleware. Register it last so it
// catches errors set by all other handlers via c.Error(err).
func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		handle(c, c.Errors.Last().Err)
	}
}

func handle(c *gin.Context, err error) {
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		// Untyped errors become internal errors; the original is wrapped
		// so it still appears in the "err" field during development.
		ae = apperr.Internal.Wrap("系统内部错误", err)
	}

	resp := response{
		Code: string(ae.Code()),
		Msg:  ae.Message(),
	}

	if rid, ok := c.Get("request_id"); ok {
		resp.RequestID = rid.(string)
	}

	if apperr.IsDebug() {
		if cause := ae.Unwrap(); cause != nil {
			resp.Err = cause.Error()
		}
		resp.Stack = ae.Frames()
	}

	c.JSON(statusFor(ae.Code()), resp)
}
