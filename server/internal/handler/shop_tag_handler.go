package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/service"
)

// ShopTagHandler 探店标签字典处理器。
// 餐厅与菜品各有一套独立字典，这里用「域工厂」返回 gin.HandlerFunc，
// 避免为两个域各写一遍完全相同的七个方法。
type ShopTagHandler struct {
	shopTagService *service.ShopTagService
}

func NewShopTagHandler(shopTagService *service.ShopTagService) *ShopTagHandler {
	return &ShopTagHandler{shopTagService: shopTagService}
}

type shopTagCategoryRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

type shopTagRequest struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name" binding:"required,max=50"`
	MutexGroup string `json:"mutex_group"`
	SortOrder  int    `json:"sort_order"`
}

// List 标签词表（全局预设 + 本人自定义），供前端选择器与 id→名称解析使用
func (h *ShopTagHandler) List(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		cats, err := h.shopTagService.Categories(c.Request.Context(), httpx.GetUserID(c), domain)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, cats)
	}
}

// CreateCategory 新建自定义分类
func (h *ShopTagHandler) CreateCategory(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req shopTagCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
			return
		}
		cat, err := h.shopTagService.CreateCategory(c.Request.Context(), httpx.GetUserID(c), domain, req.Name, req.Color)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, cat)
	}
}

// UpdateCategory 重命名/改色/改排序（仅本人）
func (h *ShopTagHandler) UpdateCategory(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req shopTagCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
			return
		}
		cat, err := h.shopTagService.UpdateCategory(c.Request.Context(), httpx.GetUserID(c), domain,
			c.Param("id"), req.Name, req.Color, req.SortOrder)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, cat)
	}
}

// DeleteCategory 删除分类（级联删除其下标签与实体关联）
func (h *ShopTagHandler) DeleteCategory(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.shopTagService.DeleteCategory(c.Request.Context(), httpx.GetUserID(c), domain, c.Param("id")); err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, gin.H{"deleted": true})
	}
}

// CreateTag 新建自定义标签；category_id 为空时自动归入「自定义」分类
func (h *ShopTagHandler) CreateTag(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req shopTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
			return
		}
		userID := httpx.GetUserID(c)
		ctx := c.Request.Context()

		categoryID := req.CategoryID
		if categoryID == "" {
			cat, err := h.shopTagService.EnsureCustomCategory(ctx, userID, domain)
			if err != nil {
				httpx.RespondErrorWithErr(c, err)
				return
			}
			categoryID = cat.ID
		}

		tag, err := h.shopTagService.CreateTag(ctx, userID, domain, categoryID, req.Name, req.MutexGroup)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, tag)
	}
}

// UpdateTag 重命名/改互斥组/改排序（仅本人）
func (h *ShopTagHandler) UpdateTag(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req shopTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
			return
		}
		tag, err := h.shopTagService.UpdateTag(c.Request.Context(), httpx.GetUserID(c), domain,
			c.Param("id"), req.Name, req.MutexGroup, req.SortOrder)
		if err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, tag)
	}
}

// DeleteTag 删除标签（级联解除实体关联）
func (h *ShopTagHandler) DeleteTag(domain repository.ShopTagDomain) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.shopTagService.DeleteTag(c.Request.Context(), httpx.GetUserID(c), domain, c.Param("id")); err != nil {
			httpx.RespondErrorWithErr(c, err)
			return
		}
		httpx.RespondOK(c, gin.H{"deleted": true})
	}
}
