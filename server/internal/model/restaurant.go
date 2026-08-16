package model

import "time"

// Restaurant 餐厅
type Restaurant struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	CuisineType string    `json:"cuisine_type"`
	Description string    `json:"description"`
	AvgRating   float64   `json:"avg_rating"`
	Images      []string  `json:"images"`
	Lat         *float64  `json:"lat"`
	Lng         *float64  `json:"lng"`
	DishCount   int       `json:"dish_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Dish 店内菜品
type Dish struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        *float64  `json:"price"`
	Rating       int       `json:"rating"` // 1-5
	Images       []string  `json:"images"`
	EatenAt      *string   `json:"eaten_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// RestaurantDetail 餐厅详情（含菜品列表）
type RestaurantDetail struct {
	Restaurant
	Dishes []Dish `json:"dishes"`
}
