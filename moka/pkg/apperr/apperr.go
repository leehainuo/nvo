// Package apperr defines the application error type and business error codes.
//
// Design goals:
//   - Zero cost in production: stack traces are never captured when debug=false
//   - Lazy stack formatting: raw PCs stored, formatted only on serialization
//   - No stdlib naming conflict: package is "apperr", not "errors"
//   - Type-safe codes that double as constructors via method syntax
//
// Usage:
//
//	return nil, apperr.NotFound.New("用户不存在")
//	return nil, apperr.DBError.Wrap("查询失败", err)
package apperr

import (
	"fmt"
	"runtime"
	"strings"
)

// Code is a typed business error code. Values appear verbatim in API responses.
type Code string

// New creates an Error with this code.
//
//go:noinline
func (c Code) New(msg string) *Error {
	return newError(c, msg, nil)
}

// Wrap creates an Error wrapping an existing error.
//
//go:noinline
func (c Code) Wrap(msg string, cause error) *Error {
	return newError(c, msg, cause)
}

const (
	// Auth
	Unauthenticated    Code = "UNAUTHENTICATED"
	InvalidCredentials Code = "INVALID_CREDENTIALS"
	TokenExpired       Code = "TOKEN_EXPIRED"
	PermissionDenied   Code = "PERMISSION_DENIED"

	// Resource
	NotFound      Code = "NOT_FOUND"
	AlreadyExists Code = "ALREADY_EXISTS"
	Conflict      Code = "CONFLICT"

	// Input
	InvalidInput Code = "INVALID_INPUT"
	InvalidParam Code = "INVALID_PARAM"
	MissingParam Code = "MISSING_PARAM"

	// System
	Internal   Code = "INTERNAL_ERROR"
	DBError    Code = "DB_ERROR"
	RedisError Code = "REDIS_ERROR"
	Timeout    Code = "TIMEOUT"

	// Quota
	RateLimit      Code = "RATE_LIMIT"
	QuotaExhausted Code = "QUOTA_EXHAUSTED"
)

// debugMode is set once at startup via Init. Never mutated afterward.
var debugMode bool

// Init configures the package. Call once before serving requests.
func Init(debug bool) { debugMode = debug }

// IsDebug reports whether debug mode is active.
func IsDebug() bool { return debugMode }

// Error is the application error type.
type Error struct {
	code  Code
	msg   string
	cause error
	pcs   []uintptr // raw program counters; nil in production (debugMode=false)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.msg, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.msg)
}

// Unwrap supports errors.As / errors.Is chain traversal.
func (e *Error) Unwrap() error { return e.cause }

// Code returns the business error code.
func (e *Error) Code() Code { return e.code }

// Message returns the human-readable message.
func (e *Error) Message() string { return e.msg }

// Frames returns formatted stack frames. Resolves raw PCs only when called
// (lazy), so there is zero serialization cost when frames are not needed.
// Returns nil when debugMode is false.
func (e *Error) Frames() []string {
	if len(e.pcs) == 0 {
		return nil
	}
	frames := runtime.CallersFrames(e.pcs)
	result := make([]string, 0, 5)
	for {
		f, more := frames.Next()
		if f.File != "" && !isSkipped(f.Function) {
			result = append(result, fmt.Sprintf("%s:%d", trimPath(f.File), f.Line))
			if len(result) == 5 {
				break
			}
		}
		if !more {
			break
		}
	}
	return result
}

// newError is the single allocation point for all Error values.
// skip=4 accounts for: runtime.Callers → captureStack → newError → Code.New/Wrap
func newError(code Code, msg string, cause error) *Error {
	e := &Error{code: code, msg: msg, cause: cause}
	if debugMode {
		e.pcs = captureStack()
	}
	return e
}

// captureStack stores raw program counters. Cheap: no symbol resolution.
func captureStack() []uintptr {
	var pcs [16]uintptr
	n := runtime.Callers(4, pcs[:])
	return pcs[:n:n]
}

// skipped prefixes are framework / runtime frames that add no diagnostic value.
var skippedPrefixes = []string{
	"runtime.",
	"testing.",
	"github.com/gin-gonic/gin",
	"gorm.io/gorm",
	"net/http.",
}

func isSkipped(fn string) bool {
	for _, prefix := range skippedPrefixes {
		if strings.HasPrefix(fn, prefix) {
			return true
		}
	}
	return false
}

// trimPath keeps the last 3 path segments for compact, readable output.
func trimPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}
