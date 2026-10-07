package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/tagging"
)

var ErrForbidden = errors.New("无权操作此资源")

// RecipeService 菜谱服务
type RecipeService struct {
	pool         *pgxpool.Pool
	recipeRepo   *repository.RecipeRepo
	favoriteRepo *repository.FavoriteRepo
	tagRepo      *repository.TagRepo
	tagService   *TagService
}

func NewRecipeService(
	pool *pgxpool.Pool,
	recipeRepo *repository.RecipeRepo,
	favoriteRepo *repository.FavoriteRepo,
	tagRepo *repository.TagRepo,
	tagService *TagService,
) *RecipeService {
	return &RecipeService{
		pool:         pool,
		recipeRepo:   recipeRepo,
		favoriteRepo: favoriteRepo,
		tagRepo:      tagRepo,
		tagService:   tagService,
	}
}

// prepareTags 校验手动标签，并依据规则推导食材级标签（自动部分剔除与手选重复的项）
func (s *RecipeService) prepareTags(ctx context.Context, userID string, rec *model.Recipe) error {
	manual, err := s.tagService.ResolveManualTags(ctx, userID, rec.Tags)
	if err != nil {
		return err
	}
	rec.Tags = manual
	rec.IngredientTags = s.tagService.ComputeIngredientTags(ctx, userID, taggingInput(rec), manual)
	return nil
}

func taggingInput(rec *model.Recipe) tagging.Input {
	names := make([]string, len(rec.Ingredients))
	for i, ing := range rec.Ingredients {
		names[i] = ing.Name
	}
	return tagging.Input{
		Name:            rec.Name,
		Description:     rec.Description,
		Ingredients:     names,
		CookTimeMinutes: rec.CookTimeMinutes,
	}
}

// fillTags 批量填充菜谱的菜谱级与食材级标签 id
func (s *RecipeService) fillTags(ctx context.Context, list []*model.Recipe) {
	if len(list) == 0 {
		return
	}
	ids := make([]string, len(list))
	for i, r := range list {
		ids[i] = r.ID
	}
	manual, ingredient, err := s.tagRepo.TagsForRecipes(ctx, ids)
	if err != nil {
		return
	}
	for _, r := range list {
		if v, ok := manual[r.ID]; ok {
			r.Tags = v
		} else {
			r.Tags = []string{}
		}
		if v, ok := ingredient[r.ID]; ok {
			r.IngredientTags = v
		} else {
			r.IngredientTags = []string{}
		}
	}
}

// Create 创建菜谱（菜谱行 + 两组标签关联在同一事务内写入）
func (s *RecipeService) Create(ctx context.Context, userID string, rec *model.Recipe) (*model.Recipe, error) {
	rec.UserID = userID
	if err := s.prepareTags(ctx, userID, rec); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.recipeRepo.CreateTx(ctx, tx, rec); err != nil {
		return nil, err
	}
	if err := s.tagRepo.ReplaceRecipeTagsTx(ctx, tx, rec.ID, rec.Tags); err != nil {
		return nil, err
	}
	if err := s.tagRepo.ReplaceRecipeIngredientTagsTx(ctx, tx, rec.ID, rec.IngredientTags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
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
	s.fillTags(ctx, []*model.Recipe{rec})
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
// tagID 过滤菜谱级标签，ingredientTagID 过滤食材级标签
func (s *RecipeService) List(ctx context.Context, userID string, q *model.PageQuery, keyword, difficulty, tagID, ingredientTagID, sort string, isFavorite bool) (*model.Paginated, error) {
	list, total, err := s.recipeRepo.List(ctx, userID, q, keyword, difficulty, tagID, ingredientTagID, sort, isFavorite)
	if err != nil {
		return nil, err
	}
	s.fillTags(ctx, list)
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

// Update 更新菜谱（校验归属，菜谱行 + 两组标签关联在同一事务内写入）
func (s *RecipeService) Update(ctx context.Context, userID, id string, rec *model.Recipe) (*model.Recipe, error) {
	existing, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	rec.ID = existing.ID
	rec.UserID = existing.UserID
	if err := s.prepareTags(ctx, userID, rec); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.recipeRepo.UpdateTx(ctx, tx, rec); err != nil {
		return nil, err
	}
	if err := s.tagRepo.ReplaceRecipeTagsTx(ctx, tx, rec.ID, rec.Tags); err != nil {
		return nil, err
	}
	if err := s.tagRepo.ReplaceRecipeIngredientTagsTx(ctx, tx, rec.ID, rec.IngredientTags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
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
func (s *RecipeService) Random(ctx context.Context, userID, tagID, ingredientTagID, difficulty string) (*model.Recipe, error) {
	rec, err := s.recipeRepo.Random(ctx, userID, tagID, ingredientTagID, difficulty)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	s.fillTags(ctx, []*model.Recipe{rec})
	s.fillFavorited(ctx, userID, []*model.Recipe{rec})
	return rec, nil
}
