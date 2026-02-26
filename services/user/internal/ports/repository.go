package ports

import (
	"context"

	"github.com/baizhigit/godev/services/user/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, u domain.User) error
	FindByID(ctx context.Context, id domain.UserID) (domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (domain.User, error)
	List(ctx context.Context, pageSize int, pageToken string) ([]domain.User, string, error)
	Update(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id domain.UserID) error
}
