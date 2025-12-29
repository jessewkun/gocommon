package common

import (
	"context"
	"sync"
	"testing"

	"github.com/jessewkun/gocommon/constant"
	"github.com/stretchr/testify/assert"
)

// setupTest 用于保存并重置 PropagatedContextKeys，以便每个测试独立运行
func setupTest() (initialKeys []constant.ContextKey, restoreFunc func()) {
	initialKeys = GetAllPropagatedContextKey()      // Get a copy of initial keys
	propagatedContextKeys = []constant.ContextKey{} // Use a new empty slice to avoid modifying previous test state

	restoreFunc = func() {
		propagatedKeysMutex.Lock()
		propagatedContextKeys = initialKeys // Restore the initial state
		propagatedKeysMutex.Unlock()
	}
	return
}

func TestRegisterPropagatedContextKey(t *testing.T) {
	_, restoreFunc := setupTest()
	defer restoreFunc()

	key1 := constant.ContextKey("testKey1")

	t.Run("should register a new key", func(t *testing.T) {
		RegisterPropagatedContextKey(key1)
		assert.Contains(t, GetAllPropagatedContextKey(), key1)
	})

	t.Run("should not register a duplicate key", func(t *testing.T) {
		initialLen := len(propagatedContextKeys)
		RegisterPropagatedContextKey(key1) // Register again
		assert.Len(t, GetAllPropagatedContextKey(), initialLen, "Duplicate key should not increase slice length")
	})

	t.Run("should handle concurrent registrations safely", func(t *testing.T) {
		var wg sync.WaitGroup
		newKeys := []constant.ContextKey{"conKey1", "conKey2", "conKey3"}
		initialLen := len(propagatedContextKeys)

		for _, key := range newKeys {
			wg.Add(1)
			go func(k constant.ContextKey) {
				defer wg.Done()
				RegisterPropagatedContextKey(k)
			}(key)
		}
		wg.Wait()

		assert.Len(t, GetAllPropagatedContextKey(), initialLen+len(newKeys), "All new keys should be registered once")
		for _, key := range newKeys {
			assert.Contains(t, GetAllPropagatedContextKey(), key)
		}
	})
}

func TestCopyCtx(t *testing.T) {
	_, restoreFunc := setupTest()
	defer restoreFunc()

	t.Run("should copy initial relevant values", func(t *testing.T) {
		// Arrange
		// Explicitly register keys needed for this test scenario
		RegisterPropagatedContextKey(constant.CtxTraceID)
		RegisterPropagatedContextKey(constant.CtxStudentID)
		RegisterPropagatedContextKey(constant.CtxTeacherID)

		traceID := "trace-123"
		studentID := uint64(200)
		teacherID := uint64(300)

		sourceCtx := context.Background()
		sourceCtx = context.WithValue(sourceCtx, constant.CtxTraceID, traceID)
		sourceCtx = context.WithValue(sourceCtx, constant.CtxStudentID, studentID)
		sourceCtx = context.WithValue(sourceCtx, constant.CtxTeacherID, teacherID)

		// Act
		newCtx := CopyCtx(sourceCtx)

		// Assert
		assert.Equal(t, traceID, newCtx.Value(constant.CtxTraceID))
		assert.Equal(t, studentID, newCtx.Value(constant.CtxStudentID))
		assert.Equal(t, teacherID, newCtx.Value(constant.CtxTeacherID))
	})

	t.Run("should copy dynamically registered values", func(t *testing.T) {
		// Arrange
		newKey := constant.ContextKey("dynamic_key")
		newVal := "dynamic_value"
		RegisterPropagatedContextKey(newKey)              // Dynamically register a new key
		RegisterPropagatedContextKey(constant.CtxTraceID) // Ensure CtxTraceID is registered

		sourceCtx := context.Background()
		sourceCtx = context.WithValue(sourceCtx, newKey, newVal)
		sourceCtx = context.WithValue(sourceCtx, constant.CtxTraceID, "existing_trace")

		// Act
		newCtx := CopyCtx(sourceCtx)

		// Assert
		assert.Equal(t, newVal, newCtx.Value(newKey))
		assert.Equal(t, "existing_trace", newCtx.Value(constant.CtxTraceID)) // Ensure existing keys still work
	})

	t.Run("should handle missing values gracefully", func(t *testing.T) {
		// Arrange
		RegisterPropagatedContextKey(constant.CtxTraceID)   // Ensure CtxTraceID is registered
		RegisterPropagatedContextKey(constant.CtxStudentID) // Ensure CtxStudentID is registered

		traceID := "trace-456"
		sourceCtx := context.Background()
		sourceCtx = context.WithValue(sourceCtx, constant.CtxTraceID, traceID)

		// Act
		newCtx := CopyCtx(sourceCtx)

		// Assert
		assert.Equal(t, traceID, newCtx.Value(constant.CtxTraceID))
		assert.Nil(t, newCtx.Value(constant.CtxStudentID), "CtxStudentID should be nil if not in source context")
	})

	t.Run("should return a valid context from an empty source context", func(t *testing.T) {
		// Arrange
		sourceCtx := context.Background()

		// Act
		newCtx := CopyCtx(sourceCtx)

		// Assert
		assert.NotNil(t, newCtx)
		assert.Nil(t, newCtx.Value(constant.CtxTraceID))
		assert.Nil(t, newCtx.Value(constant.CtxStudentID))
	})
}
