package domain

import "time"

type UserID string
type Email string

type User struct {
	ID        UserID
	Email     Email
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpdateFields — только те поля которые пришли в FieldMask
type UpdateFields struct {
	FirstName *string
	LastName  *string
	Email     *Email
}
