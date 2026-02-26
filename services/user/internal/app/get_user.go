package app

import (
	"context"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/services/user/internal/ports"
)

type GetUserQuery struct {
	ID string
}

type GetUserHandler struct {
	repo ports.UserRepository
}

func NewGetUserHandler(repo ports.UserRepository) *GetUserHandler {
	return &GetUserHandler{repo: repo}
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (domain.User, error) {
	user, err := h.repo.FindByID(ctx, domain.UserID(q.ID))
	if err != nil {
		// репозиторий вернул свою ошибку — оборачиваем в доменную
		return domain.User{}, domain.ErrUserNotFound.WithCause(err)
	}
	return user, nil
}
