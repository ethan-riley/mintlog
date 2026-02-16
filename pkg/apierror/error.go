package apierror

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func WithDetails(code int, message, details string) *Error {
	return &Error{Code: code, Message: message, Details: details}
}

func Write(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func BadRequest(msg string) *Error        { return New(http.StatusBadRequest, msg) }
func Unauthorized(msg string) *Error      { return New(http.StatusUnauthorized, msg) }
func Forbidden(msg string) *Error         { return New(http.StatusForbidden, msg) }
func NotFound(msg string) *Error          { return New(http.StatusNotFound, msg) }
func TooManyRequests(msg string) *Error   { return New(http.StatusTooManyRequests, msg) }
func Internal(msg string) *Error          { return New(http.StatusInternalServerError, msg) }
