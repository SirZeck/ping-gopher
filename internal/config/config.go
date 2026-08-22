package config

import (
	"flag"
	"os"
)

// Config holds runtime configuration options for PingGopher.
type Config struct {
	Role         string
	Port         string
	DatabasePath string
	RedisAddr    string
	JWTSecret    string
}

// LoadConfig initializes configuration from CLI flags and environment variables.
func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Role, "role", getEnv("ROLE", "all"), "Deployment role: all, api, worker, scheduler")
	flag.StringVar(&cfg.Port, "port", getEnv("PORT", "8080"), "HTTP API server port")
	flag.StringVar(&cfg.DatabasePath, "db", getEnv("DB_PATH", "pinggopher.db"), "SQLite DB path or Postgres DSN")
	flag.StringVar(&cfg.RedisAddr, "redis", getEnv("REDIS_ADDR", "localhost:6379"), "Redis broker address for gopher-queue")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", getEnv("JWT_SECRET", "pinggopher-secret-key-change-in-prod"), "JWT signing key")

	if !flag.Parsed() {
		flag.Parse()
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
