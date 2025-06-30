package response

import (
	"errors"

	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/constant"

	"github.com/gin-gonic/gin"
)

// NewAPIResult create a new api result
func NewAPIResult(c *gin.Context, code int, message string, data interface{}) *APIResult {
	resp := &APIResult{
		Code:    code,
		Message: message,
		Data:    data,
		TraceID: c.GetString(string(constant.CtxTraceID)),
	}

	// 设置返回结果，在中间件中获取记录日志
	c.Set(string(constant.CtxAPIOutput), resp)
	return resp
}

// SuccessResp success response
func SuccessResp(c *gin.Context, data interface{}) *APIResult {
	if data == nil {
		data = struct{}{}
	}
	return NewAPIResult(c, CodeSuccess, "success", data)
}

// ErrorResp error response
func ErrorResp(c *gin.Context, err error) *APIResult {
	if !errors.As(err, &common.CustomError{}) {
		err = newDefaultError(err)
	}
	e := err.(common.CustomError)
	return NewAPIResult(c, e.Code, e.Error(), struct{}{})
}

// CustomResp 自定义返回
func CustomResp(ctx *gin.Context, code int, message string, data interface{}) *APIResult {
	return NewAPIResult(ctx, code, message, data)
}

// SystemErrorResp system error response
func SystemErrorResp(c *gin.Context) *APIResult {
	return ErrorResp(c, SystemError)
}

// ParamErrorResp param error response
func ParamErrorResp(c *gin.Context) *APIResult {
	return ErrorResp(c, ParamError)
}

// ForbiddenErrorResp forbidden error response
func ForbiddenErrorResp(c *gin.Context) *APIResult {
	return ErrorResp(c, ForbiddenError)
}

// NotfoundErrorResp not found error response
func NotfoundErrorResp(c *gin.Context) *APIResult {
	return ErrorResp(c, NotfoundError)
}

// RateLimiterErrorResp rate limiter error response
func RateLimiterErrorResp(c *gin.Context) *APIResult {
	return ErrorResp(c, RateLimiterError)
}
