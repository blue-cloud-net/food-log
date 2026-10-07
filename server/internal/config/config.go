package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 运行模式
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// 数据目录下的固定名称
const (
	credentialsDirName = "credentials"
	tmpDirName         = "tmp"
	imagesDirName      = "images"
	adminPasswordFile  = "admin-password.txt"
)

// Config 应用配置
type Config struct {
	AppEnv       string
	ServerPort   string
	DataDir      string
	DatabaseURL  string
	JWTSecret    string
	ClientOrigin []string

	// 管理员引导（仅生产模式生效）
	AdminUsername string
	AdminPassword string

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
		AppEnv:       getEnv("APP_ENV", EnvProduction),
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DataDir:      getEnv("DATA_DIR", "./data"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://foodlog:foodlog123@localhost:5432/foodlog?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "foodlog-secret-change-me"),
		ClientOrigin: splitOrigins(getEnv("CLIENT_ORIGIN", "http://localhost:5173")),

		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),

		AIProvider: getEnv("AI_PROVIDER", ""),
		AIBaseURL:  getEnv("AI_BASE_URL", ""),
		AIAPIKey:   getEnv("AI_API_KEY", ""),
		AIModel:    getEnv("AI_MODEL", ""),
		AITimeout:  getDuration("AI_TIMEOUT", 3),
	}
}

// IsDev 是否开发模式（仅开发模式加载演示数据）
func (c *Config) IsDev() bool {
	return c.AppEnv == EnvDevelopment
}

// CredentialsDir 凭证目录（自动生成的 admin 密码等）
func (c *Config) CredentialsDir() string {
	return filepath.Join(c.DataDir, credentialsDirName)
}

// AdminPasswordFile 自动生成的 admin 密码文件路径
func (c *Config) AdminPasswordFile() string {
	return filepath.Join(c.CredentialsDir(), adminPasswordFile)
}

// TmpDir 上传中转临时目录
func (c *Config) TmpDir() string {
	return filepath.Join(c.DataDir, tmpDirName)
}

// ImagesRoot 图片根目录
func (c *Config) ImagesRoot() string {
	return filepath.Join(c.DataDir, imagesDirName)
}

// ImagesDir 指定用途的图片目录（kind: recipe | restaurant）
func (c *Config) ImagesDir(kind string) string {
	return filepath.Join(c.ImagesRoot(), kind)
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
