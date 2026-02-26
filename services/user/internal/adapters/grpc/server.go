package grpc

import (
	"context"

	"github.com/baizhigit/godev/services/user/internal/app"
	"github.com/baizhigit/godev/services/user/internal/domain"
	userv1 "github.com/baizhigit/godev/shared/proto/gen/go/platform/user/v1"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	h app.Handlers
}

func NewServer(h app.Handlers) *Server {
	return &Server{h: h}
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user, err := s.h.CreateUser.Handle(ctx, app.CreateUserCommand{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userv1.CreateUserResponse{User: toProto(user)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	// разбираем FieldMask — обновляем только указанные поля
	fields := domain.UpdateFields{}
	for _, path := range req.UpdateMask.GetPaths() {
		switch path {
		case "first_name":
			fields.FirstName = &req.FirstName
		case "last_name":
			fields.LastName = &req.LastName
		case "email":
			e := domain.Email(req.Email)
			fields.Email = &e
		}
	}

	user, err := s.h.UpdateUser.Handle(ctx, app.UpdateUserCommand{
		ID:     req.Id,
		Fields: fields,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &userv1.UpdateUserResponse{User: toProto(user)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if err := s.h.DeleteUser.Handle(ctx, app.DeleteUserCommand{ID: req.Id}); err != nil {
		return nil, toGRPCError(err)
	}
	return &userv1.DeleteUserResponse{}, nil
}
