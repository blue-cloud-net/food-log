package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/model"
	"foodlog/server/internal/service"
)

// ExportHandler 数据导出/导入处理器
type ExportHandler struct {
	exportService *service.ExportService
}

func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

// Export 导出数据（format=json|csv）
func (h *ExportHandler) Export(c *gin.Context) {
	userID := httpx.GetUserID(c)
	switch c.DefaultQuery("format", "json") {
	case "csv":
		data, err := h.exportService.ExportCSV(c.Request.Context(), userID)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		c.Header("Content-Disposition", `attachment; filename="recipes.csv"`)
		c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
	default:
		data, err := h.exportService.ExportData(c.Request.Context(), userID)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		c.Header("Content-Disposition", `attachment; filename="foodlog-backup.json"`)
		c.JSON(http.StatusOK, data)
	}
}

// Import 导入备份数据。
// body 直接为导出的 JSON 备份文件；mode 通过 query 指定（append 默认 / overwrite）
func (h *ExportHandler) Import(c *gin.Context) {
	var data model.ExportData
	if err := c.ShouldBindJSON(&data); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	mode := c.DefaultQuery("mode", "append")
	if mode != "append" && mode != "overwrite" {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "mode 仅支持 append 或 overwrite")
		return
	}

	userID := httpx.GetUserID(c)
	if err := h.exportService.ImportData(c.Request.Context(), userID, &data, mode); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"imported": true, "mode": mode})
}
