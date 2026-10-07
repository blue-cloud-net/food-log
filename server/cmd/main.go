package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"foodlog/server/internal/config"
	"foodlog/server/internal/database"
	"foodlog/server/internal/router"
	"foodlog/server/internal/storage"
)

func main() {
	cfg := config.Load()

	// 准备数据目录（凭证 / 上传中转 / 图片）
	if err := storage.EnsureDirs(cfg); err != nil {
		log.Fatalf("%v", err)
	}
	// 清理上次异常退出残留的上传中转文件
	if err := storage.CleanTmp(cfg); err != nil {
		log.Fatalf("%v", err)
	}

	// 连接数据库
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer pool.Close()

	// 启动自愈：空库初始化 / 增量迁移 / 生产模式引导管理员账号
	if err := database.Apply(ctx, pool, cfg.IsDev()); err != nil {
		log.Fatalf("%v", err)
	}
	if err := database.BootstrapAdmin(ctx, pool, cfg); err != nil {
		log.Fatalf("%v", err)
	}

	// 启动 HTTP 服务
	r := router.Setup(cfg, pool)
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	log.Printf("🍽️  Food Log 后端启动: http://localhost:%s (mode: %s, data: %s)",
		cfg.ServerPort, cfg.AppEnv, cfg.DataDir)

	// 优雅退出
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务异常退出: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("✅ 服务已停止")
}
