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
	Name            string   `json:"name" binding:"required,max=200"`
	Address         string   `json:"address"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
	RecommendRating int      `json:"recommend_rating" binding:"omitempty,min=1,max=5"`
	ValueRating     int      `json:"value_rating" binding:"omitempty,min=1,max=5"`
	AmbienceRating  int      `json:"ambience_rating" binding:"omitempty,min=1,max=5"`
	ServiceRating   int      `json:"service_rating" binding:"omitempty,min=1,max=5"`
	Images          []string `json:"images"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
	IsVisited       bool     `json:"is_visited"`
}

type dishRequest struct {
	Name        string   `json:"name" binding:"required,max=200"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Rating      int      `json:"rating" binding:"omitempty,min=1,max=5"`
	Tags        []string `json:"tags"`
	Images      []string `json:"images"`
	EatenAt     *string  `json:"eaten_at"`
	IsLiked     bool     `json:"is_liked"`
}

// List 餐厅列表（keyword 搜店名/地址，tag 按餐厅标签 id 筛选，sort 见 repository.restaurantSortCols）
// visited=true|false 用于「已探店/未探店」
func (h *RestaurantHandler) List(c *gin.Context) {
	userID := httpx.GetUserID(c)
	q := parsePageQuery(c)

	list, err := h.restaurantService.List(c.Request.Context(), userID, q,
		c.Query("keyword"), c.Query("tag"), c.Query("sort"), c.Query("visited"))
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
		Name:            req.Name,
		Address:         req.Address,
		Description:     req.Description,
		Tags:            req.Tags,
		RecommendRating: req.RecommendRating,
		ValueRating:     req.ValueRating,
		AmbienceRating:  req.AmbienceRating,
		ServiceRating:   req.ServiceRating,
		Images:          req.Images,
		Lat:             req.Lat,
		Lng:             req.Lng,
		IsVisited:       req.IsVisited,
	}

	result, err := h.restaurantService.Create(c.Request.Context(), userID, rst)
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, result)
}

// Get 餐厅详情（含菜品，dish_tag 可按菜品标签 id 过滤菜品列表）
func (h *RestaurantHandler) Get(c *gin.Context) {
	userID := httpx.GetUserID(c)
	detail, err := h.restaurantService.GetDetail(c.Request.Context(), userID, c.Param("id"), c.Query("dish_tag"))
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
		Name:            req.Name,
		Address:         req.Address,
		Description:     req.Description,
		Tags:            req.Tags,
		RecommendRating: req.RecommendRating,
		ValueRating:     req.ValueRating,
		AmbienceRating:  req.AmbienceRating,
		ServiceRating:   req.ServiceRating,
		Images:          req.Images,
		Lat:             req.Lat,
		Lng:             req.Lng,
		IsVisited:       req.IsVisited,
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
		Tags:        req.Tags,
		Images:      req.Images,
		EatenAt:     req.EatenAt,
		IsLiked:     req.IsLiked,
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
		Tags:        req.Tags,
		Images:      req.Images,
		EatenAt:     req.EatenAt,
		IsLiked:     req.IsLiked,
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

type visitedRequest struct {
	Visited bool `json:"visited"`
}

// SetVisited 标记「已探店 / 未探店」
func (h *RestaurantHandler) SetVisited(c *gin.Context) {
	var req visitedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := httpx.GetUserID(c)
	if err := h.restaurantService.SetVisited(c.Request.Context(), userID, c.Param("id"), req.Visited); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"is_visited": req.Visited})
}

// ListDishes 菜品列表（liked=true 只看喜欢，keyword/tag/restaurant_id 可选）
func (h *RestaurantHandler) ListDishes(c *gin.Context) {
	userID := httpx.GetUserID(c)
	q := parsePageQuery(c)

	list, err := h.restaurantService.ListDishes(c.Request.Context(), userID, q,
		c.Query("liked") == "true", c.Query("keyword"), c.Query("tag"), c.Query("restaurant_id"))
	if err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, list)
}

// SetDishLiked 标记菜品「喜欢」
func (h *RestaurantHandler) SetDishLiked(c *gin.Context) {
	var req likedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, httpx.CodeBadRequest, "参数错误: "+err.Error())
		return
	}
	userID := httpx.GetUserID(c)
	if err := h.restaurantService.SetDishLiked(c.Request.Context(), userID, c.Param("id"), req.Liked); err != nil {
		httpx.RespondErrorWithErr(c, err)
		return
	}
	httpx.RespondOK(c, gin.H{"is_liked": req.Liked})
}
