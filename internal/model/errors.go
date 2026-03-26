package model

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrInvalidArgument ErrorCode = "invalid_argument"
	ErrNotFound        ErrorCode = "not_found"
	ErrUnsupported     ErrorCode = "unsupported"
	ErrConflict        ErrorCode = "conflict"
	ErrUnavailable     ErrorCode = "unavailable"
	ErrUnauthenticated ErrorCode = "unauthenticated"
	ErrInternal        ErrorCode = "internal"
	ErrPrecondition    ErrorCode = "failed_precondition"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Errorf(code ErrorCode, format string, args ...any) error {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func Wrap(code ErrorCode, err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

func CodeOf(err error) ErrorCode {
	if err == nil {
		return ""
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrInternal
}

func MessageOf(err error) string {
	if err == nil {
		return ""
	}
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		return appErr.Message
	}
	return err.Error()
}
