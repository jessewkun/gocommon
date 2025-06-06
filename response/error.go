package response

import (
	"errors"

	"github.com/jessewkun/gocommon/common"
)

// 默认业务错误码
const (
	CodeSuccess        = 0
	DEFAULT_ERROR_CODE = 10000
)

// 系统错误
var SystemError = common.CustomError{1000, errors.New("系统错误，请稍后重试")}

// 参数错误
var ParamError = common.CustomError{1001, errors.New("参数错误")}

// 权限错误
var ForbiddenError = common.CustomError{1002, errors.New("Permission denied")}

// 未找到
var NotfoundError = common.CustomError{1003, errors.New("Not found")}

// 限流
var RateLimiterError = common.CustomError{1100, errors.New("Too many requests")}

// newDefaultError 创建默认错误
func newDefaultError(err error) common.CustomError {
	return common.NewCustomError(DEFAULT_ERROR_CODE, err)
}
