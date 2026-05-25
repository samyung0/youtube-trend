package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	GoogleClientID string
	GoogleSecret   string
	GoogleRedirect string
	YouTubeAPIKey  string
	DeepSeekAPIKey string
	FrontendURL    string
	AllowedOrigins string
}

func Load() (*Config, error) {
	env := os.Getenv("YOUTUBE_TREND_ENV")
	if "" == env {
		env = "development"
	}

	godotenv.Load("../.env") // The Original .env
	godotenv.Load("../.env." + env)
	if "test" != env {
		godotenv.Load("../.env.local")
	}
	godotenv.Load("../.env." + env + ".local")

	c := &Config{
		Port:           getEnv("PORT", "4450"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirect: getEnv("GOOGLE_REDIRECT_URL", "http://localhost:4450/api/auth/callback"),
		YouTubeAPIKey:  os.Getenv("YOUTUBE_API_KEY"),
		DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
