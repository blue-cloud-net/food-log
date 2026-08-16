package config

import (
	"os"
	"strings"
)

// Config 应用配置
type Config struct {
	ServerPort   string
	DatabaseURL  string
	JWTSecret    string
	UploadDir    string
	ClientOrigin []string
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://foodlog:foodlog123@localhost:5432/foodlog?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "foodlog-secret-change-me"),
		UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
		ClientOrigin: splitOrigins(getEnv("CLIENT_ORIGIN", "http://localhost:5173")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitOrigins(s string) []string {
	parts := strings.Split(s, ",")
	var origins []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}
