// Package storage 负责数据目录（凭证 / 上传中转 / 图片）的布局与维护
package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"foodlog/server/internal/config"
)

// 图片用途（数据目录 images/ 下的子目录名）
const (
	KindRecipe     = "recipe"
	KindRestaurant = "restaurant"
	KindInventory  = "inventory"
)

// Kinds 支持的图片用途
var Kinds = []string{KindRecipe, KindRestaurant, KindInventory}

// IsValidKind 校验图片用途是否受支持
func IsValidKind(kind string) bool {
	for _, k := range Kinds {
		if k == kind {
			return true
		}
	}
	return false
}

// EnsureDirs 创建数据目录骨架
// 目录结构：{DATA_DIR}/{credentials,tmp,images/{recipe,restaurant,inventory}}
func EnsureDirs(cfg *config.Config) error {
	type dir struct {
		path string
		perm os.FileMode
	}

	dirs := []dir{
		{cfg.CredentialsDir(), 0o700}, // 含明文密码，权限收紧
		{cfg.TmpDir(), 0o755},
	}
	for _, kind := range Kinds {
		dirs = append(dirs, dir{cfg.ImagesDir(kind), 0o755})
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d.path, d.perm); err != nil {
			return fmt.Errorf("创建数据目录 %s 失败: %w", d.path, err)
		}
	}
	return nil
}

// CleanTmp 清空上传中转目录
// 在服务启动时调用（此时不存在并发上传），用于回收上次异常退出留下的残留文件
func CleanTmp(cfg *config.Config) error {
	entries, err := os.ReadDir(cfg.TmpDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取上传临时目录失败: %w", err)
	}

	cleaned := 0
	for _, entry := range entries {
		p := filepath.Join(cfg.TmpDir(), entry.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("清理上传临时文件 %s 失败: %w", p, err)
		}
		cleaned++
	}
	if cleaned > 0 {
		log.Printf("🧹 已清理上传临时残留文件 %d 个", cleaned)
	}
	return nil
}
