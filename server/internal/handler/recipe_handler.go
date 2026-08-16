package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/ai"
	"foodlog/server/internal/httpx"
	"foodlog/server/internal/model"
	"foodlog/server/internal/service"
	"foodlog/server/internal/tagging"
)

// RecipeHandler 菜谱处理器
type RecipeHandler struct {
	recipeService *service.RecipeService
	aiProvider    ai.Provider // 可为 nil（AI 未配置）
}

func NewRecipeHandler(recipeService *service.RecipeService, aiProvider ai.Provider) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService, aiProvider: aiProvider}
}

type recipeRequest struct {
	Name            string            `json:"name" binding:"required,max=200"`
	Description     string            `json:"description"`
	Ingredients     []model.Ingredient `json:"ingredients"`
	Steps           []model.Step      `json:"steps"`
	CookTimeMinutes int               `json:"cook_time_minutes"`
	Difficulty      string            `json:"difficulty"`
	Rating          int               `json:"rating"`
	Tags            []string          `json:"tags"`
	Images          []string          `json:"images"`
}

// List 菜谱列表
func (h *RecipeHandler) List(c *gin.Context) {
	userID := httpx.GetUserID(c)
	q := parsePageQuery(c)

	list, err := h.recipeService.List(c.Request.Context(), userID, q,
		c.Query("keyword"), c.Query("difficulty"), c.Query("tag"), c.Query("sort"),
		c.Query("favorite") == "true")
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, list)
}

// Create 创建菜谱
func (h *RecipeHandler) Create(c *gin.Context) {
	var req recipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	if req.Name == "" {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "菜名不能为空")
		return
	}

	userID := httpx.GetUserID(c)
	rec := &model.Recipe{
		Name:            req.Name,
		Description:     req.Description,
		Ingredients:     req.Ingredients,
		Steps:           req.Steps,
		CookTimeMinutes: req.CookTimeMinutes,
		Difficulty:      req.Difficulty,
		Rating:          req.Rating,
		Tags:            req.Tags,
		Images:          req.Images,
	}

	result, err := h.recipeService.Create(c.Request.Context(), userID, rec)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// Get 菜谱详情
func (h *RecipeHandler) Get(c *gin.Context) {
	userID := httpx.GetUserID(c)
	rec, err := h.recipeService.Get(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, rec)
}

// Update 更新菜谱
func (h *RecipeHandler) Update(c *gin.Context) {
	var req recipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	rec := &model.Recipe{
		Name:            req.Name,
		Description:     req.Description,
		Ingredients:     req.Ingredients,
		Steps:           req.Steps,
		CookTimeMinutes: req.CookTimeMinutes,
		Difficulty:      req.Difficulty,
		Rating:          req.Rating,
		Tags:            req.Tags,
		Images:          req.Images,
	}

	result, err := h.recipeService.Update(c.Request.Context(), userID, c.Param("id"), rec)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// Delete 删除菜谱
func (h *RecipeHandler) Delete(c *gin.Context) {
	userID := httpx.GetUserID(c)
	if err := h.recipeService.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}

func parsePageQuery(c *gin.Context) *model.PageQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	q := &model.PageQuery{Page: page, PageSize: pageSize}
	q.Normalize()
	return q
}

// Favorite 收藏菜谱
func (h *RecipeHandler) Favorite(c *gin.Context) {
	userID := httpx.GetUserID(c)
	if err := h.recipeService.Favorite(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"favorited": true})
}

// Unfavorite 取消收藏
func (h *RecipeHandler) Unfavorite(c *gin.Context) {
	userID := httpx.GetUserID(c)
	if err := h.recipeService.Unfavorite(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"favorited": false})
}

// Random 随机选菜
func (h *RecipeHandler) Random(c *gin.Context) {
	userID := httpx.GetUserID(c)
	rec, err := h.recipeService.Random(c.Request.Context(), userID,
		c.Query("tag"), c.Query("difficulty"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, rec)
}

// TagList 预设标签词表
func (h *RecipeHandler) TagList(c *gin.Context) {
	httpx.RespondOK(c, tagging.TagCategories)
}

type recognizeRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

// Recognize AI 图片识别（返回菜名/食材/建议标签）
func (h *RecipeHandler) Recognize(c *gin.Context) {
	if h.aiProvider == nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "未配置 AI，无法识别图片")
		return
	}
	var req recognizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	rec, err := h.aiProvider.RecognizeImage(c.Request.Context(), absoluteURL(c, req.ImageURL), c.GetHeader("X-AI-Key"))
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "识别失败: "+err.Error())
		return
	}
	httpx.RespondOK(c, rec)
}

// absoluteURL 将相对路径（如 /uploads/xx.jpg）拼成完整 URL，便于 AI 服务拉取
func absoluteURL(c *gin.Context, u string) string {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + u
}
