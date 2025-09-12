package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	ShortURLLen int `env:"SHORT_URL_LEN"`
	SecretToken string `env:"SECRET_KEY"`
	UserCookieName string `env:"USER_COOKIE_NAME"`
	ApplicationName string `env:"APP_NAME"`
}

func ParseFlags() *Config {
	var cfg Config

	cfg.RunAddr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "" //"./output.json"
	cfg.DatabaseDSN = "" //"user=postgres password=zmxncbv dbname=repository sslmode=disable host=localhost port=5432"
	cfg.ShortURLLen = 7
	cfg.UserCookieName = "auth_user"
	cfg.ApplicationName = "go_musthave_service"
	cfg.SecretToken = "secret"

	_ = env.Parse(&cfg)

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "HTTP server listen address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL used for short links")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File for output data")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DB for output data")
	flag.Parse()  
	return &cfg
}
