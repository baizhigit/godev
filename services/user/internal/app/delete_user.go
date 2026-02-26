package app

import (
	"context"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/services/user/internal/ports"
)

type DeleteUserCommand struct {
	ID string
}

type DeleteUserHandler struct {
	repo ports.UserRepository
}

func NewDeleteUserHandler(repo ports.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{repo: repo}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) error {
	// проверяем существование перед удалением
	// чтобы вернуть 404 а не молчаливый успех
	if _, err := h.repo.FindByID(ctx, domain.UserID(cmd.ID)); err != nil {
		return domain.ErrUserNotFound.WithCause(err)
	}

	if err := h.repo.Delete(ctx, domain.UserID(cmd.ID)); err != nil {
		return domain.ErrInternal.WithCause(err)
	}
	return nil
}
