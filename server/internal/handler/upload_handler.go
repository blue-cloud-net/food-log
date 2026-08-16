package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
)

// UploadHandler 图片上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload 上传图片（支持多文件）
func (h *UploadHandler) Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, "请选择图片文件")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, "未选择图片")
		return
	}
	if len(files) > 9 {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, "一次最多上传 9 张图片")
		return
	}

	results, err := h.uploadService.SaveImages(files)
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, err.Error())
		return
	}

	httpx.RespondOK(c, gin.H{"files": results})
}
