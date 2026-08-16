package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用配置
type Config struct {
	ServerPort   string
	DatabaseURL  string
	JWTSecret    string
	UploadDir    string
	ClientOrigin []string

	// AI 配置（AIProvider: openai | ollama | "" 禁用）
	AIProvider string
	AIBaseURL  string
	AIAPIKey   string
	AIModel    string
	AITimeout  time.Duration
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://foodlog:foodlog123@localhost:5432/foodlog?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "foodlog-secret-change-me"),
		UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
		ClientOrigin: splitOrigins(getEnv("CLIENT_ORIGIN", "http://localhost:5173")),

		AIProvider: getEnv("AI_PROVIDER", ""),
		AIBaseURL:  getEnv("AI_BASE_URL", ""),
		AIAPIKey:   getEnv("AI_API_KEY", ""),
		AIModel:    getEnv("AI_MODEL", ""),
		AITimeout:  getDuration("AI_TIMEOUT", 3),
	}
}

// AIEnabled 是否启用了 AI 能力
func (c *Config) AIEnabled() bool {
	return c.AIProvider == "ollama" || (c.AIProvider == "openai" && c.AIAPIKey != "") ||
		(c.AIProvider == "" && c.AIAPIKey != "")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallbackSec int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return time.Duration(fallbackSec) * time.Second
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
