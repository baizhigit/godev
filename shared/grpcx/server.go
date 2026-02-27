package grpcx

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	defaults := []grpc.ServerOption{
		UnaryInterceptors(),
		StreamInterceptors(),
	}
	return grpc.NewServer(append(defaults, opts...)...)
}

// RunGRPC — принимает ctx для graceful shutdown
func RunGRPC(ctx context.Context, srv *grpc.Server, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("grpcx: listen: %w", err)
	}

	// health check — до reflection
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// reflection — после всех RegisterXxxServer вызовов
	reflection.Register(srv)

	slog.Info("gRPC server started", "port", port)

	// слушаем ctx для graceful shutdown
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(lis) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpcx: serve: %w", err)
	case <-ctx.Done():
		slog.Info("gRPC graceful stop...")
		srv.GracefulStop()
		return nil
	}
}

func RunHTTPGateway(ctx context.Context, handler http.Handler, port int) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpcx: gateway: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
