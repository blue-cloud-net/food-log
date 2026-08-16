package service

import (
	"context"
	"errors"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

var ErrNotFound = errors.New("资源不存在")

// RestaurantService 餐厅服务
type RestaurantService struct {
	restaurantRepo *repository.RestaurantRepo
	dishRepo       *repository.DishRepo
}

func NewRestaurantService(restaurantRepo *repository.RestaurantRepo, dishRepo *repository.DishRepo) *RestaurantService {
	return &RestaurantService{restaurantRepo: restaurantRepo, dishRepo: dishRepo}
}

// Create 创建餐厅
func (s *RestaurantService) Create(ctx context.Context, userID string, rst *model.Restaurant) (*model.Restaurant, error) {
	rst.UserID = userID
	if err := s.restaurantRepo.Create(ctx, rst); err != nil {
		return nil, err
	}
	return rst, nil
}

// Get 获取餐厅（校验归属）
func (s *RestaurantService) Get(ctx context.Context, userID, id string) (*model.Restaurant, error) {
	rst, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rst == nil {
		return nil, ErrNotFound
	}
	if rst.UserID != userID {
		return nil, ErrForbidden
	}
	return rst, nil
}

// GetDetail 获取餐厅详情（含菜品列表）
func (s *RestaurantService) GetDetail(ctx context.Context, userID, id string) (*model.RestaurantDetail, error) {
	rst, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	dishes, err := s.dishRepo.ListByRestaurant(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]model.Dish, len(dishes))
	for i, d := range dishes {
		items[i] = *d
	}
	return &model.RestaurantDetail{Restaurant: *rst, Dishes: items}, nil
}

// List 餐厅列表
func (s *RestaurantService) List(ctx context.Context, userID string, q *model.PageQuery, keyword, cuisineType, sort string) (*model.Paginated, error) {
	list, total, err := s.restaurantRepo.List(ctx, userID, q, keyword, cuisineType, sort)
	if err != nil {
		return nil, err
	}
	items := make([]any, len(list))
	for i, r := range list {
		items[i] = r
	}
	return &model.Paginated{List: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Update 更新餐厅（校验归属）
func (s *RestaurantService) Update(ctx context.Context, userID, id string, rst *model.Restaurant) (*model.Restaurant, error) {
	existing, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	rst.ID = existing.ID
	rst.UserID = existing.UserID
	if err := s.restaurantRepo.Update(ctx, rst); err != nil {
		return nil, err
	}
	return rst, nil
}

// Delete 删除餐厅（级联删除菜品）
func (s *RestaurantService) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Get(ctx, userID, id)
	if err != nil {
		return err
	}
	return s.restaurantRepo.Delete(ctx, id)
}

// AddDish 添加菜品（校验餐厅归属）
func (s *RestaurantService) AddDish(ctx context.Context, userID, restaurantID string, d *model.Dish) (*model.Dish, error) {
	if _, err := s.Get(ctx, userID, restaurantID); err != nil {
		return nil, err
	}
	d.RestaurantID = restaurantID
	d.UserID = userID
	if err := s.dishRepo.Create(ctx, d); err != nil {
		return nil, err
	}
	if err := s.restaurantRepo.UpdateAvgRating(ctx, restaurantID); err != nil {
		return nil, err
	}
	return d, nil
}

// UpdateDish 更新菜品（校验归属）
func (s *RestaurantService) UpdateDish(ctx context.Context, userID, dishID string, d *model.Dish) (*model.Dish, error) {
	existing, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.UserID != userID {
		return nil, ErrForbidden
	}
	d.ID = existing.ID
	d.RestaurantID = existing.RestaurantID
	d.UserID = existing.UserID
	if err := s.dishRepo.Update(ctx, d); err != nil {
		return nil, err
	}
	if err := s.restaurantRepo.UpdateAvgRating(ctx, d.RestaurantID); err != nil {
		return nil, err
	}
	return d, nil
}

// DeleteDish 删除菜品（校验归属）
func (s *RestaurantService) DeleteDish(ctx context.Context, userID, dishID string) error {
	existing, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if existing.UserID != userID {
		return ErrForbidden
	}
	if err := s.dishRepo.Delete(ctx, dishID); err != nil {
		return err
	}
	return s.restaurantRepo.UpdateAvgRating(ctx, existing.RestaurantID)
}
