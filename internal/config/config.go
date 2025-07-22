package config

import "flag"

type Config struct {
	RunAddr string
	BaseURL string
}

func ParseFlags() *Config {
	var cfg Config
    flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "HTTP server listen address")
    flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL used for short links")
    flag.Parse()
	return &cfg
}
