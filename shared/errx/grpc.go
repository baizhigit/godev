package errx

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCCode — маппинг HTTP статуса в gRPC код
func ToGRPCCode(e *Error) codes.Code {
	m := map[int]codes.Code{
		400: codes.InvalidArgument,
		401: codes.Unauthenticated,
		403: codes.PermissionDenied,
		404: codes.NotFound,
		409: codes.AlreadyExists,
		422: codes.InvalidArgument,
		500: codes.Internal,
	}
	if code, ok := m[e.Status]; ok {
		return code
	}
	return codes.Internal
}

// ToGRPCError — конвертация errx.Error в gRPC status error
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}
	e, ok := As(err)
	if !ok {
		return status.Error(codes.Internal, "internal error")
	}
	return status.Error(ToGRPCCode(e), e.Message)
}
