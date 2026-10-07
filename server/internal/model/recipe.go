package model

import "time"

// Ingredient 食材
type Ingredient struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// Step 步骤
type Step struct {
	Order   int    `json:"order"`
	Content string `json:"content"`
	Image   string `json:"image,omitempty"`
}

// Recipe 自制菜谱
type Recipe struct {
	ID              string       `json:"id"`
	UserID          string       `json:"user_id"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Ingredients     []Ingredient `json:"ingredients"`
	Steps           []Step       `json:"steps"`
	CookTimeMinutes int          `json:"cook_time_minutes"`
	Difficulty      string       `json:"difficulty"`      // easy / medium / hard
	Rating          int          `json:"rating"`          // 1-5
	Tags            []string     `json:"tags"`            // 菜谱级标签 id（用户手选）
	IngredientTags  []string     `json:"ingredient_tags"` // 食材级标签 id（自动匹配派生）
	Images          []string     `json:"images"`
	MadeAt          *string      `json:"made_at"`         // 做过日期（YYYY-MM-DD，nil = 未做）
	IsLiked         bool         `json:"is_liked"`        // 喜欢（独立于收藏）
	IsFavorited     bool         `json:"is_favorited"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
