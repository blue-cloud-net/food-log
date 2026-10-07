package service

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	"foodlog/server/internal/config"
	"foodlog/server/internal/storage"
)

// 图片处理参数
const (
	MaxImageWidth  = 1920     // 原图最大宽度
	ThumbWidth     = 400      // 缩略图宽度
	JPEGQuality    = 82       // JPEG 压缩质量
	MaxUploadBytes = 20 << 20 // 单文件 20MB
)

// 允许的图片扩展名
var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}

// UploadResult 上传结果
type UploadResult struct {
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url"`
}

// UploadService 图片上传服务
//
// 文件布局：{DATA_DIR}/images/{kind}/YYYY/MM/{uuid}.jpg 与 {uuid}_thumb.jpg
// 中转布局：{DATA_DIR}/tmp  —— multipart 落盘与原子写入的中间文件，处理完即删
type UploadService struct {
	imagesRoot string
	tmpDir     string
}

func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{
		imagesRoot: cfg.ImagesRoot(),
		tmpDir:     cfg.TmpDir(),
	}
}

// SaveImages 保存并处理多张图片
// kind 为图片用途（storage.KindRecipe / storage.KindRestaurant），决定落地子目录
func (s *UploadService) SaveImages(kind string, files []*multipart.FileHeader) ([]UploadResult, error) {
	if !storage.IsValidKind(kind) {
		return nil, fmt.Errorf("不支持的图片用途: %s", kind)
	}

	results := make([]UploadResult, 0, len(files))
	var written []string

	for _, fh := range files {
		res, paths, err := s.saveOne(kind, fh)
		if err != nil {
			// 回滚本次请求已落位的文件
			for _, p := range written {
				os.Remove(p)
			}
			return nil, err
		}
		written = append(written, paths...)
		results = append(results, *res)
	}
	return results, nil
}

// saveOne 处理单张图片，返回结果与已落位的绝对路径（供失败回滚）
func (s *UploadService) saveOne(kind string, fh *multipart.FileHeader) (*UploadResult, []string, error) {
	if fh.Size > MaxUploadBytes {
		return nil, nil, fmt.Errorf("文件 %s 超过 20MB 限制", fh.Filename)
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExts[ext] {
		return nil, nil, fmt.Errorf("不支持的图片格式: %s", ext)
	}

	// 1) multipart 先落盘到 tmp，避免大图整体驻留内存
	tmpFile, err := s.spool(fh)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(tmpFile)

	src, err := decodeImage(tmpFile)
	if err != nil {
		return nil, nil, fmt.Errorf("无法解析图片 %s: %w", fh.Filename, err)
	}

	// 2) 目标目录：images/{kind}/YYYY/MM
	dateDir := time.Now().Format("2006/01")
	dir := filepath.Join(s.imagesRoot, kind, dateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("创建图片目录失败: %w", err)
	}

	name := uuid.NewString()
	origPath := filepath.Join(dir, name+".jpg")
	thumbPath := filepath.Join(dir, name+"_thumb.jpg")

	// 3) 编码到 tmp 再原子重命名，避免产生半成品文件
	orig, err := s.encodeAtomic(resizeToMaxWidth(src, MaxImageWidth), origPath)
	if err != nil {
		return nil, nil, err
	}
	thumb, err := s.encodeAtomic(imaging.Resize(src, ThumbWidth, 0, imaging.Lanczos), thumbPath)
	if err != nil {
		os.Remove(orig)
		return nil, nil, err
	}

	urlDir := "/images/" + kind + "/" + dateDir + "/"
	log.Printf("📸 上传图片[%s]: %s", kind, urlDir+name+".jpg")

	return &UploadResult{
		URL:      urlDir + name + ".jpg",
		ThumbURL: urlDir + name + "_thumb.jpg",
	}, []string{orig, thumb}, nil
}

// spool 将 multipart 文件写入 tmp 目录，返回临时文件绝对路径
func (s *UploadService) spool(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	path := filepath.Join(s.tmpDir, "upload-"+uuid.NewString())
	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	return path, nil
}

// encodeAtomic 将图片编码为 JPEG：先写入 tmp，再原子重命名到目标路径
func (s *UploadService) encodeAtomic(img image.Image, dst string) (string, error) {
	tmp, err := os.CreateTemp(s.tmpDir, "encode-*.jpg")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmp.Name()

	if err := jpeg.Encode(tmp, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("图片编码失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	// CreateTemp 权限为 0600，放宽为常规文件权限便于宿主机（bind mount）查看
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("写入图片失败: %w", err)
	}
	return dst, nil
}

// decodeImage 从文件解码图片
func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

// resizeToMaxWidth 按最大宽度等比缩放；未超过上限时保持原图
func resizeToMaxWidth(src image.Image, maxWidth int) image.Image {
	if src.Bounds().Dx() <= maxWidth {
		return src
	}
	return imaging.Resize(src, maxWidth, 0, imaging.Lanczos)
}
