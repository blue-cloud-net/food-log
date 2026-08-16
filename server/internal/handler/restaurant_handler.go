package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"foodlog/server/internal/httpx"
	"foodlog/server/internal/model"
	"foodlog/server/internal/service"
)

// RestaurantHandler 餐厅处理器
type RestaurantHandler struct {
	restaurantService *service.RestaurantService
}

func NewRestaurantHandler(restaurantService *service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{restaurantService: restaurantService}
}

type restaurantRequest struct {
	Name        string   `json:"name" binding:"required,max=200"`
	Address     string   `json:"address"`
	CuisineType string   `json:"cuisine_type"`
	Description string   `json:"description"`
	AvgRating   float64  `json:"avg_rating"`
	Images      []string `json:"images"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
}

type dishRequest struct {
	Name        string   `json:"name" binding:"required,max=200"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Rating      int      `json:"rating"`
	Images      []string `json:"images"`
	EatenAt     *string  `json:"eaten_at"`
}

// List 餐厅列表
func (h *RestaurantHandler) List(c *gin.Context) {
	userID := httpx.GetUserID(c)
	q := parsePageQuery(c)

	list, err := h.restaurantService.List(c.Request.Context(), userID, q,
		c.Query("keyword"), c.Query("cuisine_type"), c.Query("sort"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, list)
}

// Create 创建餐厅
func (h *RestaurantHandler) Create(c *gin.Context) {
	var req restaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	rst := &model.Restaurant{
		Name:        req.Name,
		Address:     req.Address,
		CuisineType: req.CuisineType,
		Description: req.Description,
		AvgRating:   req.AvgRating,
		Images:      req.Images,
		Lat:         req.Lat,
		Lng:         req.Lng,
	}

	result, err := h.restaurantService.Create(c.Request.Context(), userID, rst)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// Get 餐厅详情（含菜品）
func (h *RestaurantHandler) Get(c *gin.Context) {
	userID := httpx.GetUserID(c)
	detail, err := h.restaurantService.GetDetail(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, detail)
}

// Update 更新餐厅
func (h *RestaurantHandler) Update(c *gin.Context) {
	var req restaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	rst := &model.Restaurant{
		Name:        req.Name,
		Address:     req.Address,
		CuisineType: req.CuisineType,
		Description: req.Description,
		AvgRating:   req.AvgRating,
		Images:      req.Images,
		Lat:         req.Lat,
		Lng:         req.Lng,
	}

	result, err := h.restaurantService.Update(c.Request.Context(), userID, c.Param("id"), rst)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// Delete 删除餐厅
func (h *RestaurantHandler) Delete(c *gin.Context) {
	userID := httpx.GetUserID(c)
	if err := h.restaurantService.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}

// AddDish 添加菜品
func (h *RestaurantHandler) AddDish(c *gin.Context) {
	var req dishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	dish := &model.Dish{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Rating:      req.Rating,
		Images:      req.Images,
		EatenAt:     req.EatenAt,
	}

	result, err := h.restaurantService.AddDish(c.Request.Context(), userID, c.Param("id"), dish)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// UpdateDish 更新菜品
func (h *RestaurantHandler) UpdateDish(c *gin.Context) {
	var req dishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}

	userID := httpx.GetUserID(c)
	dish := &model.Dish{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Rating:      req.Rating,
		Images:      req.Images,
		EatenAt:     req.EatenAt,
	}

	result, err := h.restaurantService.UpdateDish(c.Request.Context(), userID, c.Param("id"), dish)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// DeleteDish 删除菜品
func (h *RestaurantHandler) DeleteDish(c *gin.Context) {
	userID := httpx.GetUserID(c)
	if err := h.restaurantService.DeleteDish(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"deleted": true})
}
