package domain

import "github.com/baizhigit/godev/shared/errx"

var (
	ErrUserNotFound      = errx.NotFound("USER_NOT_FOUND", "user not found")
	ErrUserAlreadyExists = errx.Conflict("USER_ALREADY_EXISTS", "user already exists")
	ErrInvalidEmail      = errx.BadRequest("INVALID_EMAIL", "invalid email format")
	ErrInternal          = errx.Internal("INTERNAL_ERROR", "internal server error")
)
