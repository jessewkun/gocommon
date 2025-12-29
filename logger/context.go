// Package logger 提供日志记录功能
package logger

import (
	"context"
	"time"

	"github.com/jessewkun/gocommon/common"
	"go.uber.org/zap"
)

func FieldsFromCtx(ctx context.Context) []zap.Field {
	var fields []zap.Field

	allKeys := common.GetAllPropagatedContextKey()

	for _, key := range allKeys {
		if value := ctx.Value(key); value != nil {
			switch v := value.(type) {
			case string:
				fields = append(fields, zap.String(string(key), v))
			case bool:
				fields = append(fields, zap.Bool(string(key), v))
			case int:
				fields = append(fields, zap.Int(string(key), v))
			case int8:
				fields = append(fields, zap.Int8(string(key), v))
			case int16:
				fields = append(fields, zap.Int16(string(key), v))
			case int32:
				fields = append(fields, zap.Int32(string(key), v))
			case int64:
				fields = append(fields, zap.Int64(string(key), v))
			case uint:
				fields = append(fields, zap.Uint(string(key), v))
			case uint8:
				fields = append(fields, zap.Uint8(string(key), v))
			case uint16:
				fields = append(fields, zap.Uint16(string(key), v))
			case uint32:
				fields = append(fields, zap.Uint32(string(key), v))
			case uint64:
				fields = append(fields, zap.Uint64(string(key), v))
			case float32:
				fields = append(fields, zap.Float32(string(key), v))
			case float64:
				fields = append(fields, zap.Float64(string(key), v))
			case time.Time:
				fields = append(fields, zap.Time(string(key), v))
			default:
				fields = append(fields, zap.Any(string(key), v))
			}
		}
	}
	return fields
}
