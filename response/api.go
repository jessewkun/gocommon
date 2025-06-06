package response

import (
	"errors"

	"github.com/jessewkun/gocommon/common"
	"github.com/jessewkun/gocommon/constant"

	"github.com/gin-gonic/gin"
)

// NewApiResult create a new api result
func NewApiResult(c *gin.Context, code int, message string, data interface{}) *ApiResult {
	resp := &ApiResult{
		Code:    code,
		Message: message,
		Data:    data,
		TraceId: c.GetString(constant.CTX_TRACE_ID),
	}

	// 设置返回结果，在中间件中获取记录日志
	c.Set(constant.CTX_API_OUTPUT, resp)
	return resp
}

// SuccessResp success response
func SuccessResp(c *gin.Context, data interface{}) *ApiResult {
	if data == nil {
		data = struct{}{}
	}
	return NewApiResult(c, CodeSuccess, "success", data)
}

// ErrorResp error response
func ErrorResp(c *gin.Context, err error) *ApiResult {
	if !errors.As(err, &common.CustomError{}) {
		err = newDefaultError(err)
	}
	e := err.(common.CustomError)
	return NewApiResult(c, e.Code, e.Error(), struct{}{})
}

// Custom 自定义返回
func CustomResp(ctx *gin.Context, code int, message string, data interface{}) *ApiResult {
	return NewApiResult(ctx, code, message, data)
}

// SystemErrorResp system error response
func SystemErrorResp(c *gin.Context) *ApiResult {
	return ErrorResp(c, SystemError)
}

// ParamErrorResp param error response
func ParamErrorResp(c *gin.Context) *ApiResult {
	return ErrorResp(c, ParamError)
}

// ForbiddenErrorResp forbidden error response
func ForbiddenErrorResp(c *gin.Context) *ApiResult {
	return ErrorResp(c, ForbiddenError)
}

// NotfoundErrorResp not found error response
func NotfoundErrorResp(c *gin.Context) *ApiResult {
	return ErrorResp(c, NotfoundError)
}

// RateLimiterErrorResp rate limiter error response
func RateLimiterErrorResp(c *gin.Context) *ApiResult {
	return ErrorResp(c, RateLimiterError)
}
