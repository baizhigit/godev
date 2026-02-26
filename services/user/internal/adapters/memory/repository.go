package memory

import (
	"context"
	"sync"

	"github.com/baizhigit/godev/services/user/internal/domain"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[domain.UserID]domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[domain.UserID]domain.User),
	}
}

func (r *UserRepository) Save(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) FindByID(_ context.Context, id domain.UserID) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, email domain.Email) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (r *UserRepository) List(_ context.Context, pageSize int, _ string) ([]domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, "", nil
}

func (r *UserRepository) Update(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; !ok {
		return domain.ErrUserNotFound
	}
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) Delete(_ context.Context, id domain.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}
