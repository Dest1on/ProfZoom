package common

import "errors"

type ErrorCode string

const (
	CodeNotFound          ErrorCode = "not_found"
	CodeUnauthorized      ErrorCode = "unauthorized"
	CodeForbidden         ErrorCode = "forbidden"
	CodeConflict          ErrorCode = "conflict"
	CodeValidation        ErrorCode = "validation"
	CodeTelegramNotLinked ErrorCode = "telegram_not_linked"
	CodeRateLimited       ErrorCode = "rate_limited"
	CodeDeliveryFailed    ErrorCode = "delivery_failed"
	CodeInternal          ErrorCode = "internal"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Fields  map[string]string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func NewValidationError(message string, fields map[string]string) *AppError {
	return &AppError{Code: CodeValidation, Message: message, Fields: fields}
}

func Is(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
