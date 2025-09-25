package response

import (
	"errors"

	"github.com/jessewkun/gocommon/common"
)

// 默认业务错误码
const (
	CodeSuccess      = 0
	DefaultErrorCode = 10000
)

// SystemErrors 系统错误
var SystemErrors = common.CustomError{Code: 1000, Err: errors.New("系统错误，请稍后重试")}

// ParamErrors 参数错误
var ParamErrors = common.CustomError{Code: 1001, Err: errors.New("参数错误")}

// ForbiddenErrors 权限错误
var ForbiddenErrors = common.CustomError{Code: 1002, Err: errors.New("permission denied")}

// NotfoundErrors 未找到
var NotfoundErrors = common.CustomError{Code: 1003, Err: errors.New("not found")}

// RateLimiterErrors 限流
var RateLimiterErrors = common.CustomError{Code: 1004, Err: errors.New("too many requests")}

// UnknownErrors 未知错误
var UnknownErrors = common.CustomError{Code: 1005, Err: errors.New("未知错误")}

// newDefaultError 创建默认错误
func newDefaultError(err error) common.CustomError {
	return common.CustomError{Code: DefaultErrorCode, Err: err}
}
