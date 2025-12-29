package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomError_String(t *testing.T) {
	t.Run("with non-nil error", func(t *testing.T) {
		err := errors.New("test error")
		ce := CustomError{Code: 10001, Err: err}
		expected := fmt.Sprintf("code: %d, err: %s", 10001, err.Error())
		if ce.String() != expected {
			t.Errorf("String() = %q, want %q", ce.String(), expected)
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		ce := CustomError{Code: 10002, Err: nil}
		expected := "code: 10002, err: nil"
		if ce.String() != expected {
			t.Errorf("String() = %q, want %q", ce.String(), expected)
		}
	})
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
	t.Run("valid business error", func(t *testing.T) {
		err := errors.New("test error")
		ce := NewCustomError(10001, err)
		if ce.Code != 10001 || ce.Err != err {
			t.Error("NewCustomError failed to create a valid error")
		}
	})

	t.Run("panics with system error code", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewCustomError should have panicked with code less than 10001")
			}
		}()
		_ = NewCustomError(10000, errors.New("should panic"))
	})
}

func TestNewSystemError(t *testing.T) {
	t.Run("valid system error", func(t *testing.T) {
		err := errors.New("test error")
		ce := NewSystemError(1000, err)
		if ce.Code != 1000 || ce.Err != err {
			t.Error("NewSystemError failed to create a valid error")
		}
	})

	t.Run("panics with business error code", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewSystemError should have panicked with code greater than or equal to 10001")
			}
		}()
		_ = NewSystemError(10001, errors.New("should panic"))
	})
}

func TestIsCode(t *testing.T) {
	t.Run("error with matching code", func(t *testing.T) {
		err := NewCustomError(10001, errors.New("test"))
		assert.True(t, IsCode(err, 10001))
	})

	t.Run("error with non-matching code", func(t *testing.T) {
		err := NewCustomError(10001, errors.New("test"))
		assert.False(t, IsCode(err, 10002))
	})

	t.Run("wrapped error with matching code", func(t *testing.T) {
		innerErr := NewCustomError(10001, errors.New("inner"))
		wrappedErr := fmt.Errorf("wrapped: %w", innerErr)
		assert.True(t, IsCode(wrappedErr, 10001))
	})

	t.Run("wrapped error with non-matching code", func(t *testing.T) {
		innerErr := NewCustomError(10001, errors.New("inner"))
		wrappedErr := fmt.Errorf("wrapped: %w", innerErr)
		assert.False(t, IsCode(wrappedErr, 10002))
	})

	t.Run("non-custom error", func(t *testing.T) {
		err := errors.New("plain error")
		assert.False(t, IsCode(err, 10001))
	})

	t.Run("nil error", func(t *testing.T) {
		var err error = nil
		assert.False(t, IsCode(err, 10001))
	})
}
