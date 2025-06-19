package common

import (
	"errors"
	"fmt"
	"testing"
)

func TestCustomError_String(t *testing.T) {
	err := errors.New("test error")
	ce := CustomError{Code: 10001, Err: err}
	expected := fmt.Sprintf("code: %d, err: %s", 10001, err.Error())
	if ce.String() != expected {
		t.Errorf("String() = %q, want %q", ce.String(), expected)
	}
}

func TestCustomError_Error(t *testing.T) {
	err := errors.New("test error")
	ce := CustomError{Code: 10001, Err: err}
	if ce.Error() != err.Error() {
		t.Errorf("Error() = %q, want %q", ce.Error(), err.Error())
	}

	ce2 := CustomError{Code: 10002, Err: nil}
	expected := "code: 10002"
	if ce2.Error() != expected {
		t.Errorf("Error() = %q, want %q", ce2.Error(), expected)
	}
}

func TestCustomError_Unwrap(t *testing.T) {
	err := errors.New("test error")
	ce := CustomError{Code: 10001, Err: err}
	if ce.Unwrap() != err {
		t.Errorf("Unwrap() = %v, want %v", ce.Unwrap(), err)
	}
}

func TestNewCustomError(t *testing.T) {
	err := errors.New("test error")
	ce := NewCustomError(10001, err)
	if ce.Code != 10001 || ce.Err != err {
		t.Error("NewCustomError 创建失败")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCustomError 未对错误码小于10001进行 panic")
		}
	}()
	_ = NewCustomError(10000, err)
}
