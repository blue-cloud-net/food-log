package model

import "time"

// User 用户
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserStats 用户统计
type UserStats struct {
	RecipeCount     int `json:"recipe_count"`
	RestaurantCount int `json:"restaurant_count"`
	DishCount       int `json:"dish_count"`
}

// UserProfile 返回给前端的用户信息（含统计）
type UserProfile struct {
	User
	UserStats
}
