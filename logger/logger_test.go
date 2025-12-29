package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/constant"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupCommonPropagatedKeysForTest 保存并重置 common.PropagatedContextKeys，以便测试隔离
func setupCommonPropagatedKeysForTest() (initialKeys []constant.ContextKey, restoreFunc func()) {
	initialKeys = common.GetAllPropagatedContextKey()
	// 清空以便测试添加自己的键
	restoreFunc = func() {
		common.ClearAllPropagatedContextKey()
		for _, key := range initialKeys {
			common.RegisterPropagatedContextKey(key)
		}
	}
	return
}

func TestFieldsFromCtx(t *testing.T) {
	_, restore := setupCommonPropagatedKeysForTest()
	defer restore()

	// 注册测试所需的键
	common.RegisterPropagatedContextKey(constant.CtxTraceID)
	common.RegisterPropagatedContextKey(constant.CtxStudentID)
	common.RegisterPropagatedContextKey(constant.CtxTeacherID)
	common.RegisterPropagatedContextKey(constant.ContextKey("dynamic_string"))
	common.RegisterPropagatedContextKey(constant.ContextKey("dynamic_bool"))
	common.RegisterPropagatedContextKey(constant.ContextKey("dynamic_float"))
	common.RegisterPropagatedContextKey(constant.ContextKey("dynamic_time"))
	common.RegisterPropagatedContextKey(constant.ContextKey("dynamic_struct"))

	t.Run("should extract all propagated keys with correct types", func(t *testing.T) {
		// Arrange
		testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		sampleStruct := struct {
			Field1 string
			Field2 int
		}{"value", 123}

		ctx := context.Background()
		ctx = context.WithValue(ctx, constant.CtxTraceID, "trace-123")
		ctx = context.WithValue(ctx, constant.CtxStudentID, 202) // int type
		ctx = context.WithValue(ctx, constant.CtxTeacherID, int64(303))
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_string"), "hello")
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_bool"), true)
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_float"), 1.23)
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_time"), testTime)
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_struct"), sampleStruct)

		// Act
		fields := FieldsFromCtx(ctx)

		assert.Contains(t, fields, zap.String(string(constant.CtxTraceID), "trace-123"))
		assert.Contains(t, fields, zap.Int(string(constant.CtxStudentID), 202))
		assert.Contains(t, fields, zap.Int64(string(constant.CtxTeacherID), 303))
		assert.Contains(t, fields, zap.String("dynamic_string", "hello"))
		assert.Contains(t, fields, zap.Bool("dynamic_bool", true))
		assert.Contains(t, fields, zap.Float64("dynamic_float", 1.23))
		assert.Contains(t, fields, zap.Time("dynamic_time", testTime))

		assert.Len(t, fields, 8) // 8 registered keys with values
	})

	t.Run("should handle missing context values gracefully", func(t *testing.T) {
		// Arrange: Only TraceID and a dynamic string are present
		ctx := context.Background()
		ctx = context.WithValue(ctx, constant.CtxTraceID, "trace-456")
		ctx = context.WithValue(ctx, constant.ContextKey("dynamic_string"), "world")

		// Act
		fields := FieldsFromCtx(ctx)

		// Assert
		assert.Contains(t, fields, zap.String(string(constant.CtxTraceID), "trace-456"))
		assert.Contains(t, fields, zap.String("dynamic_string", "world"))
		assert.Len(t, fields, 2) // Only 2 values present
	})

	t.Run("should handle empty context", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		fields := FieldsFromCtx(ctx)

		// Assert
		assert.Empty(t, fields)
	})

	t.Run("should handle various integer types correctly", func(t *testing.T) {
		// Arrange
		common.RegisterPropagatedContextKey(constant.ContextKey("int_val"))
		common.RegisterPropagatedContextKey(constant.ContextKey("int8_val"))
		common.RegisterPropagatedContextKey(constant.ContextKey("uint_val"))
		common.RegisterPropagatedContextKey(constant.ContextKey("uint64_val"))

		ctx := context.Background()
		ctx = context.WithValue(ctx, constant.ContextKey("int_val"), int(1))
		ctx = context.WithValue(ctx, constant.ContextKey("int8_val"), int8(2))
		ctx = context.WithValue(ctx, constant.ContextKey("uint_val"), uint(3))
		ctx = context.WithValue(ctx, constant.ContextKey("uint64_val"), uint64(4))

		// Act
		fields := FieldsFromCtx(ctx)

		// Assert
		assert.Contains(t, fields, zap.Int("int_val", 1))
		assert.Contains(t, fields, zap.Int8("int8_val", 2))
		assert.Contains(t, fields, zap.Uint("uint_val", 3))
		assert.Contains(t, fields, zap.Uint64("uint64_val", 4))
	})

	t.Run("should be thread-safe when reading propagated keys", func(t *testing.T) {
		var wg sync.WaitGroup
		numReaders := 10
		numWriters := 2

		// Add some initial keys
		common.RegisterPropagatedContextKey(constant.ContextKey("init_key1"))
		common.RegisterPropagatedContextKey(constant.ContextKey("init_key2"))

		// Start multiple readers
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = FieldsFromCtx(context.Background())

			}()
		}

		// Start some writers (Register new keys concurrently)
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				common.RegisterPropagatedContextKey(constant.ContextKey(fmt.Sprintf("con_key_%d", idx)))
			}(i)
		}
		wg.Wait()
		// No panics or deadlocks mean thread-safety is likely okay
	})
}
