package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	zl "moka/pkg/util/log"
)

const (
	blue    = "\033[97;44m"
	cyan    = "\033[97;46m"
	green   = "\033[97;42m"
	magenta = "\033[97;45m"
	red     = "\033[97;41m"
	reset   = "\033[0m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
)

func Logger() gin.HandlerFunc {
	return LoggerWithWriter(os.Stdout)
}

func LoggerWithWriter(out io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path  := c.Request.URL.Path
		raw   := c.Request.URL.RawQuery

		c.Next()

		timestamp  := time.Now()
		latency    := timestamp.Sub(start)
		clientIP   := c.ClientIP()
		method     := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 颜色
		statusColor  := colorForStatus(statusCode)
		methodColor  := colorForMethod(method)
		latencyColor := colorForLatency(latency)

		// 格式化输出
		fmt.Fprintf(out, "[Moka] %v |%s %3d %s|%s %8v %s| %15s |%s %-7s %s\033[32m %#v\033[0m\n",
			timestamp.Format("2006-01-02 - 15:04:05"),
			statusColor, statusCode, reset,
			latencyColor, latency, reset,
			clientIP,
			methodColor, method, reset,
			path,
		)

		if gin.Mode() == gin.ReleaseMode {
			msg := fmt.Sprintf("%s | %d | %v | %s | %s %s",
				timestamp.Format("2006-01-02 - 15:04:05"),
				statusCode,
				latency,
				clientIP,
				method,
				path,
			)
			zl.Info(msg)
		}
	}
}

// colorForLatency 返回延迟时间对应的颜色
func colorForLatency(latency time.Duration) string {
	switch {
	case latency < 100*time.Millisecond:
		return white
	case latency < 200*time.Millisecond:
		return green
	case latency < 300*time.Millisecond:
		return cyan
	case latency < 500*time.Millisecond:
		return blue
	case latency < time.Second:
		return yellow
	case latency < 2*time.Second:
		return magenta
	default:
		return red
	}
}

// colorForStatus 返回状态码对应的颜色
func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

// colorForMethod 返回 HTTP 方法对应的颜色
func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
