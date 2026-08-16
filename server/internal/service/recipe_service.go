package service

import (
	"context"
	"errors"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

var ErrForbidden = errors.New("无权操作此资源")

// RecipeService 菜谱服务
type RecipeService struct {
	recipeRepo *repository.RecipeRepo
}

func NewRecipeService(recipeRepo *repository.RecipeRepo) *RecipeService {
	return &RecipeService{recipeRepo: recipeRepo}
}

// Create 创建菜谱
func (s *RecipeService) Create(ctx context.Context, userID string, rec *model.Recipe) (*model.Recipe, error) {
	rec.UserID = userID
	if err := s.recipeRepo.Create(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// Get 获取菜谱（校验归属）
func (s *RecipeService) Get(ctx context.Context, userID, id string) (*model.Recipe, error) {
	rec, err := s.recipeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	if rec.UserID != userID {
		return nil, ErrForbidden
	}
	return rec, nil
}

// List 菜谱列表
func (s *RecipeService) List(ctx context.Context, userID string, q *model.PageQuery, keyword, difficulty, tag, sort string) (*model.Paginated, error) {
	list, total, err := s.recipeRepo.List(ctx, userID, q, keyword, difficulty, tag, sort)
	if err != nil {
		return nil, err
	}
	items := make([]any, len(list))
	for i, r := range list {
		items[i] = r
	}
	return &model.Paginated{List: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Update 更新菜谱（校验归属）
func (s *RecipeService) Update(ctx context.Context, userID, id string, rec *model.Recipe) (*model.Recipe, error) {
	existing, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	rec.ID = existing.ID
	rec.UserID = existing.UserID
	if err := s.recipeRepo.Update(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// Delete 删除菜谱（校验归属）
func (s *RecipeService) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Get(ctx, userID, id)
	if err != nil {
		return err
	}
	return s.recipeRepo.Delete(ctx, id)
}
