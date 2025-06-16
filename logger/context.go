package logger

import (
	"context"

	"github.com/jessewkun/gocommon/constant"
	"go.uber.org/zap"
)

func FieldsFromCtx(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if v, ok := ctx.Value(constant.CTX_TRACE_ID).(string); ok && v != "" {
		fields = append(fields, zap.String(constant.CTX_TRACE_ID, v))
	}
	if v := ctx.Value(constant.CTX_USER_ID); v != nil {
		switch v := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			fields = append(fields, zap.Int(constant.CTX_USER_ID, v.(int)))
		case string:
			fields = append(fields, zap.String(constant.CTX_USER_ID, v))
		default:
			fields = append(fields, zap.Any(constant.CTX_USER_ID, v))
		}
	}
	return fields
}
