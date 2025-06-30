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

// SystemError 系统错误
var SystemError = common.CustomError{Code: 1000, Err: errors.New("系统错误，请稍后重试")}

// ParamError 参数错误
var ParamError = common.CustomError{Code: 1001, Err: errors.New("参数错误")}

// ForbiddenError 权限错误
var ForbiddenError = common.CustomError{Code: 1002, Err: errors.New("permission denied")}

// NotfoundError 未找到
var NotfoundError = common.CustomError{Code: 1003, Err: errors.New("not found")}

// RateLimiterError 限流
var RateLimiterError = common.CustomError{Code: 1100, Err: errors.New("too many requests")}

// newDefaultError 创建默认错误
func newDefaultError(err error) common.CustomError {
	return common.CustomError{Code: DefaultErrorCode, Err: err}
}
