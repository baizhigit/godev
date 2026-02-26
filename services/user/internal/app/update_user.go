package app

import (
	"context"
	"time"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/services/user/internal/ports"
)

type UpdateUserCommand struct {
	ID     string
	Fields domain.UpdateFields
}

type UpdateUserHandler struct {
	repo ports.UserRepository
}

func NewUpdateUserHandler(repo ports.UserRepository) *UpdateUserHandler {
	return &UpdateUserHandler{repo: repo}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (domain.User, error) {
	user, err := h.repo.FindByID(ctx, domain.UserID(cmd.ID))
	if err != nil {
		return domain.User{}, domain.ErrUserNotFound
	}

	// применяем только те поля которые пришли
	if cmd.Fields.FirstName != nil {
		user.FirstName = *cmd.Fields.FirstName
	}
	if cmd.Fields.LastName != nil {
		user.LastName = *cmd.Fields.LastName
	}
	if cmd.Fields.Email != nil {
		user.Email = *cmd.Fields.Email
	}
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(ctx, user); err != nil {
		return domain.User{}, domain.ErrInternal.WithCause(err)
	}
	return user, nil
}
