package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Service
	ServiceName string `env:"SERVICE_NAME" envDefault:"user"`
	Version     string `env:"VERSION"      envDefault:"dev"`
	Env         string `env:"ENV"          envDefault:"dev"`

	// HTTP Gateway
	HTTPPort int `env:"HTTP_PORT" envDefault:"8080"`

	// Database
	DB DBConfig `envPrefix:"DB_"`
	// OTEL
	OTEL OTELConfig `envPrefix:"OTEL_"`
	// gRPC
	GRPC GRPCConfig `envPrefix:"GRPC_"`
}

type DBConfig struct {
	URL      string `env:"URL,required"`              // → DB_URL
	MaxConns int32  `env:"MAX_CONNS" envDefault:"25"` // → DB_MAX_CONNS
	MinConns int32  `env:"MIN_CONNS" envDefault:"5"`  // → DB_MIN_CONNS
}

type OTELConfig struct {
	Endpoint string `env:"ENDPOINT" envDefault:"localhost:4318"` // → OTEL_ENDPOINT
	Enabled  bool   `env:"ENABLED"  envDefault:"true"`           // → OTEL_ENABLED
}

type GRPCConfig struct {
	Port int `env:"PORT" envDefault:"50051"` // → GRPC_PORT
}

// MustLoad — паникует если обязательные переменные не заданы.
// Вызывается один раз в main() до всего остального.
func MustLoad() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("config: %v", err))
	}
	return cfg
}

func (c Config) IsDev() bool  { return c.Env == "dev" }
func (c Config) IsProd() bool { return c.Env == "prod" }
