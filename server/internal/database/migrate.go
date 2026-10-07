package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/migrations"
)

// migrationLockKey 迁移互斥锁标识，避免多实例同时初始化
const migrationLockKey int64 = 8273611201

// Apply 执行数据库迁移（服务启动时调用）
//
// 行为：
//   - 空库（不存在 users 表）→ 全量初始化
//   - 已初始化 → 仅执行尚未记录在 schema_migrations 中的文件
//   - migrations.DemoSeedFile 仅在开发模式（isDev）下执行，生产环境跳过
//
// 每个文件在独立事务内执行，成功后写入 schema_migrations 记账，因此可重复调用。
func Apply(ctx context.Context, pool *pgxpool.Pool, isDev bool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer conn.Release()

	// 会话级互斥锁：同一时刻只允许一个实例执行迁移
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("获取迁移锁失败: %w", err)
	}
	defer func() {
		if _, err := conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLockKey); err != nil {
			log.Printf("⚠️  释放迁移锁失败: %v", err)
		}
	}()

	fresh, err := isFreshDatabase(ctx, conn)
	if err != nil {
		return err
	}
	if fresh {
		log.Println("🆕 检测到空库，开始初始化...")
	}

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("初始化迁移标记表失败: %w", err)
	}

	applied, err := appliedFiles(ctx, conn)
	if err != nil {
		return err
	}

	files, err := migrations.List()
	if err != nil {
		return fmt.Errorf("读取内嵌迁移文件失败: %w", err)
	}

	for _, f := range files {
		if f.Name == migrations.DemoSeedFile && !isDev {
			log.Printf("⏭️  跳过 %s（演示数据仅开发模式加载）", f.Name)
			continue
		}
		if applied[f.Name] {
			log.Printf("⏭️  跳过 %s（已执行）", f.Name)
			continue
		}

		log.Printf("▶️  执行迁移 %s", f.Name)
		if err := applyFile(ctx, conn, f); err != nil {
			return err
		}
	}

	if fresh {
		log.Println("✅ 数据库初始化完成")
	} else {
		log.Println("✅ 数据库迁移检查完成")
	}
	return nil
}

// isFreshDatabase 判断是否为空库（业务表不存在）
func isFreshDatabase(ctx context.Context, conn *pgxpool.Conn) (bool, error) {
	var present bool
	if err := conn.QueryRow(ctx, "SELECT to_regclass('public.users') IS NOT NULL").Scan(&present); err != nil {
		return false, fmt.Errorf("检测数据库状态失败: %w", err)
	}
	return !present, nil
}

// appliedFiles 读取已执行迁移文件名集合
func appliedFiles(ctx context.Context, conn *pgxpool.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("读取迁移记录失败: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("解析迁移记录失败: %w", err)
		}
		applied[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取迁移记录失败: %w", err)
	}
	return applied, nil
}

// applyFile 在单个事务内执行迁移文件并记账
func applyFile(ctx context.Context, conn *pgxpool.Conn, f migrations.File) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开启事务失败 (%s): %w", f.Name, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 无参数调用走 simple protocol，支持文件内的多条语句与 DO $$ 块
	if _, err := tx.Exec(ctx, f.SQL); err != nil {
		return fmt.Errorf("执行迁移 %s 失败: %w", f.Name, err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", f.Name); err != nil {
		return fmt.Errorf("记录迁移 %s 失败: %w", f.Name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("提交迁移 %s 失败: %w", f.Name, err)
	}

	log.Printf("   ✅ %s", f.Name)
	return nil
}
