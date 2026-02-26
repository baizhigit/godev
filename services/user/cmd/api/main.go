package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/baizhigit/godev/shared/grpcx"
	"github.com/baizhigit/godev/shared/observability"
	userv1 "github.com/baizhigit/godev/shared/proto/gen/go/platform/user/v1"

	grpcserver "github.com/baizhigit/godev/services/user/internal/adapters/grpc"
	"github.com/baizhigit/godev/services/user/internal/adapters/postgres"
	"github.com/baizhigit/godev/services/user/internal/app"
	"github.com/baizhigit/godev/services/user/internal/config"
)

func main() {
	// ── Config ───────────────────────────────────────────────
	// первым делом — если конфиг невалиден, сразу падаем
	cfg := config.MustLoad()

	// ── Logger ───────────────────────────────────────────────
	// настраиваем до observability — чтобы логировать ошибки Setup
	level := slog.LevelInfo
	if cfg.IsDev() {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	slog.Info("starting service", "name", cfg.ServiceName, "version", cfg.Version, "env", cfg.Env)

	// ── Observability ────────────────────────────────────────
	ctx := context.Background()

	var shutdownOtel func(context.Context) error
	if cfg.OTEL.Enabled {
		var err error
		shutdownOtel, err = observability.Setup(ctx, observability.Config{
			ServiceName: cfg.ServiceName,
			Version:     cfg.Version,
			Endpoint:    cfg.OTEL.Endpoint,
		})
		if err != nil {
			slog.Error("observability setup failed", "err", err)
			os.Exit(1)
		}
	}

	// ── Database ─────────────────────────────────────────────
	pool, err := postgres.NewPool(ctx, postgres.DBConfig{
		URL:      cfg.DB.URL,
		MaxConns: cfg.DB.MaxConns,
		MinConns: cfg.DB.MinConns,
	})
	if err != nil {
		slog.Error("database init failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("database connected")

	// ── Adapters → App ───────────────────────────────────────
	userRepo := postgres.NewUserRepository(pool)

	// ── Use cases ────────────────────────────────────────────
	handlers := app.Handlers{
		CreateUser: app.NewCreateUserHandler(userRepo),
		GetUser:    app.NewGetUserHandler(userRepo),
		ListUsers:  app.NewListUsersHandler(userRepo),
		UpdateUser: app.NewUpdateUserHandler(userRepo),
		DeleteUser: app.NewDeleteUserHandler(userRepo),
	}

	// ── gRPC Server ──────────────────────────────────────────
	grpcServer := grpcx.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, grpcserver.NewServer(handlers))

	// ── gRPC-Gateway ─────────────────────────────────────────
	gwMux := runtime.NewServeMux()
	gwAddr := fmt.Sprintf("localhost:%d", cfg.GRPC.Port)
	if err := userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		gwAddr,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		slog.Error("gateway registration failed", "err", err)
		os.Exit(1)
	}

	// ── Graceful shutdown ────────────────────────────────────
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
		if err != nil {
			slog.Error("grpc listen failed", "err", err)
			os.Exit(1)
		}
		slog.Info("gRPC server started", "port", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "err", err)
		}
	}()

	go func() {
		slog.Info("HTTP gateway started", "port", cfg.HTTPPort)
		if err := grpcx.RunHTTPGateway(sigCtx, gwMux, cfg.HTTPPort); err != nil {
			slog.Error("HTTP gateway failed", "err", err)
		}
	}()

	// ждём сигнала
	<-sigCtx.Done()
	slog.Info("shutdown signal received")

	// даём in-flight запросам завершиться
	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped")

	if shutdownOtel != nil {
		otelCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := shutdownOtel(otelCtx); err != nil {
			slog.Error("otel shutdown failed", "err", err)
		}
	}

	slog.Info("service stopped")
}
