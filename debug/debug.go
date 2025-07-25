package debug

import (
	"context"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/logger"
)

const TAG = "DEBUG"

// InitDebug 初始化debug
func InitDebug(flag string) DebugFunc {
	return func(c context.Context, format string, v ...interface{}) {
		if IsDebug(flag) {
			hookPrint(c, format, v...)
		}
	}
}

// IsDebug 是否开启debug
func IsDebug(flag string) bool {
	enable := false
	for _, part := range Cfg.Module {
		if part == flag {
			enable = true
			break
		}
	}
	return enable
}

// hookPrint 输出debug信息
func hookPrint(c context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if Cfg.Mode == "log" {
		logger.Debug(c, TAG, msg)
	} else {
		fmt.Printf("[DEBUG][%s][%s] %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			TAG,
			msg)
	}
}
