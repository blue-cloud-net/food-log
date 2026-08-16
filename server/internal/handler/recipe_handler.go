package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/model"
	"foodlog/server/internal/service"
)

// RecipeHandler 菜谱处理器
type RecipeHandler struct {
	recipeService *service.RecipeService
}

func NewRecipeHandler(recipeService *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService}
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
		c.Query("keyword"), c.Query("difficulty"), c.Query("tag"), c.Query("sort"))
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
