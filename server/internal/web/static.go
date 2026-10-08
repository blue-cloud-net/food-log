// Package web 提供前端构建产物（dist/）的静态资源服务与 SPA 深链接回退。
//
// 前端产物不再编入二进制（不使用 go:embed），而是运行时直接从磁盘目录读取，
// 目录由配置项 STATIC_DIR 指定：
//   - 容器运行：/app/dist（Dockerfile 由前端构建阶段拷贝而来）
//   - 本地构建：仓库根目录 ./dist（scripts/build.sh 产出）
//
// 目录不存在或缺少 index.html 时返回占位提示页，后端仍可独立提供 API。
package web

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 缓存策略（与原 nginx.conf 保持一致）
const (
	cacheImmutable = "public, max-age=2592000, immutable" // 带 hash 的构建产物
	cacheIcon      = "public, max-age=604800"             // 应用图标
	cacheNoStore   = "no-cache, no-store, must-revalidate"
)

// noFallbackPrefixes 这些路径缺失时必须返回 404，而不是回退到 index.html
var noFallbackPrefixes = []string{"assets/", "icons/"}

// noFallbackFiles 同上，针对具体文件
var noFallbackFiles = map[string]bool{
	"sw.js":                true,
	"manifest.webmanifest": true,
	"registerSW.js":        true,
}

// placeholderHTML 静态目录中没有前端产物时的占位提示页
var placeholderHTML = []byte(`<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>美食日志</title></head>
<body style="font-family:system-ui;padding:2rem;line-height:1.6">
<h1>前端资源未构建</h1>
<p>当前静态目录中没有前端产物（<code>index.html</code> 缺失）。</p>
<ul>
  <li>开发环境请直接访问 Vite 开发服务器（默认 <code>http://localhost:5173</code>）</li>
  <li>容器部署请执行 <code>docker compose build</code> 或 <code>bash scripts/build.sh</code></li>
  <li>本地运行二进制可用 <code>STATIC_DIR=/path/to/dist</code> 指定前端产物目录</li>
</ul>
</body></html>`)

// Handler 返回基于磁盘目录的静态资源 + SPA 回退处理器
//
// 命中真实文件则带缓存头返回；未命中时仅对非资源路径回退到 index.html，
// 使前端路由（如 /recipes/xxx）支持深链接。
func Handler(distDir string) http.Handler {
	root := filepath.Clean(distDir)
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}

	indexPath := filepath.Join(root, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		log.Printf("⚠️  前端静态目录不可用（%s）：未找到 index.html，页面请求将返回占位提示", root)
	}

	fileServer := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.Trim(path.Clean(r.URL.Path), "/")
		if !fileExists(root, name) {
			if isAssetPath(name) {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, indexPath)
			return
		}

		setCacheHeaders(w, name)
		fileServer.ServeHTTP(w, r)
	})
}

// fileExists 判断静态目录中是否存在该文件（目录不算）
func fileExists(root, name string) bool {
	if name == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(name)))
	return err == nil && !info.IsDir()
}

// isAssetPath 判断是否为必须存在、不允许回退的静态资源路径
func isAssetPath(name string) bool {
	if noFallbackFiles[name] || strings.HasPrefix(name, "workbox-") {
		return true
	}
	for _, prefix := range noFallbackPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// serveIndex 返回首页（SPA 回退）；index.html 缺失时返回占位提示页
func serveIndex(w http.ResponseWriter, indexPath string) {
	content, err := os.ReadFile(indexPath)
	if err != nil {
		content = placeholderHTML
	}
	w.Header().Set("Cache-Control", cacheNoStore)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// setCacheHeaders 按资源类型设置缓存头
func setCacheHeaders(w http.ResponseWriter, name string) {
	switch {
	case strings.HasPrefix(name, "assets/"):
		w.Header().Set("Cache-Control", cacheImmutable)
	case strings.HasPrefix(name, "icons/"):
		w.Header().Set("Cache-Control", cacheIcon)
	case name == "sw.js", name == "manifest.webmanifest", name == "registerSW.js",
		strings.HasPrefix(name, "workbox-"):
		w.Header().Set("Cache-Control", cacheNoStore)
	}
}
