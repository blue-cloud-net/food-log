package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"foodlog/server/internal/config"
)

// BootstrapAdmin 确保生产环境存在管理员账号
//
// 密码取值优先级：
//  1. 环境变量 ADMIN_PASSWORD
//  2. 数据目录凭证文件 credentials/admin-password.txt（已存在则复用）
//  3. 生成 UUIDv7 并写入该文件（权限 0600）
//
// 管理员已存在时不做任何修改。开发模式下账号由演示数据提供，直接跳过。
func BootstrapAdmin(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	if cfg.IsDev() {
		log.Println("⏭️  开发模式：管理员账号来自演示数据（admin / admin）")
		return nil
	}

	username := cfg.AdminUsername
	exists, err := userExists(ctx, pool, username)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("✅ 管理员账号 %s 已存在，跳过引导", username)
		return nil
	}

	password, err := resolveAdminPassword(cfg)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// users.email 为 NOT NULL UNIQUE，而登录仅使用 username，此邮箱仅作占位
	email := username + "@example.com"
	tag, err := pool.Exec(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		username, email, string(hash))
	if err != nil {
		return fmt.Errorf("创建管理员账号失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		log.Printf("✅ 管理员账号 %s 已存在，跳过引导", username)
		return nil
	}

	log.Printf("✅ 已创建管理员账号 %s", username)
	return nil
}

// userExists 判断用户名是否已存在
func userExists(ctx context.Context, pool *pgxpool.Pool, username string) (bool, error) {
	var exists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("查询管理员账号失败: %w", err)
	}
	return exists, nil
}

// resolveAdminPassword 解析管理员密码，必要时生成随机密码并落盘
func resolveAdminPassword(cfg *config.Config) (string, error) {
	if pwd := strings.TrimSpace(cfg.AdminPassword); pwd != "" {
		return pwd, nil
	}

	path := cfg.AdminPasswordFile()
	content, err := os.ReadFile(path)
	switch {
	case err == nil:
		if pwd := strings.TrimSpace(string(content)); pwd != "" {
			log.Printf("🔑 使用已生成的随机密码：%s", path)
			return pwd, nil
		}
	case !os.IsNotExist(err):
		return "", fmt.Errorf("读取凭证文件 %s 失败: %w", path, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("生成随机密码失败: %w", err)
	}
	password := id.String()

	if err := os.WriteFile(path, []byte(password+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("写入凭证文件 %s 失败: %w", path, err)
	}
	log.Printf("🔑 未设置 ADMIN_PASSWORD，已生成随机管理员密码（见 %s）", path)
	return password, nil
}
