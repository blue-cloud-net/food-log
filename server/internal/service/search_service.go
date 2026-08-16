package service

import (
	"context"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// SearchService 全局搜索服务（菜谱/餐厅/菜品）
type SearchService struct {
	recipeRepo     *repository.RecipeRepo
	restaurantRepo *repository.RestaurantRepo
	dishRepo       *repository.DishRepo
}

func NewSearchService(recipeRepo *repository.RecipeRepo, restaurantRepo *repository.RestaurantRepo, dishRepo *repository.DishRepo) *SearchService {
	return &SearchService{recipeRepo: recipeRepo, restaurantRepo: restaurantRepo, dishRepo: dishRepo}
}

// Search 聚合搜索，各类返回前 10 条；单类查询失败不影响其它类
func (s *SearchService) Search(ctx context.Context, userID, keyword string) *model.SearchResult {
	result := &model.SearchResult{
		Recipes:     []*model.Recipe{},
		Restaurants: []*model.Restaurant{},
		Dishes:      []*model.Dish{},
	}
	const limit = 10

	if recipes, err := s.recipeRepo.SearchByName(ctx, userID, keyword, limit); err == nil {
		result.Recipes = recipes
	}
	if restaurants, err := s.restaurantRepo.SearchByName(ctx, userID, keyword, limit); err == nil {
		result.Restaurants = restaurants
	}
	if dishes, err := s.dishRepo.SearchByName(ctx, userID, keyword, limit); err == nil {
		result.Dishes = dishes
	}
	return result
}
