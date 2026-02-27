package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcserver "github.com/baizhigit/godev/services/user/internal/adapters/grpc"
	"github.com/baizhigit/godev/services/user/internal/adapters/httpserver"
	"github.com/baizhigit/godev/services/user/internal/adapters/memory"
	"github.com/baizhigit/godev/services/user/internal/adapters/postgres"
	"github.com/baizhigit/godev/services/user/internal/app"
	"github.com/baizhigit/godev/services/user/internal/config"
	"github.com/baizhigit/godev/services/user/internal/ports"
	"github.com/baizhigit/godev/shared/grpcx"
	userv1 "github.com/baizhigit/godev/shared/proto/gen/go/platform/user/v1"
)

func main() {
	// ── Config ───────────────────────────────────────────────
	cfg := config.MustLoad()

	// ── Logger ───────────────────────────────────────────────
	level := slog.LevelInfo
	if cfg.IsDev() {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	slog.Info("starting service",
		"name", cfg.ServiceName,
		"version", cfg.Version,
		"env", cfg.Env,
	)

	// ctx для всего lifecycle приложения
	ctx := context.Background()

	// ── Repository ───────────────────────────────────────────
	var userRepo ports.UserRepository

	if cfg.HasDB() {
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

		if err := postgres.RunMigrations(pool); err != nil {
			slog.Error("migrations failed", "err", err)
			os.Exit(1)
		}

		userRepo = postgres.NewUserRepository(pool)
		slog.Info("using postgres repository")
	} else {
		userRepo = memory.NewUserRepository()
		slog.Warn("using in-memory repository — data will be lost on restart")
	}

	// ── Use cases ────────────────────────────────────────────
	handlers := app.Handlers{
		CreateUser: app.NewCreateUserHandler(userRepo),
		GetUser:    app.NewGetUserHandler(userRepo),
		ListUsers:  app.NewListUsersHandler(userRepo),
		UpdateUser: app.NewUpdateUserHandler(userRepo),
		DeleteUser: app.NewDeleteUserHandler(userRepo),
	}

	// ── gRPC ─────────────────────────────────────────────────
	grpcServer := grpcx.NewServer()
	// RegisterXxxServer — ОБЯЗАТЕЛЬНО до RunGRPC
	// reflection.Register внутри RunGRPC увидит все сервисы
	userv1.RegisterUserServiceServer(grpcServer, grpcserver.NewServer(handlers))

	// ── gRPC-Gateway ─────────────────────────────────────────
	gwMux := runtime.NewServeMux()
	if err := userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwMux,
		fmt.Sprintf("localhost:%d", cfg.GRPC.Port),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		slog.Error("gateway registration failed", "err", err)
		os.Exit(1)
	}

	// ── HTTP router ──────────────────────────────────────────
	httpMux := http.NewServeMux()
	httpMux.Handle("/docs", httpserver.SwaggerHandler())
	httpMux.Handle("/swagger.json", httpserver.SwaggerHandler())
	httpMux.Handle("/", gwMux)

	// ── Signal context ───────────────────────────────────────
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// ── Запуск через errgroup ────────────────────────────────
	g, gCtx := errgroup.WithContext(sigCtx)

	g.Go(func() error {
		return grpcx.RunGRPC(gCtx, grpcServer, cfg.GRPC.Port)
	})

	g.Go(func() error {
		return grpcx.RunHTTPGateway(gCtx, httpMux, cfg.HTTPPort)
	})

	slog.Info("service ready",
		"grpc", fmt.Sprintf("localhost:%d", cfg.GRPC.Port),
		"http", fmt.Sprintf("http://localhost:%d", cfg.HTTPPort),
		"docs", fmt.Sprintf("http://localhost:%d/docs", cfg.HTTPPort),
	)

	if err := g.Wait(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}

	slog.Info("service stopped")
}
