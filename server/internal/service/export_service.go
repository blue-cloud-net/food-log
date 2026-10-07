package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// ExportService 数据导出/导入服务
type ExportService struct {
	pool           *pgxpool.Pool
	recipeRepo     *repository.RecipeRepo
	restaurantRepo *repository.RestaurantRepo
	dishRepo       *repository.DishRepo
	favoriteRepo   *repository.FavoriteRepo
	tagRepo        *repository.TagRepo
	tagService     *TagService
	shopTagRepo    *repository.ShopTagRepo
	shopTagService *ShopTagService
}

func NewExportService(pool *pgxpool.Pool, recipeRepo *repository.RecipeRepo, restaurantRepo *repository.RestaurantRepo, dishRepo *repository.DishRepo, favoriteRepo *repository.FavoriteRepo, tagRepo *repository.TagRepo, tagService *TagService, shopTagRepo *repository.ShopTagRepo, shopTagService *ShopTagService) *ExportService {
	return &ExportService{
		pool:           pool,
		recipeRepo:     recipeRepo,
		restaurantRepo: restaurantRepo,
		dishRepo:       dishRepo,
		favoriteRepo:   favoriteRepo,
		tagRepo:        tagRepo,
		tagService:     tagService,
		shopTagRepo:    shopTagRepo,
		shopTagService: shopTagService,
	}
}

// ExportData 组装完整导出数据（JSON）
// 注意：导出文件里菜谱的 tags / ingredient_tags 以及餐厅、菜品的 tags 均为标签
// 「名称」而非 id，这样备份可跨数据库导入（id 在不同库间不通用）。
func (s *ExportService) ExportData(ctx context.Context, userID string) (*model.ExportData, error) {
	recipes, err := s.recipeRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.attachTagNames(ctx, recipes); err != nil {
		return nil, err
	}
	restaurants, err := s.restaurantRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.attachRestaurantTagNames(ctx, restaurants); err != nil {
		return nil, err
	}
	dishes, err := s.dishRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.attachDishTagNames(ctx, dishes); err != nil {
		return nil, err
	}
	favorites, err := s.favoriteRepo.ListIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &model.ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now(),
		Recipes:     recipes,
		Restaurants: restaurants,
		Dishes:      dishes,
		Favorites:   favorites,
	}, nil
}

// attachTagNames 把菜谱的标签 id 替换为标签名称（导出场景）
func (s *ExportService) attachTagNames(ctx context.Context, recipes []*model.Recipe) error {
	if len(recipes) == 0 {
		return nil
	}
	ids := make([]string, len(recipes))
	for i, r := range recipes {
		ids[i] = r.ID
	}
	manual, ingredient, err := s.tagRepo.TagsForRecipes(ctx, ids)
	if err != nil {
		return err
	}

	allIDs := make([]string, 0, len(manual)+len(ingredient))
	seen := map[string]bool{}
	collect := func(group map[string][]string) {
		for _, tagIDs := range group {
			for _, id := range tagIDs {
				if !seen[id] {
					seen[id] = true
					allIDs = append(allIDs, id)
				}
			}
		}
	}
	collect(manual)
	collect(ingredient)
	nameOf, err := s.tagRepo.NamesByIDs(ctx, allIDs)
	if err != nil {
		return err
	}

	toNames := func(tagIDs []string) []string {
		out := make([]string, 0, len(tagIDs))
		for _, id := range tagIDs {
			if name, ok := nameOf[id]; ok {
				out = append(out, name)
			}
		}
		return out
	}
	for _, r := range recipes {
		r.Tags = toNames(manual[r.ID])
		r.IngredientTags = toNames(ingredient[r.ID])
	}
	return nil
}

// attachRestaurantTagNames 把餐厅标签 id 替换为标签名称（导出场景）
func (s *ExportService) attachRestaurantTagNames(ctx context.Context, list []*model.Restaurant) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]string, len(list))
	for i, rst := range list {
		ids[i] = rst.ID
	}
	byID, err := s.shopTagNameMap(ctx, repository.ShopTagDomainRestaurant, ids)
	if err != nil {
		return err
	}
	for _, rst := range list {
		if names, ok := byID[rst.ID]; ok {
			rst.Tags = names
		} else {
			rst.Tags = []string{}
		}
	}
	return nil
}

