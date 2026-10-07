// Package web 内嵌前端构建产物并提供静态资源 + SPA 回退服务
//
// dist/ 目录在构建镜像或执行 scripts/build.sh 时被前端产物覆盖；
// 仓库中仅保留 .gitkeep 占位，保证未构建前端时也能通过 go build。
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// 缓存策略（与原 nginx.conf 保持一致）
const (
	cacheImmutable = "public, max-age=2592000, immutable" // 带 hash 的构建产物
	cacheIcon      = "public, max-age=604800"             // 应用图标
	cacheNoStore   = "no-cache, no-store, must-revalidate"
)

//go:embed all:dist
var embedded embed.FS

// distFS 以 dist 为根的文件系统
var distFS fs.FS

// indexHTML 首页内容；未构建前端时为占位提示页
var indexHTML []byte

// noFallbackPrefixes 这些路径缺失时必须返回 404，而不是回退到 index.html
var noFallbackPrefixes = []string{"assets/", "icons/"}

// noFallbackFiles 同上，针对具体文件
var noFallbackFiles = map[string]bool{
	"sw.js":                true,
	"manifest.webmanifest": true,
	"registerSW.js":        true,
}

func init() {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return
	}
	distFS = sub

	if content, err := fs.ReadFile(sub, "index.html"); err == nil {
		indexHTML = content
		return
	}
	indexHTML = []byte(`<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>美食日志</title></head>
<body style="font-family:system-ui;padding:2rem;line-height:1.6">
<h1>前端资源未构建</h1>
<p>当前二进制中没有前端产物（<code>dist/</code> 为空）。</p>
<ul>
  <li>开发环境请直接访问 Vite 开发服务器（默认 <code>http://localhost:5173</code>）</li>
  <li>生产构建请执行 <code>docker compose build</code> 或 <code>bash scripts/build.sh</code></li>
</ul>
</body></html>`)
}

// Handler 返回静态资源 + SPA 回退处理器
//
// 命中真实文件则带缓存头返回；未命中时仅对非资源路径回退到 index.html，
// 使前端路由（如 /recipes/xxx）支持深链接。
func Handler() http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.Trim(path.Clean(r.URL.Path), "/")
		if !fileExists(name) {
			if isAssetPath(name) {
				http.NotFound(w, r)
				return
			}
			serveIndex(w)
			return
		}

		setCacheHeaders(w, name)
		fileServer.ServeHTTP(w, r)
	})
}

// fileExists 判断 dist 中是否存在该文件（目录不算）
func fileExists(name string) bool {
	if name == "" || distFS == nil {
		return false
	}
	info, err := fs.Stat(distFS, name)
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

// serveIndex 返回首页（SPA 回退）
func serveIndex(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", cacheNoStore)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexHTML)
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
