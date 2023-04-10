package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	RunAddress           string        `env:"RUN_ADDRESS"`
	DatabaseURI          string        `env:"DATABASE_URI"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSigningKey        string        `env:"JWT_SIGNING_KEY" envDefault:"secret42"`
	JWTExpireDuration    time.Duration `env:"JWT_EXPIRE_DURATION" envDefault:"24h"`
}

func NewConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if cfg.RunAddress == "" {
		flag.StringVar(&cfg.RunAddress, "a", "", "define service address and port")
	}
	if cfg.DatabaseURI == "" {
		flag.StringVar(&cfg.DatabaseURI, "d", "", "define database connection path")
	}
	if cfg.AccrualSystemAddress == "" {
		flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "define accrual system address")
	}
	flag.Parse()
	return &cfg, nil
}
