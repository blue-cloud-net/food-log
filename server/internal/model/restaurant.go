package model

import "time"

// Restaurant 餐厅
type Restaurant struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Description     string    `json:"description"`
	Tags            []string  `json:"tags"`             // 餐厅标签 id 列表（用户手选）
	RecommendRating int       `json:"recommend_rating"` // 1-5 推荐度
	ValueRating     int       `json:"value_rating"`     // 1-5 性价比
	AmbienceRating  int       `json:"ambience_rating"`  // 1-5 环境
	ServiceRating   int       `json:"service_rating"`   // 1-5 服务
	Images          []string  `json:"images"`
	Lat             *float64  `json:"lat"`
	Lng             *float64  `json:"lng"`
	IsVisited       bool      `json:"is_visited"` // 是否已探店（false = 未探店）
	DishCount       int       `json:"dish_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Dish 店内菜品
type Dish struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        *float64  `json:"price"`
	Rating       int       `json:"rating"` // 1-5（前端展示为「推荐度」）
	Tags         []string  `json:"tags"`   // 菜品标签 id 列表（用户手选）
	Images       []string  `json:"images"`
	EatenAt      *string   `json:"eaten_at"`
	IsLiked      bool      `json:"is_liked"`                    // 喜欢该菜品
	RestaurantName string  `json:"restaurant_name,omitempty"`   // 列表场景下所属餐厅名
	CreatedAt    time.Time `json:"created_at"`
}

// RestaurantDetail 餐厅详情（含菜品列表）
type RestaurantDetail struct {
	Restaurant
	Dishes []Dish `json:"dishes"`
}