// attachDishTagNames 把菜品标签 id 替换为标签名称（导出场景）
func (s *ExportService) attachDishTagNames(ctx context.Context, list []*model.Dish) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]string, len(list))
	for i, d := range list {
		ids[i] = d.ID
	}
	byID, err := s.shopTagNameMap(ctx, repository.ShopTagDomainDish, ids)
	if err != nil {
		return err
	}
	for _, d := range list {
		if names, ok := byID[d.ID]; ok {
			d.Tags = names
		} else {
			d.Tags = []string{}
		}
	}
	return nil
}

// shopTagNameMap 返回 实体 id → 标签名称列表，仅包含有关联标签的实体
func (s *ExportService) shopTagNameMap(ctx context.Context, domain repository.ShopTagDomain, entityIDs []string) (map[string][]string, error) {
	byID, err := s.shopTagRepo.TagsFor(ctx, domain, entityIDs)
	if err != nil {
		return nil, err
	}

	allIDs := []string{}
	seen := map[string]bool{}
	for _, tagIDs := range byID {
		for _, id := range tagIDs {
			if !seen[id] {
				seen[id] = true
				allIDs = append(allIDs, id)
			}
		}
	}
	nameOf, err := s.shopTagRepo.NamesByIDs(ctx, domain, allIDs)
	if err != nil {
		return nil, err
	}

	out := map[string][]string{}
	for entityID, tagIDs := range byID {
		names := make([]string, 0, len(tagIDs))
		for _, id := range tagIDs {
			if name, ok := nameOf[id]; ok {
				names = append(names, name)
			}
		}
		out[entityID] = names
	}
	return out, nil
}

