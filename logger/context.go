// Package logger 提供日志记录功能
package logger

import (
	"context"

	"github.com/jessewkun/gocommon/constant"
	"go.uber.org/zap"
)

func FieldsFromCtx(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if v, ok := ctx.Value(constant.CtxTraceID).(string); ok && v != "" {
		fields = append(fields, zap.String(string(constant.CtxTraceID), v))
	}
	if v := ctx.Value(constant.CtxUserID); v != nil {
		switch v := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			fields = append(fields, zap.Int(string(constant.CtxUserID), v.(int)))
		case string:
			fields = append(fields, zap.String(string(constant.CtxUserID), v))
		default:
			fields = append(fields, zap.Any(string(constant.CtxUserID), v))
		}
	}
	return fields
}
