// adapters/postgres/mapping.go
package postgres

import (
	"encoding/base64"
	"strconv"

	"github.com/baizhigit/godev/services/user/internal/adapters/postgres/sqlcgen"
	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/google/uuid"
)

// sqlcgen → domain
func toDomain(row sqlc.User) domain.User {
	return domain.User{
		ID:        domain.UserID(row.ID.String()), // uuid.UUID → string
		Email:     domain.Email(row.Email),
		FirstName: row.FirstName,
		LastName:  row.LastName,
		CreatedAt: row.CreatedAt, // уже time.Time, без .Time
		UpdatedAt: row.UpdatedAt, // уже time.Time
	}
}

// domain → sqlcgen params
func toInsertParams(u domain.User) sqlc.InsertUserParams {
	return sqlc.InsertUserParams{
		ID:        uuid.MustParse(string(u.ID)), // string → uuid.UUID
		Email:     string(u.Email),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt, // time.Time напрямую
		UpdatedAt: u.UpdatedAt,
	}
}

func toUpdateParams(u domain.User) sqlc.UpdateUserParams {
	return sqlc.UpdateUserParams{
		ID:        uuid.MustParse(string(u.ID)),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     string(u.Email),
	}
}

// page token helpers
func encodePageToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodePageToken(token string) int {
	if token == "" {
		return 0
	}
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	offset, err := strconv.Atoi(string(b))
	if err != nil {
		return 0
	}
	return offset
}
