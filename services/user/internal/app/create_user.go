package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/services/user/internal/ports"
)

type CreateUserCommand struct {
	Email     string
	FirstName string
	LastName  string
}

type CreateUserHandler struct {
	repo ports.UserRepository
}

func NewCreateUserHandler(repo ports.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{repo: repo}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (domain.User, error) {
	email := domain.Email(strings.ToLower(cmd.Email))

	if _, err := h.repo.FindByEmail(ctx, email); err == nil {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	now := time.Now()
	user := domain.User{
		ID:        domain.UserID(uuid.New().String()),
		Email:     email,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return user, h.repo.Save(ctx, user)
}
