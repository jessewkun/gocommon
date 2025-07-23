package response

import (
	"encoding/json"
	"errors"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/constant"
)

func newTestContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func TestNewAPIResult(t *testing.T) {
	c := newTestContext()
	c.Set(string(constant.CtxTraceID), "trace123")
	resp := NewAPIResult(c, 0, "ok", map[string]interface{}{"a": 1})
	if resp.Code != 0 || resp.Message != "ok" || resp.TraceID != "trace123" {
		t.Errorf("NewAPIResult 返回值不正确: %+v", resp)
	}
}

func TestSuccessResp(t *testing.T) {
	c := newTestContext()
	resp := SuccessResp(c, map[string]interface{}{"a": 1})
	if resp.Code != CodeSuccess || resp.Message != "success" {
		t.Errorf("SuccessResp 返回值不正确: %+v", resp)
	}
}

func TestErrorResp(t *testing.T) {
	c := newTestContext()
	err := common.NewCustomError(10001, errors.New("test error"))
	resp := ErrorResp(c, err)
	if resp.Code != 10001 || resp.Message != "test error" {
		t.Errorf("ErrorResp 返回值不正确: %+v", resp)
	}
	// 非 CustomError
	resp2 := ErrorResp(c, errors.New("other error"))
	if resp2.Code != DefaultErrorCode {
		t.Errorf("ErrorResp 非自定义错误未返回默认错误码: %+v", resp2)
	}
}

func TestCustomResp(t *testing.T) {
	c := newTestContext()
	resp := CustomResp(c, 123, "msg", 456)
	if resp.Code != 123 || resp.Message != "msg" || resp.Data != 456 {
		t.Errorf("CustomResp 返回值不正确: %+v", resp)
	}
}

func TestSystemErrorResp(t *testing.T) {
	c := newTestContext()
	resp := SystemErrorResp(c)
	if resp.Code != SystemError.Code {
		t.Errorf("SystemErrorResp 返回值不正确: %+v", resp)
	}
}

func TestParamErrorResp(t *testing.T) {
	c := newTestContext()
	resp := ParamErrorResp(c)
	if resp.Code != ParamError.Code {
		t.Errorf("ParamErrorResp 返回值不正确: %+v", resp)
	}
}

func TestForbiddenErrorResp(t *testing.T) {
	c := newTestContext()
	resp := ForbiddenErrorResp(c)
	if resp.Code != ForbiddenError.Code {
		t.Errorf("ForbiddenErrorResp 返回值不正确: %+v", resp)
	}
}

func TestNotfoundErrorResp(t *testing.T) {
	c := newTestContext()
	resp := NotfoundErrorResp(c)
	if resp.Code != NotfoundError.Code {
		t.Errorf("NotfoundErrorResp 返回值不正确: %+v", resp)
	}
}

func TestRateLimiterErrorResp(t *testing.T) {
	c := newTestContext()
	resp := RateLimiterErrorResp(c)
	if resp.Code != RateLimiterError.Code {
		t.Errorf("RateLimiterErrorResp 返回值不正确: %+v", resp)
	}
}

func TestApiResult_String(t *testing.T) {
	r := &APIResult{Code: 1, Message: "msg", Data: 2, TraceID: "tid"}
	str := r.String()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(str), &m); err != nil {
		t.Errorf("APIResult.String() 不是合法json: %s", str)
	}
}

func TestBindAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := newTestContext()
	c.Request, _ = http.NewRequest("GET", "/?A=1", nil)
	type Req struct {
		A int `form:"A" binding:"required"`
	}
	var req Req
	err := BindAndValidate(c, &req)
	if err != nil {
		t.Errorf("BindAndValidate GET 失败: %+v", err)
	}
	if req.A != 1 {
		t.Errorf("BindAndValidate GET 失败: %+v", req)
	}
}
