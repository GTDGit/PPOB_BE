package gerbang

import "fmt"

// Error represents Gerbang API error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("gerbang error [%d]: %s", e.Code, e.Message)
}

// ParseError parses API response into error
func ParseError(resp *Response) error {
	return &Error{
		Code:    resp.Code,
		Message: resp.Message,
	}
}

// Common error codes
const (
	ErrCodeInvalidRequest     = 400
	ErrCodeUnauthorized       = 401
	ErrCodeForbidden          = 403
	ErrCodeNotFound           = 404
	ErrCodeConflict           = 409
	ErrCodeInsufficientFunds  = 422
	ErrCodeInternalError      = 500
	ErrCodeServiceUnavailable = 503
)

// IsInsufficientFunds checks if error is insufficient funds
func IsInsufficientFunds(err error) bool {
	if gerr, ok := err.(*Error); ok {
		return gerr.Code == ErrCodeInsufficientFunds
	}
	return false
}

// IsNotFound checks if error is not found
func IsNotFound(err error) bool {
	if gerr, ok := err.(*Error); ok {
		return gerr.Code == ErrCodeNotFound
	}
	return false
}

// IsUnauthorized checks if error is unauthorized
func IsUnauthorized(err error) bool {
	if gerr, ok := err.(*Error); ok {
		return gerr.Code == ErrCodeUnauthorized
	}
	return false
}
