package app

import (
	"context"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/services/user/internal/ports"
)

type ListUsersQuery struct {
	PageSize  int
	PageToken string
}

type ListUsersResult struct {
	Users         []domain.User
	NextPageToken string
}

type ListUsersHandler struct {
	repo ports.UserRepository
}

func NewListUsersHandler(repo ports.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{repo: repo}
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (ListUsersResult, error) {
	// дефолт если клиент не передал page_size
	if q.PageSize == 0 {
		q.PageSize = 20
	}

	users, nextToken, err := h.repo.List(ctx, q.PageSize, q.PageToken)
	if err != nil {
		return ListUsersResult{}, domain.ErrInternal.WithCause(err)
	}

	return ListUsersResult{
		Users:         users,
		NextPageToken: nextToken,
	}, nil
}
