package app

type Handlers struct {
	CreateUser *CreateUserHandler
	GetUser    *GetUserHandler
	ListUsers  *ListUsersHandler
	UpdateUser *UpdateUserHandler
	DeleteUser *DeleteUserHandler
}
