package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	zl "moka/pkg/util/log"
)

func Recovery(stack bool) gin.HandlerFunc {
	logger := log.New(os.Stderr, "\n\n\x1b[31m", log.LstdFlags)

	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				var isBrokenPipe bool
				err, ok := rec.(error)
				if ok {
					isBrokenPipe = errors.Is(err, syscall.EPIPE) ||
						errors.Is(err, syscall.ECONNRESET) ||
						errors.Is(err, http.ErrAbortHandler)
				}

				var msg string

				if isBrokenPipe {
					msg = fmt.Sprintf("%s\n%s%s", rec, secureRequestDump(c.Request), "\033[0m")
					logger.Print(msg)

					if gin.Mode() != gin.DebugMode {
						zl.Error(msg)
					}

					c.Error(err)
					c.Abort()
					return
				}

				if stack {
					msg = fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s",
						timeFormat(time.Now()),
						secureRequestDump(c.Request),
						rec,
						string(debug.Stack()),
						"\033[0m")
				} else {
					msg = fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s%s",
						timeFormat(time.Now()),
						rec,
						string(debug.Stack()),
						"\033[0m")
				}

				logger.Print(msg)

				if gin.Mode() == gin.ReleaseMode {
					zl.Error(msg)
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// timeFormat 返回格式化的时间字符串
func timeFormat(t time.Time) string {
	return t.Format("2006-01-02 - 15:04:05")
}

// secureRequestDump 返回安全的 HTTP 请求转储，隐藏 Authorization 头
func secureRequestDump(r *http.Request) string {
	httpRequest, _ := httputil.DumpRequest(r, false)
	lines := strings.Split(string(httpRequest), "\r\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Authorization:") {
			lines[i] = "Authorization: *"
		}
	}
	return strings.Join(lines, "\r\n")
}
