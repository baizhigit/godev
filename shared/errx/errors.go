package errx

import (
	"errors"
	"fmt"
	"net/http"
)

// Error — структурированная ошибка с HTTP статусом и машиночитаемым кодом.
// Код домена (Code) отдаётся клиенту.
// Cause — внутренняя ошибка, только для логов, клиенту не передаётся.
type Error struct {
	Code    string // "USER_NOT_FOUND" — клиент может обрабатывать программно
	Message string // human-readable сообщение
	Status  int    // HTTP status code
	Cause   error  // оригинальная ошибка для логов
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %w", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// WithCause — добавить внутреннюю причину ошибки (для логов)
func (e *Error) WithCause(cause error) *Error {
	copy := *e
	copy.Cause = cause
	return &copy
}

// Is — позволяет использовать errors.Is для сравнения по Code
func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

// ─── Конструкторы ────────────────────────────────────────────────────────────

func New(status int, code, message string) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

func BadRequest(code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *Error {
	return New(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *Error {
	return New(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *Error {
	return New(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *Error {
	return New(http.StatusConflict, code, message)
}

func UnprocessableEntity(code, message string) *Error {
	return New(http.StatusUnprocessableEntity, code, message)
}

func Internal(code, message string) *Error {
	return New(http.StatusInternalServerError, code, message)
}

// ─── Хелперы ─────────────────────────────────────────────────────────────────

// As — удобный враппер для errors.As
func As(err error) (*Error, bool) {
	var e *Error
	return e, errors.As(err, &e)
}

// IsNotFound — быстрая проверка без errors.As
func IsNotFound(err error) bool {
	e, ok := As(err)
	return ok && e.Status == http.StatusNotFound
}

func IsConflict(err error) bool {
	e, ok := As(err)
	return ok && e.Status == http.StatusConflict
}
