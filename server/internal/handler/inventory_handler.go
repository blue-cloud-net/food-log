package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/service"
)

// InventoryHandler 库存食材处理器
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// List 库存列表（关键词 / 位置 / 大类 / 临期筛选 + 排序，不分页）
func (h *InventoryHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("expiring_within_days"))
	filter := repository.InventoryFilter{
		Keyword:            c.Query("keyword"),
		LocationID:         c.Query("location_id"),
		Area:               c.Query("area"),
		ExpiringWithinDays: days,
		Sort:               c.Query("sort"),
	}
	list, err := h.inventoryService.List(c.Request.Context(), httpx.GetUserID(c), filter)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"list": list, "total": len(list)})
}

// Get 库存条目详情
func (h *InventoryHandler) Get(c *gin.Context) {
	it, err := h.inventoryService.Get(c.Request.Context(), httpx.GetUserID(c), c.Param("id"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, it)
}

type inventoryItemRequest struct {
	LocationID string   `json:"location_id" binding:"required"`
	Name       string   `json:"name" binding:"required,max=100"`
	Amount     string   `json:"amount"`
	Unit       string   `json:"unit"`
	Category   string   `json:"category"`
	ExpireAt   string   `json:"expire_at"`
	Note       string   `json:"note"`
	Images     []string `json:"images"`
}

// toModel 把请求体转换为数据模型
func (r *inventoryItemRequest) toModel() *model.InventoryItem {
	it := &model.InventoryItem{
		LocationID: r.LocationID,
		Name:       r.Name,
		Amount:     r.Amount,
		Unit:       r.Unit,
		Category:   r.Category,
		Note:       r.Note,
		Images:     r.Images,
	}
	if r.ExpireAt != "" {
		expireAt := r.ExpireAt
		it.ExpireAt = &expireAt
	}
	return it
}

// Create 新增库存食材
func (h *InventoryHandler) Create(c *gin.Context) {
	var req inventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	it, err := h.inventoryService.Create(c.Request.Context(), httpx.GetUserID(c), req.toModel())
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, it)
}

// Update 更新库存食材
func (h *InventoryHandler) Update(c *gin.Context) {
	var req inventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	it, err := h.inventoryService.Update(c.Request.Context(), httpx.GetUserID(c), c.Param("id"), req.toModel())
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, it)
}

// Delete 删除库存食材（如「已用完」）
func (h *InventoryHandler) Delete(c *gin.Context) {
	if err := h.inventoryService.Delete(c.Request.Context(), httpx.GetUserID(c), c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}
