package common

import (
	"context"

	"github.com/jessewkun/gocommon/constant"
)

// CopyCtx 复制新的 context
//
// 避免在gin框架中，http请求结束后，context被cancel，导致在请求中新开的 goroutine 中使用context时出现 ctx canceled 错误
func CopyCtx(ctx context.Context) context.Context {
	newCtx := context.Background()
	if v := ctx.Value(constant.CTX_USER_ID); v != nil {
		newCtx = context.WithValue(newCtx, constant.CTX_USER_ID, v)
	}
	if v := ctx.Value(constant.CTX_TRACE_ID); v != nil {
		newCtx = context.WithValue(newCtx, constant.CTX_TRACE_ID, v)
	}
	return newCtx
}
