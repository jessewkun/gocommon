// Package debug 提供调试日志输出功能
package debug

import (
	"context"
	"fmt"
	"time"

	"github.com/jessewkun/gocommon/logger"
)

// Log 如果指定模块的调试开关已开启，则输出调试信息。
// 这是推荐使用的唯一调试函数。
func Log(ctx context.Context, module string, format string, v ...interface{}) {
	if defaultDebugger.isDebug(module) {
		// 将 module 作为 tag 传递
		printMessage(ctx, module, format, v...)
	}
}

// IsDebug 检查指定模块的调试是否开启。
func IsDebug(module string) bool {
	return defaultDebugger.isDebug(module)
}

// printMessage 内部函数，负责实际的打印逻辑
func printMessage(ctx context.Context, module string, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	defaultDebugger.mu.RLock()
	mode := defaultDebugger.Config.Mode
	defaultDebugger.mu.RUnlock()

	tag := "DEBUG_" + module
	if mode == "log" {
		logger.Debug(ctx, tag, "%s", msg) // 使用 module 作为 tag
	} else {
		fmt.Printf("[DEBUG][%s][%s] %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			tag, // 使用 module 作为 tag
			msg)
	}
}
