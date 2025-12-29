package common

import (
	"errors"
	"fmt"
)

// CustomError 自定义错误
type CustomError struct {
	Code int
	Err  error
}

func (ce CustomError) String() string {
	if ce.Err != nil {
		return fmt.Sprintf("code: %d, err: %s", ce.Code, ce.Err.Error())
	}
	return fmt.Sprintf("code: %d, err: nil", ce.Code)
}

// Error 实现 error 接口
func (ce CustomError) Error() string {
	if ce.Err != nil {
		return ce.Err.Error()
	}
	return fmt.Sprintf("code: %d", ce.Code)
}

// Unwrap 实现 errors.Unwrap 接口
func (ce CustomError) Unwrap() error {
	return ce.Err
}

// NewCustomError 创建业务自定义错误。
// 业务错误码必须大于等于 10001。
func NewCustomError(code int, err error) CustomError {
	if code < 10001 {
		panic("business error code must be greater than or equal to 10001")
	}
	return CustomError{code, err}
}

// NewSystemError 创建系统自定义错误。
// 系统错误码必须小于 10001。
func NewSystemError(code int, err error) CustomError {
	if code >= 10001 {
		panic("system error code must be less than 10001")
	}
	return CustomError{code, err}
}

// IsCode 判断错误链中是否存在指定错误码的 CustomError
func IsCode(err error, code int) bool {
	var customErr CustomError
	if errors.As(err, &customErr) {
		return customErr.Code == code
	}
	return false
}
