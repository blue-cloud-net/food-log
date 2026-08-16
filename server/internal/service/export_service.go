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
}

func NewExportService(pool *pgxpool.Pool, recipeRepo *repository.RecipeRepo, restaurantRepo *repository.RestaurantRepo, dishRepo *repository.DishRepo, favoriteRepo *repository.FavoriteRepo) *ExportService {
	return &ExportService{pool: pool, recipeRepo: recipeRepo, restaurantRepo: restaurantRepo, dishRepo: dishRepo, favoriteRepo: favoriteRepo}
}

// ExportData 组装完整导出数据（JSON）
func (s *ExportService) ExportData(ctx context.Context, userID string) (*model.ExportData, error) {
	recipes, err := s.recipeRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	restaurants, err := s.restaurantRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	dishes, err := s.dishRepo.ListAll(ctx, userID)
	if err != nil {
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

// ExportCSV 导出菜谱 CSV 内容
func (s *ExportService) ExportCSV(ctx context.Context, userID string) ([]byte, error) {
	recipes, err := s.recipeRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"菜名", "描述", "食材", "步骤", "耗时(分钟)", "难度", "评分", "标签", "创建时间"}); err != nil {
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
		rec := cloneRecipe(r)
		rec.UserID = userID
		if err := s.recipeRepo.Create(ctx, rec); err != nil {
			return fmt.Errorf("导入菜谱 %q 失败: %w", r.Name, err)
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
		if err := s.restaurantRepo.Create(ctx, &rec); err != nil {
			return fmt.Errorf("导入餐厅 %q 失败: %w", rst.Name, err)
		}
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
		if err := s.dishRepo.Create(ctx, &rec); err != nil {
			return fmt.Errorf("导入菜品 %q 失败: %w", d.Name, err)
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
