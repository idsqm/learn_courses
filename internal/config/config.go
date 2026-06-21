package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	ServerPort string `env:"SERVER_PORT,default=8081"`
	DBURL      string `env:"DB_URL,required"`
	JWTSecret  string `env:"JWT_SECRET,required"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}
