package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
)

// TagHandler 标签字典处理器（词表查询 + 用户自定义分类/标签维护）
type TagHandler struct {
	tagService *service.TagService
}

func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

// List 标签词表（全局预设 + 本人自定义），供前端选择器与 key→名称解析使用
func (h *TagHandler) List(c *gin.Context) {
	cats, err := h.tagService.Categories(c.Request.Context(), httpx.GetUserID(c))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, cats)
}

type tagCategoryRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// CreateCategory 新建自定义分类
func (h *TagHandler) CreateCategory(c *gin.Context) {
	var req tagCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	cat, err := h.tagService.CreateCategory(c.Request.Context(), httpx.GetUserID(c), req.Name, req.Color)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, cat)
}

// UpdateCategory 重命名/改色/改排序（仅本人）
func (h *TagHandler) UpdateCategory(c *gin.Context) {
	var req tagCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	cat, err := h.tagService.UpdateCategory(c.Request.Context(), httpx.GetUserID(c), c.Param("id"), req.Name, req.Color, req.SortOrder)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, cat)
}

// DeleteCategory 删除分类（级联删除其下标签与菜谱关联）
func (h *TagHandler) DeleteCategory(c *gin.Context) {
	if err := h.tagService.DeleteCategory(c.Request.Context(), httpx.GetUserID(c), c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}

type tagRequest struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name" binding:"required,max=50"`
	MutexGroup string `json:"mutex_group"`
	SortOrder  int    `json:"sort_order"`
}

// CreateTag 新建自定义标签；category_id 为空时自动归入「自定义」分类
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := httpx.GetUserID(c)

	categoryID := req.CategoryID
	if categoryID == "" {
		cat, err := h.tagService.EnsureCustomCategory(c.Request.Context(), userID)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		categoryID = cat.ID
	}

	tag, err := h.tagService.CreateTag(c.Request.Context(), userID, categoryID, req.Name, req.MutexGroup)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, tag)
}

// UpdateTag 重命名/改互斥组/改排序（仅本人）
func (h *TagHandler) UpdateTag(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	tag, err := h.tagService.UpdateTag(c.Request.Context(), httpx.GetUserID(c), c.Param("id"), req.Name, req.MutexGroup, req.SortOrder)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, tag)
}

// DeleteTag 删除标签（级联解除菜谱关联）
func (h *TagHandler) DeleteTag(c *gin.Context) {
	if err := h.tagService.DeleteTag(c.Request.Context(), httpx.GetUserID(c), c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}
