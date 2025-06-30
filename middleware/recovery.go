package middleware

import (
	"bytes"
	"fmt"
	"net/http"
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
				c.JSON(http.StatusOK, response.SystemErrorResp(c))
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
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
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
