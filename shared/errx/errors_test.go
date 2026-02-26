package errx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/baizhigit/godev/shared/errx"
)

func TestError_Is(t *testing.T) {
	ErrNotFound := errx.NotFound("USER_NOT_FOUND", "user not found")
	wrapped := fmt.Errorf("repo layer: %w", ErrNotFound)

	// errors.Is работает через цепочку Unwrap
	if !errors.Is(wrapped, ErrNotFound) {
		t.Fatal("expected errors.Is to match")
	}
}

func TestError_WithCause(t *testing.T) {
	pgErr := errors.New("connection refused")
	err := errx.Internal("DB_ERROR", "database error").WithCause(pgErr)

	if !errors.Is(err, pgErr) {
		t.Fatal("expected cause to be unwrappable")
	}
}
