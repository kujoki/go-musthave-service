package config

import "flag"

var (
    RunAddr string
    BaseURL string
)

func ParseFlags() {
    flag.StringVar(&RunAddr, "a", "localhost:8080", "HTTP server listen address")
    flag.StringVar(&BaseURL, "b", "http://localhost:8080", "Base URL used for short links")
    flag.Parse()
}
