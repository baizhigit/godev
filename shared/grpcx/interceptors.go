// shared/grpcx/interceptors.go
package grpcx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"buf.build/go/protovalidate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UnaryInterceptors — цепочка для unary RPC (обычные запрос/ответ)
func UnaryInterceptors() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		RecoveryUnaryInterceptor(),   // 1. паника → error (всегда первый)
		TracingUnaryInterceptor(),    // 2. создаём span
		LoggingUnaryInterceptor(),    // 3. логируем (уже с trace_id из span)
		ValidationUnaryInterceptor(), // 4. валидация входящего запроса
	)
}

// StreamInterceptors — цепочка для streaming RPC
func StreamInterceptors() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(
		RecoveryStreamInterceptor(),
		TracingStreamInterceptor(),
		LoggingStreamInterceptor(),
	)
}

// ─── Recovery ────────────────────────────────────────────────────────────────

func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "grpc panic recovered",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				err = status.Errorf(grpccodes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func RecoveryStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("grpc stream panic recovered",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				err = status.Errorf(grpccodes.Internal, "internal error")
			}
		}()
		return handler(srv, ss)
	}
}

// ─── Tracing ─────────────────────────────────────────────────────────────────

func TracingUnaryInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("grpc")
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		span.SetAttributes(attribute.String("rpc.method", info.FullMethod))

		resp, err := handler(ctx, req)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
		return resp, err
	}
}

func TracingStreamInterceptor() grpc.StreamServerInterceptor {
	tracer := otel.Tracer("grpc")
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, span := tracer.Start(ss.Context(), info.FullMethod)
		defer span.End()

		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
		err := handler(srv, wrapped)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
}

// ─── Logging ─────────────────────────────────────────────────────────────────

func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		attrs := []any{
			"method", info.FullMethod,
			"duration_ms", time.Since(start).Milliseconds(),
			"code", grpccodes.Code(status.Code(err)).String(),
		}

		if err != nil {
			slog.ErrorContext(ctx, "grpc request failed", append(attrs, "error", err)...)
		} else {
			slog.InfoContext(ctx, "grpc request", attrs...)
		}

		return resp, err
	}
}

func LoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)

		slog.InfoContext(ss.Context(), "grpc stream",
			"method", info.FullMethod,
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
		)
		return err
	}
}

// ─── Validation ───────────────────────────────────────────────────────────────

func ValidationUnaryInterceptor() grpc.UnaryServerInterceptor {
	validator, err := protovalidate.New()
	if err != nil {
		panic(fmt.Sprintf("grpcx: failed to init protovalidate: %v", err))
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := validator.Validate(msg); err != nil {
				var valErr *protovalidate.ValidationError
				if errors.As(err, &valErr) {
					st, _ := status.New(grpccodes.InvalidArgument, "validation failed").
						WithDetails(valErr.ToProto())
					return nil, st.Err()
				}
				return nil, status.Error(grpccodes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// wrappedStream — позволяет подменить context у ServerStream
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }
