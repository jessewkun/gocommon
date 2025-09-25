package common

import "fmt"

// CustomError 自定义错误
type CustomError struct {
	Code int
	Err  error
}

func (ce CustomError) String() string {
	return fmt.Sprintf("code: %d, err: %s", ce.Code, ce.Err.Error())
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

// Is 实现 errors.Is 接口
func (ce CustomError) Is(target error) bool {
	if targetCustomErr, ok := target.(CustomError); ok {
		return ce.Code == targetCustomErr.Code
	}
	return false
}

// NewCustomError 创建自定义错误
// 业务自定义错误码必须大于10000，小于10000的错误码为系统错误码，10000为默认业务错误码
func NewCustomError(code int, err error) CustomError {
	if code < 10001 {
		panic("error code must greater than 10000")
	}
	return CustomError{code, err}
}
