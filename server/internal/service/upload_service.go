package service

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// 图片处理参数
const (
	MaxImageWidth  = 1920 // 原图最大宽度
	ThumbWidth     = 400  // 缩略图宽度
	JPEGQuality    = 82   // JPEG 压缩质量
	MaxUploadBytes = 20 << 20 // 单文件 20MB
)

// UploadResult 上传结果
type UploadResult struct {
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url"`
}

// UploadService 图片上传服务
type UploadService struct {
	uploadDir string
}

func NewUploadService(uploadDir string) *UploadService {
	return &UploadService{uploadDir: uploadDir}
}

// SaveImages 保存并处理多张图片
func (s *UploadService) SaveImages(files []*multipart.FileHeader) ([]UploadResult, error) {
	var results []UploadResult
	for _, fh := range files {
		res, err := s.saveOne(fh)
		if err != nil {
			// 清理已保存的文件
			for _, r := range results {
				os.Remove(filepath.Join(s.uploadDir, r.URL))
				if r.ThumbURL != "" {
					os.Remove(filepath.Join(s.uploadDir, r.ThumbURL))
				}
			}
			return nil, err
		}
		results = append(results, *res)
	}
	return results, nil
}

func (s *UploadService) saveOne(fh *multipart.FileHeader) (*UploadResult, error) {
	// 大小校验
	if fh.Size > MaxUploadBytes {
		return nil, fmt.Errorf("文件 %s 超过 20MB 限制", fh.Filename)
	}

	// MIME 校验
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		return nil, fmt.Errorf("不支持的图片格式: %s", ext)
	}

	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 解码图片
	src, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("无法解析图片: %w", err)
	}

	// 按日期分目录
	dateDir := time.Now().Format("2006/01")
	dir := filepath.Join(s.uploadDir, dateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	name := uuid.NewString()
	origPath := filepath.Join(dir, name+".jpg")
	thumbPath := filepath.Join(dir, name+"_thumb.jpg")

	// 原图：缩放到最大宽度并转 JPEG
	img := imaging.Fit(src, MaxImageWidth, 0, imaging.Lanczos)
	if err := saveJPEG(img, origPath); err != nil {
		return nil, err
	}

	// 缩略图：400px 宽
	thumb := imaging.Fit(src, ThumbWidth, 0, imaging.Lanczos)
	if err := saveJPEG(thumb, thumbPath); err != nil {
		os.Remove(origPath)
		return nil, err
	}

	url := "/uploads/" + dateDir + "/" + name + ".jpg"
	thumbURL := "/uploads/" + dateDir + "/" + name + "_thumb.jpg"
	log.Printf("📸 上传图片: %s (%dx%d)", url, img.Bounds().Dx(), img.Bounds().Dy())

	return &UploadResult{URL: url, ThumbURL: thumbURL}, nil
}

func saveJPEG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: JPEGQuality})
}
