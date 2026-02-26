package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/baizhigit/godev/services/user/internal/adapters/postgres/sqlcgen"
	"github.com/baizhigit/godev/services/user/internal/domain"
)

// UserRepository реализует ports.UserRepository
// знает про pgx и sqlcgen, но domain не знает про это
type UserRepository struct {
	q *sqlc.Queries
}

// NewUserRepository — то что вызывается в main.go
// принимает *pgxpool.Pool, создаёт sqlcgen.Queries
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: sqlc.New(pool)} // sqlcgen.New — сгенерированный конструктор
}

// ─── Реализация ports.UserRepository ─────────────────────────────────────────

func (r *UserRepository) Save(ctx context.Context, u domain.User) error {
	_, err := r.q.InsertUser(ctx, toInsertParams(u))
	if err != nil {
		return fmt.Errorf("UserRepository.Save: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, uuid.MustParse(string(id)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("UserRepository.FindByID: %w", err)
	}
	return toDomain(row), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, string(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("UserRepository.FindByEmail: %w", err)
	}
	return toDomain(row), nil
}

func (r *UserRepository) List(ctx context.Context, pageSize int, pageToken string) ([]domain.User, string, error) {
	offset := decodePageToken(pageToken)

	rows, err := r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, "", fmt.Errorf("UserRepository.List: %w", err)
	}

	total, err := r.q.CountUsers(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("UserRepository.List count: %w", err)
	}

	users := make([]domain.User, len(rows))
	for i, row := range rows {
		users[i] = toDomain(row)
	}

	nextToken := ""
	if nextOffset := offset + pageSize; int64(nextOffset) < total {
		nextToken = encodePageToken(nextOffset)
	}

	return users, nextToken, nil
}

func (r *UserRepository) Update(ctx context.Context, u domain.User) error {
	_, err := r.q.UpdateUser(ctx, toUpdateParams(u))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("UserRepository.Update: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id domain.UserID) error {
	if err := r.q.DeleteUser(ctx, uuid.MustParse(string(id))); err != nil {
		return fmt.Errorf("UserRepository.Delete: %w", err)
	}
	return nil
}