// ExportCSV 导出菜谱 CSV 内容
func (s *ExportService) ExportCSV(ctx context.Context, userID string) ([]byte, error) {
	recipes, err := s.recipeRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.attachTagNames(ctx, recipes); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"菜名", "描述", "食材", "步骤", "耗时(分钟)", "难度", "评分", "菜谱标签", "食材标签", "创建时间"}); err != nil {
		return nil, err
	}
	for _, r := range recipes {
		if err := w.Write([]string{
			r.Name,
			r.Description,
			joinIngredients(r.Ingredients),
			joinSteps(r.Steps),
			itov(r.CookTimeMinutes),
			r.Difficulty,
			itov(r.Rating),
			strings.Join(r.Tags, "/"),
			strings.Join(r.IngredientTags, "/"),
			r.CreatedAt.Format("2006-01-02 15:04"),
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// ImportData 导入备份数据。
// mode: append（默认，同名菜谱跳过） | overwrite（先清空用户数据再导入）
func (s *ExportService) ImportData(ctx context.Context, userID string, data *model.ExportData, mode string) error {
	if data == nil {
		return errors.New("导入数据为空")
	}
	if mode == "overwrite" {
		if err := s.clearUserData(ctx, userID); err != nil {
			return err
		}
	}

	// append 模式：已有菜谱名去重
	existingNames := map[string]bool{}
	if mode == "append" {
		existing, err := s.recipeRepo.ListAll(ctx, userID)
		if err != nil {
			return err
		}
		for _, r := range existing {
			existingNames[r.Name] = true
		}
	}

	// 1. 导入菜谱，记录 oldID -> newID
	idMap := map[string]string{}
	for _, r := range data.Recipes {
		if r == nil || existingNames[r.Name] {
			continue
		}
		oldID := r.ID

		// 备份里的标签是名称，导入时映射为当前库的标签 id（未收录的名称建为自定义标签）
		manualIDs, err := s.tagService.RecognizeTagNames(ctx, userID, r.Tags)
		if err != nil {
			return fmt.Errorf("导入菜谱 %q 的标签失败: %w", r.Name, err)
		}
		ingredientIDs, err := s.tagService.RecognizeTagNames(ctx, userID, r.IngredientTags)
		if err != nil {
			return fmt.Errorf("导入菜谱 %q 的食材标签失败: %w", r.Name, err)
		}

		rec := cloneRecipe(r)
		rec.UserID = userID
		if err := s.recipeRepo.Create(ctx, rec); err != nil {
			return fmt.Errorf("导入菜谱 %q 失败: %w", r.Name, err)
		}
		if err := s.tagRepo.ReplaceRecipeTags(ctx, rec.ID, manualIDs); err != nil {
			return fmt.Errorf("导入菜谱 %q 的标签关联失败: %w", r.Name, err)
		}
		if err := s.tagRepo.ReplaceRecipeIngredientTags(ctx, rec.ID, ingredientIDs); err != nil {
			return fmt.Errorf("导入菜谱 %q 的食材标签关联失败: %w", r.Name, err)
		}
		idMap[oldID] = rec.ID
	}

	// 2. 导入餐厅，记录 oldRestID -> newRestID
	restMap := map[string]string{}
	for _, rst := range data.Restaurants {
		if rst == nil {
			continue
		}
		oldID := rst.ID
		rec := *rst
		rec.ID = ""
		rec.UserID = userID
		rec.DishCount = 0

		// 备份里的标签是名称，导入时映射为当前库的标签 id（未收录的名称建为自定义标签）
		tagIDs, err := s.shopTagService.RecognizeTagNames(ctx, userID, repository.ShopTagDomainRestaurant, rec.Tags)
		if err != nil {
			return fmt.Errorf("导入餐厅 %q 的标签失败: %w", rst.Name, err)
		}
		if err := s.restaurantRepo.Create(ctx, &rec); err != nil {
			return fmt.Errorf("导入餐厅 %q 失败: %w", rst.Name, err)
		}
		if err := s.shopTagRepo.ReplaceTags(ctx, repository.ShopTagDomainRestaurant, rec.ID, tagIDs); err != nil {
			return fmt.Errorf("写入餐厅 %q 的标签失败: %w", rst.Name, err)
		}
		rec.Tags = tagIDs
		restMap[oldID] = rec.ID
	}

	// 3. 导入菜品（原餐厅未导入则跳过该菜品）
	for _, d := range data.Dishes {
		if d == nil {
			continue
		}
		newRestID, ok := restMap[d.RestaurantID]
		if !ok {
			continue
		}
		rec := *d
		rec.ID = ""
		rec.UserID = userID
		rec.RestaurantID = newRestID

		tagIDs, err := s.shopTagService.RecognizeTagNames(ctx, userID, repository.ShopTagDomainDish, rec.Tags)
		if err != nil {
			return fmt.Errorf("导入菜品 %q 的标签失败: %w", d.Name, err)
		}
		if err := s.dishRepo.Create(ctx, &rec); err != nil {
			return fmt.Errorf("导入菜品 %q 失败: %w", d.Name, err)
		}
		if err := s.shopTagRepo.ReplaceTags(ctx, repository.ShopTagDomainDish, rec.ID, tagIDs); err != nil {
			return fmt.Errorf("写入菜品 %q 的标签失败: %w", d.Name, err)
		}
	}

	// 4. 导入收藏（对应菜谱未导入则跳过）
	for _, rid := range data.Favorites {
		newID, ok := idMap[rid]
		if !ok {
			continue
		}
		if err := s.favoriteRepo.Add(ctx, userID, newID); err != nil {
			return err
		}
	}
	return nil
}

// clearUserData 清空用户全部业务数据（收藏/菜品/菜谱/餐厅）
func (s *ExportService) clearUserData(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	stmts := []string{
		`DELETE FROM recipe_favorites WHERE user_id = $1`,
		`DELETE FROM dishes WHERE user_id = $1`,
		`DELETE FROM recipes WHERE user_id = $1`,
		`DELETE FROM restaurants WHERE user_id = $1`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(ctx, stmt, userID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func cloneRecipe(r *model.Recipe) *model.Recipe {
	cp := *r
	cp.ID = ""
	cp.UserID = ""
	cp.CreatedAt = time.Time{}
	cp.UpdatedAt = time.Time{}
	cp.IsFavorited = false
	return &cp
}

func joinIngredients(ings []model.Ingredient) string {
	parts := make([]string, len(ings))
	for i, ing := range ings {
		s := ing.Name
		if ing.Amount != "" {
			s += " " + ing.Amount
		}
		if ing.Unit != "" {
			s += ing.Unit
		}
		parts[i] = s
	}
	return strings.Join(parts, "; ")
}

func joinSteps(steps []model.Step) string {
	parts := make([]string, len(steps))
	for i, st := range steps {
		parts[i] = strconv.Itoa(st.Order) + ". " + st.Content
	}
	return strings.Join(parts, " | ")
}

func itov(v int) string {
	return strconv.Itoa(v)
}
