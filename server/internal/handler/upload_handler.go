package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
	"foodlog/server/internal/storage"
)

// UploadHandler 图片上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload 上传图片（支持多文件）
//
// 表单字段：
//
//	images  图片文件，最多 9 张
//	type    图片用途：recipe（默认）| restaurant，决定落地子目录
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

	kind := strings.TrimSpace(c.PostForm("type"))
	if kind == "" {
		kind = storage.KindRecipe
	}
	if !storage.IsValidKind(kind) {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, "不支持的图片用途: "+kind)
		return
	}

	results, err := h.uploadService.SaveImages(kind, files)
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeUploadFailed, err.Error())
		return
	}

	httpx.RespondOK(c, gin.H{"files": results})
}
