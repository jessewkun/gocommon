// Package response 提供API返回结果的封装
package response

import (
	"encoding/json"
	"errors"
	"net/http"

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

// NewAPIResultWs 创建一个不设置返回结果的APIResult
// 目前的方式会导致 websocket 请求在连接中断的时候才记录日志，并且日志量巨大，所以就不记录返回值了
// 注意 iolog 还是会记录，只是没有 response
func NewAPIResultWs(c *gin.Context, code int, message string, data interface{}) *APIResult {
	resp := &APIResult{
		Code:    code,
		Message: message,
		Data:    data,
		TraceID: c.GetString(string(constant.CtxTraceID)),
	}

	return resp
}

// Success success response
func Success(c *gin.Context, data interface{}) {
	if data == nil {
		data = struct{}{}
	}
	c.JSON(http.StatusOK, NewAPIResult(c, CodeSuccess, "success", data))
}

// SuccessWs success response for websocket
func SuccessWs(c *gin.Context, data interface{}) ([]byte, error) {
	if data == nil {
		data = struct{}{}
	}
	return json.Marshal(NewAPIResultWs(c, CodeSuccess, "success", data))
}

// Error error response
func Error(c *gin.Context, err error) {
	var customErr common.CustomError
	if !errors.As(err, &customErr) {
		// 如果不是自定义错误，则包装为默认的系统错误
		customErr = newDefaultError(err)
	}
	c.JSON(http.StatusOK, NewAPIResult(c, customErr.Code, customErr.Error(), struct{}{}))
}

// ErrorWs error response for websocket
func ErrorWs(c *gin.Context, err error) ([]byte, error) {
	var customErr common.CustomError
	if !errors.As(err, &customErr) {
		// 如果不是自定义错误，则包装为默认的系统错误
		customErr = newDefaultError(err)
	}
	return json.Marshal(NewAPIResultWs(c, customErr.Code, customErr.Error(), struct{}{}))
}

// Custom 自定义返回
func Custom(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, NewAPIResult(c, code, message, data))
}

// SystemError system error response
func SystemError(c *gin.Context) {
	Error(c, SystemErrors)
}

// ParamError param error response
func ParamError(c *gin.Context) {
	Error(c, ParamErrors)
}

// ForbiddenError forbidden error response
func ForbiddenError(c *gin.Context) {
	Error(c, ForbiddenErrors)
}

// NotfoundError not found error response
func NotfoundError(c *gin.Context) {
	Error(c, NotfoundErrors)
}

// RateLimiterError rate limiter error response
func RateLimiterError(c *gin.Context) {
	Error(c, RateLimiterErrors)
}

// UnknownError unknown error response
func UnknownError(c *gin.Context) {
	Error(c, UnknownErrors)
}
