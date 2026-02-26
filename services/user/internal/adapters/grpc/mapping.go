package grpc

import (
	"errors"

	"github.com/baizhigit/godev/services/user/internal/domain"
	"github.com/baizhigit/godev/shared/errx"
	userv1 "github.com/baizhigit/godev/shared/proto/gen/go/platform/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto(u domain.User) *userv1.User {
	return &userv1.User{
		Id:        string(u.ID),
		Email:     string(u.Email),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	var e *errx.Error
	if !errors.As(err, &e) {
		return status.Error(codes.Internal, "internal error")
	}

	code := map[int]codes.Code{
		400: codes.InvalidArgument,
		404: codes.NotFound,
		409: codes.AlreadyExists,
		500: codes.Internal,
	}[e.Status]

	return status.Error(code, e.Message)
}
