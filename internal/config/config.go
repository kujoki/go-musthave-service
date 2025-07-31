package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func ParseFlags() *Config {
	var cfg Config

	cfg.RunAddr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080"

	_ = env.Parse(&cfg)

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "HTTP server listen address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL used for short links")
	flag.Parse()  
	return &cfg
}
