// Package common 提供通用的工具函数和类型
package common

import (
	"context"
	"sync"

	"github.com/jessewkun/gocommon/constant"
)

var (
	// propagatedContextKeys 存储需要被 CopyCtx 传播的上下文键
	propagatedContextKeys = []constant.ContextKey{
		constant.CtxTraceID,
		// constant.CtxUserID,
		constant.CtxStudentID,
		constant.CtxTeacherID,
	}
	// propagatedKeysMutex 用于保护对 propagatedContextKeys 的并发访问
	propagatedKeysMutex sync.RWMutex
)

// RegisterPropagatedContextKey 注册一个需要被 CopyCtx 传播的上下文键。
func RegisterPropagatedContextKey(key constant.ContextKey) {
	propagatedKeysMutex.Lock()
	defer propagatedKeysMutex.Unlock()

	// 避免重复注册
	for _, k := range propagatedContextKeys {
		if k == key {
			return
		}
	}
	propagatedContextKeys = append(propagatedContextKeys, key)
}

// GetAllPropagatedContextKey 获取所有需要被 CopyCtx 传播的上下文键
func GetAllPropagatedContextKey() []constant.ContextKey {
	propagatedKeysMutex.RLock()
	defer propagatedKeysMutex.RUnlock()
	keys := make([]constant.ContextKey, len(propagatedContextKeys))
	copy(keys, propagatedContextKeys)
	return keys
}

// ClearAllPropagatedContextKey 清除所有需要被 CopyCtx 传播的上下文键
func ClearAllPropagatedContextKey() {
	propagatedKeysMutex.Lock()
	defer propagatedKeysMutex.Unlock()
	propagatedContextKeys = make([]constant.ContextKey, 0)
}

// ClearPropagatedContextKey 清除指定的需要被 CopyCtx 传播的上下文键
func ClearPropagatedContextKey(key constant.ContextKey) {
	propagatedKeysMutex.Lock()
	defer propagatedKeysMutex.Unlock()

	index := -1
	for i, k := range propagatedContextKeys {
		if k == key {
			index = i
			break
		}
	}

	if index >= 0 {
		propagatedContextKeys = append(propagatedContextKeys[:index], propagatedContextKeys[index+1:]...)
	}
}

// CopyCtx 复制新的 context
//
// 避免在gin框架中，http请求结束后，context被cancel，导致在请求中新开的 goroutine 中使用context时出现 ctx canceled 错误
func CopyCtx(ctx context.Context) context.Context {
	newCtx := context.Background()

	propagatedKeysMutex.RLock()
	defer propagatedKeysMutex.RUnlock()

	for _, key := range propagatedContextKeys {
		if v := ctx.Value(key); v != nil {
			newCtx = context.WithValue(newCtx, key, v)
		}
	}
	return newCtx
}
