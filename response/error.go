package response

import (
	"errors"

	"github.com/jessewkun/gocommon/common"
)

// 预定义错误码
const (
	CodeSuccess = 0
	// DefaultErrorCode 默认系统错误码
	// 当一个未知错误被包装时使用
	DefaultErrorCode = 10000
)

// SystemErrors 系统错误
var SystemErrors = common.NewSystemError(1000, errors.New("系统错误，请稍后重试"))

// ParamErrors 参数错误
var ParamErrors = common.NewSystemError(1001, errors.New("参数错误"))

// ForbiddenErrors 权限错误
var ForbiddenErrors = common.NewSystemError(1002, errors.New("permission denied"))

// NotfoundErrors 未找到
var NotfoundErrors = common.NewSystemError(1003, errors.New("not found"))

// RateLimiterErrors 限流
var RateLimiterErrors = common.NewSystemError(1004, errors.New("too many requests"))

// UnknownErrors 未知错误
var UnknownErrors = common.NewSystemError(1005, errors.New("未知错误"))

// newDefaultError 创建默认错误
func newDefaultError(err error) common.CustomError {
	return common.NewSystemError(DefaultErrorCode, err)
}
