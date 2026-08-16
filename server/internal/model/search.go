package model

// SearchResult 全局搜索结果（菜谱/餐厅/菜品分组）
type SearchResult struct {
	Recipes     []*Recipe     `json:"recipes"`
	Restaurants []*Restaurant `json:"restaurants"`
	Dishes      []*Dish       `json:"dishes"`
}
