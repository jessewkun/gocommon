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
	if v, ok := ctx.Value(constant.CTX_USER_ID).(string); ok && v != "" {
		fields = append(fields, zap.String(constant.CTX_USER_ID, v))
	}
	return fields
}
