// Package safego 提供安全的 goroutine 管理工具，防止 panic 导致服务崩溃
package safego

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/jessewkun/gocommon/logger"
)

// WaitGroupWrapper 封装 sync.WaitGroup，提供安全的 goroutine 等待机制
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// Wrap 启动一个新的 goroutine 并执行传入的函数，自动进行 panic 保护
// 参数：
//   - c: 上下文，用于链路追踪和日志记录
//   - f: 要执行的函数，会在新的 goroutine 中安全执行
//
// 使用示例：
//
//	var wg WaitGroupWrapper
//	wg.Wrap(ctx, func() {
//	    // 这里的代码如果发生 panic，会被安全捕获
//	    riskyOperation()
//	})
//	wg.Wait() // 等待所有 goroutine 完成
func (wg *WaitGroupWrapper) Wrap(c context.Context, f func()) {
	wg.Add(1)
	go func(c context.Context) {
		defer wg.Done()
		SafeGo(c, f)
	}(c)
}

// SafeGo 对函数进行 panic 保护，捕获并记录 panic 信息
// 注意：SafeGo 本身不会启动新的 goroutine，需要在已有的 goroutine 中调用
//
// 参数：
//   - c: 上下文，用于日志记录和链路追踪
//   - fun: 要执行的函数，会被 panic 保护
//
// 功能：
//   - 使用 defer recover() 捕获 panic
//   - 自动记录 panic 错误信息和完整堆栈跟踪
//   - 支持上下文传递，便于日志记录
//
// 使用示例：
//
//	// 在协程中使用（推荐）
//	go func() {
//	    SafeGo(ctx, func() {
//	        // 这里的代码如果发生 panic，会被安全捕获
//	        riskyOperation()
//	    })
//	}()
//
//	// 在主协程中使用（也可以）
//	SafeGo(ctx, func() {
//	    // 主协程中的危险操作
//	    initializeSystem()
//	})
func SafeGo(c context.Context, fun func()) {
	defer func(ctx context.Context) {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			logger.ErrorWithMsg(c, "safego", "safeGo recover: %+v, stack is %s", err, stack)
		}
	}(c)
	fun()
}
