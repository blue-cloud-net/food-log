package service

import (
	"context"
	"errors"

	"foodlog/server/internal/ai"
	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/tagging"
)

var ErrForbidden = errors.New("无权操作此资源")

// tagCompleter 把 ai.Provider 适配为 tagging.Completer
type tagCompleter struct{ p ai.Provider }

func (c tagCompleter) CompleteTags(ctx context.Context, prompt string) ([]string, error) {
	return c.p.CompleteTags(ctx, prompt, "")
}

// RecipeService 菜谱服务
type RecipeService struct {
	recipeRepo   *repository.RecipeRepo
	favoriteRepo *repository.FavoriteRepo
	aiProvider   ai.Provider // 可为 nil（禁用 AI 兜底）
}

func NewRecipeService(recipeRepo *repository.RecipeRepo, favoriteRepo *repository.FavoriteRepo, aiProvider ai.Provider) *RecipeService {
	return &RecipeService{recipeRepo: recipeRepo, favoriteRepo: favoriteRepo, aiProvider: aiProvider}
}

// enrichTags 根据菜谱内容自动生成标签并与手动标签合并
func (s *RecipeService) enrichTags(ctx context.Context, rec *model.Recipe) {
	names := make([]string, len(rec.Ingredients))
	for i, ing := range rec.Ingredients {
		names[i] = ing.Name
	}
	in := tagging.Input{
		Name:            rec.Name,
		Description:     rec.Description,
		Ingredients:     names,
		CookTimeMinutes: rec.CookTimeMinutes,
	}
	var completer tagging.Completer
	if s.aiProvider != nil {
		completer = tagCompleter{p: s.aiProvider}
	}
	rec.Tags = tagging.Generate(ctx, in, rec.Tags, completer)
}

// Create 创建菜谱
func (s *RecipeService) Create(ctx context.Context, userID string, rec *model.Recipe) (*model.Recipe, error) {
	rec.UserID = userID
	s.enrichTags(ctx, rec)
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
	s.fillFavorited(ctx, userID, []*model.Recipe{rec})
	return rec, nil
}

// fillFavorited 批量填充收藏状态
func (s *RecipeService) fillFavorited(ctx context.Context, userID string, list []*model.Recipe) {
	if s.favoriteRepo == nil || len(list) == 0 {
		return
	}
	ids, err := s.favoriteRepo.ListIDs(ctx, userID)
	if err != nil {
		return
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	for _, r := range list {
		r.IsFavorited = set[r.ID]
	}
}

// List 菜谱列表
func (s *RecipeService) List(ctx context.Context, userID string, q *model.PageQuery, keyword, difficulty, tag, sort string, isFavorite bool) (*model.Paginated, error) {
	list, total, err := s.recipeRepo.List(ctx, userID, q, keyword, difficulty, tag, sort, isFavorite)
	if err != nil {
		return nil, err
	}
	if isFavorite {
		for _, r := range list {
			r.IsFavorited = true
		}
	} else {
		s.fillFavorited(ctx, userID, list)
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
	s.enrichTags(ctx, rec)
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

// Favorite 收藏菜谱（校验归属）
func (s *RecipeService) Favorite(ctx context.Context, userID, id string) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.favoriteRepo.Add(ctx, userID, id)
}

// Unfavorite 取消收藏
func (s *RecipeService) Unfavorite(ctx context.Context, userID, id string) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.favoriteRepo.Remove(ctx, userID, id)
}

// Random 随机获取一条菜谱（可按标签/难度过滤）
func (s *RecipeService) Random(ctx context.Context, userID, tag, difficulty string) (*model.Recipe, error) {
	rec, err := s.recipeRepo.Random(ctx, userID, tag, difficulty)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	s.fillFavorited(ctx, userID, []*model.Recipe{rec})
	return rec, nil
}
