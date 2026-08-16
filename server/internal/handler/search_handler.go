package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
)

// SearchHandler 全局搜索处理器
type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search 全局搜索（菜谱/餐厅/菜品分组返回）
func (h *SearchHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "请输入搜索关键词")
		return
	}
	userID := httpx.GetUserID(c)
	result := h.searchService.Search(c.Request.Context(), userID, keyword)
	httpx.RespondOK(c, result)
}
