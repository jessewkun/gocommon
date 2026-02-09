package middleware

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/logger"
	"github.com/jessewkun/gocommon/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				trace := PanicTrace(2)
				logger.ErrorWithField(c.Request.Context(), "RECOVERY", "PANIC", map[string]interface{}{
					"recover": r,
					"panic":   string(trace),
				})
				if common.IsDebug() {
					fmt.Printf("recover: %+v\n", r)
					fmt.Printf("panic: %+v\n", string(trace))
				}
				response.SystemError(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //KB
	length := runtime.Stack(stack, true)
	stack = stack[:length]
	start := bytes.Index(stack, s)
	if start >= 0 {
		stack = stack[start:]
	}
	start = bytes.Index(stack, line)
	if start >= 0 && start+1 < len(stack) {
		stack = stack[start+1:]
	}
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}
