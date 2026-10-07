package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/service"
)

// StorageLocationHandler 存放位置字典处理器（词表查询 + 用户自定义位置维护）
type StorageLocationHandler struct {
	locationService *service.StorageLocationService
}

func NewStorageLocationHandler(locationService *service.StorageLocationService) *StorageLocationHandler {
	return &StorageLocationHandler{locationService: locationService}
}

// List 存放位置词表（全局预设 + 本人自定义），供前端分组渲染与名称解析使用
func (h *StorageLocationHandler) List(c *gin.Context) {
	list, err := h.locationService.List(c.Request.Context(), httpx.GetUserID(c))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, list)
}

type storageLocationRequest struct {
	Area      string `json:"area" binding:"required,oneof=fridge outside"`
	Name      string `json:"name" binding:"required,max=50"`
	SortOrder int    `json:"sort_order"`
}

// Create 新建自定义存放位置
func (h *StorageLocationHandler) Create(c *gin.Context) {
	var req storageLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	loc, err := h.locationService.Create(c.Request.Context(), httpx.GetUserID(c), req.Area, req.Name)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, loc)
}

// Update 重命名 / 调整大类（仅本人自定义位置）
func (h *StorageLocationHandler) Update(c *gin.Context) {
	var req storageLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	loc, err := h.locationService.Update(c.Request.Context(), httpx.GetUserID(c), c.Param("id"),
		req.Area, req.Name, req.SortOrder)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, loc)
}

// Delete 删除自定义存放位置（位置下仍有库存食材时返回 409）
func (h *StorageLocationHandler) Delete(c *gin.Context) {
	if err := h.locationService.Delete(c.Request.Context(), httpx.GetUserID(c), c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}
