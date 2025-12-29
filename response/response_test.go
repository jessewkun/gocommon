package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/constant"
	"github.com/stretchr/testify/assert"
)

// setupGinTestContext 创建一个用于测试的 gin.Context 和 httptest.ResponseRecorder
func setupGinTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}
	// 模拟 TraceID
	ctx.Set(string(constant.CtxTraceID), "test-trace-id")
	return ctx, w
}

func TestAPIResult_String(t *testing.T) {
	result := &APIResult{
		Code:    0,
		Message: "success",
		Data:    "test data",
		TraceID: "abc",
	}
	expectedJSON := `{"code":0,"message":"success","data":"test data","trace_id":"abc"}`
	assert.JSONEq(t, expectedJSON, result.String())
}

func TestNewAPIResult(t *testing.T) {
	ctx, _ := setupGinTestContext()
	result := NewAPIResult(ctx, 1, "test msg", "data")

	assert.Equal(t, 1, result.Code)
	assert.Equal(t, "test msg", result.Message)
	assert.Equal(t, "data", result.Data)
	assert.Equal(t, "test-trace-id", result.TraceID)
	// 验证 ctx.Set 是否被调用
	_, exists := ctx.Get(string(constant.CtxAPIOutput))
	assert.True(t, exists, "API output should be set in context")
	assert.Equal(t, result, ctx.MustGet(string(constant.CtxAPIOutput)).(*APIResult))
}

func TestNewAPIResultWs(t *testing.T) {
	ctx, _ := setupGinTestContext()
	result := NewAPIResultWs(ctx, 1, "ws msg", "ws data")

	assert.Equal(t, 1, result.Code)
	assert.Equal(t, "ws msg", result.Message)
	assert.Equal(t, "ws data", result.Data)
	assert.Equal(t, "test-trace-id", result.TraceID)
	// 验证 ctx.Set 是否未被调用
	_, exists := ctx.Get(string(constant.CtxAPIOutput))
	assert.False(t, exists)
}

func TestSuccess(t *testing.T) {
	t.Run("with data", func(t *testing.T) {
		ctx, w := setupGinTestContext()
		data := gin.H{"item": "value"}
		Success(ctx, data)

		assert.Equal(t, http.StatusOK, w.Code)
		var result APIResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, CodeSuccess, result.Code)
		assert.Equal(t, "success", result.Message)
		assert.Equal(t, map[string]interface{}(data), result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})

	t.Run("with nil data", func(t *testing.T) {
		ctx, w := setupGinTestContext()
		Success(ctx, nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var result APIResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, CodeSuccess, result.Code)
		assert.Equal(t, "success", result.Message)
		assert.NotNil(t, result.Data) // Should be an empty struct{}
	})
}

func TestSuccessWs(t *testing.T) {
	t.Run("with data", func(t *testing.T) {
		ctx, _ := setupGinTestContext()
		data := gin.H{"ws_item": "ws_value"}
		bytes, err := SuccessWs(ctx, data)

		assert.NoError(t, err)
		var result APIResult
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, CodeSuccess, result.Code)
		assert.Equal(t, "success", result.Message)
		assert.Equal(t, map[string]interface{}(data), result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})

	t.Run("with nil data", func(t *testing.T) {
		ctx, _ := setupGinTestContext()
		bytes, err := SuccessWs(ctx, nil)

		assert.NoError(t, err)
		var result APIResult
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, CodeSuccess, result.Code)
		assert.Equal(t, "success", result.Message)
		assert.NotNil(t, result.Data) // Should be an empty struct{}
	})
}

func TestError(t *testing.T) {
	t.Run("with common.NewCustomError", func(t *testing.T) {
		ctx, w := setupGinTestContext()
		bizErr := common.NewCustomError(10001, errors.New("business error"))
		Error(ctx, bizErr)

		assert.Equal(t, http.StatusOK, w.Code)
		var result APIResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 10001, result.Code)
		assert.Equal(t, "business error", result.Message)
		assert.NotNil(t, result.Data) // Should be empty struct{}
		assert.Equal(t, "test-trace-id", result.TraceID)
	})

	t.Run("with common.NewSystemError", func(t *testing.T) {
		ctx, w := setupGinTestContext()
		sysErr := common.NewSystemError(1000, errors.New("system error"))
		Error(ctx, sysErr)

		assert.Equal(t, http.StatusOK, w.Code)
		var result APIResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1000, result.Code)
		assert.Equal(t, "system error", result.Message)
		assert.NotNil(t, result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})

	t.Run("with plain error", func(t *testing.T) {
		ctx, w := setupGinTestContext()
		plainErr := errors.New("something unexpected happened")
		Error(ctx, plainErr)

		assert.Equal(t, http.StatusOK, w.Code)
		var result APIResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, DefaultErrorCode, result.Code) // Should map to DefaultErrorCode
		assert.Equal(t, plainErr.Error(), result.Message)
		assert.NotNil(t, result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})
}

func TestErrorWs(t *testing.T) {
	t.Run("with common.NewCustomError", func(t *testing.T) {
		ctx, _ := setupGinTestContext()
		bizErr := common.NewCustomError(10001, errors.New("ws business error"))
		bytes, err := ErrorWs(ctx, bizErr)

		assert.NoError(t, err)
		var result APIResult
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, 10001, result.Code)
		assert.Equal(t, "ws business error", result.Message)
		assert.NotNil(t, result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})

	t.Run("with plain error", func(t *testing.T) {
		ctx, _ := setupGinTestContext()
		plainErr := errors.New("ws unexpected happened")
		bytes, err := ErrorWs(ctx, plainErr)

		assert.NoError(t, err)
		var result APIResult
		err = json.Unmarshal(bytes, &result)
		assert.NoError(t, err)
		assert.Equal(t, DefaultErrorCode, result.Code)
		assert.Equal(t, plainErr.Error(), result.Message)
		assert.NotNil(t, result.Data)
		assert.Equal(t, "test-trace-id", result.TraceID)
	})
}

func TestCustom(t *testing.T) {
	ctx, w := setupGinTestContext()
	code := 200
	message := "custom message"
	data := gin.H{"custom": "data"}
	Custom(ctx, code, message, data)

	assert.Equal(t, http.StatusOK, w.Code)
	var result APIResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, code, result.Code)
	assert.Equal(t, message, result.Message)
	assert.Equal(t, map[string]interface{}(data), result.Data)
	assert.Equal(t, "test-trace-id", result.TraceID)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name         string
		handlerFunc  func(*gin.Context)
		expectedCode int
		expectedMsg  string
	}{
		{"SystemError", SystemError, SystemErrors.Code, SystemErrors.Error()},
		{"ParamError", ParamError, ParamErrors.Code, ParamErrors.Error()},
		{"ForbiddenError", ForbiddenError, ForbiddenErrors.Code, ForbiddenErrors.Error()},
		{"NotfoundError", NotfoundError, NotfoundErrors.Code, NotfoundErrors.Error()},
		{"RateLimiterError", RateLimiterError, RateLimiterErrors.Code, RateLimiterErrors.Error()},
		{"UnknownError", UnknownError, UnknownErrors.Code, UnknownErrors.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, w := setupGinTestContext()
			tt.handlerFunc(ctx)

			assert.Equal(t, http.StatusOK, w.Code)
			var result APIResult
			err := json.Unmarshal(w.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.NotNil(t, result.Data) // Should be empty struct{}
		})
	}
}
