package model

import "time"

// ExportData 数据导出/导入格式
type ExportData struct {
	Version     string        `json:"version"`
	ExportedAt  time.Time     `json:"exported_at"`
	Recipes     []*Recipe     `json:"recipes"`
	Restaurants []*Restaurant `json:"restaurants"`
	Dishes      []*Dish       `json:"dishes"`
	Favorites   []string      `json:"favorites"` // recipe_id 列表
}
