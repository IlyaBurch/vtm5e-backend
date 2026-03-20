package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	Port                 string
	AllowedOrigins       []string
	GoogleClientID       string
	GoogleClientSecret   string
	GoogleCallbackURL    string
	FrontendURL          string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Port:               os.Getenv("PORT"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleCallbackURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		FrontendURL:        os.Getenv("FRONTEND_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "3000"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		cfg.AllowedOrigins = []string{"http://localhost:5173"}
	} else {
		cfg.AllowedOrigins = strings.Split(allowedOrigin, ",")
	}

	if cfg.FrontendURL == "" {
		cfg.FrontendURL = "http://localhost:5173"
	}

	return cfg, nil
}
