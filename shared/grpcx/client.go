package grpcx

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClientConn — создаёт gRPC клиент с трейсингом
// используется когда один сервис вызывает другой
//
// conn, err := grpcx.NewClientConn(ctx, "user-service:50051")
// client := userv1.NewUserServiceClient(conn)
func NewClientConn(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaults := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // трейсинг клиентских вызовов
	}

	conn, err := grpc.NewClient(target, append(defaults, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("grpcx: dial %s: %w", target, err)
	}
	return conn, nil
}
