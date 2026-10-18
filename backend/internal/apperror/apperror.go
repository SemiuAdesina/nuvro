package apperror

import "errors"

type Code string

const (
	CodeInvalidInput        Code = "validation_error"
	CodeUnauthenticated     Code = "authentication_required"
	CodeForbidden           Code = "forbidden"
	CodeNotFound            Code = "not_found"
	CodeConflict            Code = "conflict"
	CodeRateLimited         Code = "rate_limited"
	CodeUnavailable         Code = "upstream_unavailable"
	CodeProviderTimeout     Code = "provider_timeout"
	CodeInvalidState        Code = "invalid_state"
	CodeIdempotencyConflict Code = "idempotency_conflict"
	CodeInternal            Code = "internal_error"
)

const fallbackMessage = "The request could not be completed."

type Error struct {
	code    Code
	message string
	cause   error
}

func New(code Code, message string) *Error {
	return &Error{code: code, message: message}
}

func Wrap(code Code, message string, cause error) *Error {
	return &Error{code: code, message: message, cause: cause}
}

func (e *Error) Error() string {
	if e.cause == nil {
		return string(e.code) + ": " + e.message
	}
	return string(e.code) + ": " + e.message + ": " + e.cause.Error()
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Code() Code {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func CodeOf(err error) Code {
	var applicationError *Error
	if errors.As(err, &applicationError) {
		switch applicationError.code {
		case CodeInvalidInput, CodeUnauthenticated, CodeForbidden, CodeNotFound, CodeConflict, CodeRateLimited, CodeUnavailable, CodeProviderTimeout, CodeInvalidState, CodeIdempotencyConflict:
			return applicationError.code
		}
	}
	return CodeInternal
}

func SafeMessage(err error) string {
	var applicationError *Error
	if errors.As(err, &applicationError) && CodeOf(err) != CodeInternal {
		return applicationError.message
	}
	return fallbackMessage
}
