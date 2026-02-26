package grpcx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// NewServer — создаёт gRPC сервер с единым набором interceptors
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	defaults := []grpc.ServerOption{
		UnaryInterceptors(),
		StreamInterceptors(),
	}
	return grpc.NewServer(append(defaults, opts...)...)
}

// RunGRPC — запускает gRPC сервер, блокирует до ошибки
func RunGRPC(srv *grpc.Server, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("grpcx: listen: %w", err)
	}

	// reflection — позволяет grpcurl и Postman видеть методы без .proto файла
	// только для не-prod окружений, но в монорепо удобно всегда
	reflection.Register(srv)

	// стандартный health check (используется k8s livenessProbe через grpc)
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	return srv.Serve(lis)
}

// RunHTTPGateway — запускает grpc-gateway как HTTP/JSON прокси
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
