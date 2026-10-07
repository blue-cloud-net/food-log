package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// RecipeRepo 菜谱数据访问
type RecipeRepo struct {
	pool *pgxpool.Pool
}

func NewRecipeRepo(pool *pgxpool.Pool) *RecipeRepo {
	return &RecipeRepo{pool: pool}
}

const recipeCols = `id, user_id, name, COALESCE(description,''), ingredients, steps,
	cook_time_minutes, COALESCE(difficulty,''), rating, images, made_at, is_liked, created_at, updated_at`

// dbExecutor 同时兼容 *pgxpool.Pool 与 pgx.Tx，便于复用同一段 SQL
type dbExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanRecipe(row pgx.Row) (*model.Recipe, error) {
	var r model.Recipe
	var ingredients, steps, images []byte
	var cookTime *int
	var rating *int
	// made_at 是 DATE 列：pgx 无法直接把 date 扫进 **string，须经 *time.Time 中转再格式化为 YYYY-MM-DD
	var madeAt *time.Time
	err := row.Scan(&r.ID, &r.UserID, &r.Name, &r.Description,
		&ingredients, &steps, &cookTime, &r.Difficulty, &rating, &images,
		&madeAt, &r.IsLiked, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(ingredients, &r.Ingredients)
	json.Unmarshal(steps, &r.Steps)
	json.Unmarshal(images, &r.Images)
	if cookTime != nil {
		r.CookTimeMinutes = *cookTime
	}
	if rating != nil {
		r.Rating = *rating
	}
	if madeAt != nil {
		s := madeAt.Format("2006-01-02")
		r.MadeAt = &s
	}
	// 标签由 TagRepo 单独填充，避免 JSONB 字段与非空数组语义混淆
	r.Tags = []string{}
	r.IngredientTags = []string{}
	return &r, nil
}

// marshalArray 序列化 JSONB 数组：nil 切片序列化为 [] 而非 null
func marshalArray(v any) []byte {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		return []byte("[]")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("[]")
	}
	return b
}

// Create 创建菜谱
func (r *RecipeRepo) Create(ctx context.Context, rec *model.Recipe) error {
	return createRecipe(ctx, r.pool, rec)
}

// CreateTx 在事务内创建菜谱
func (r *RecipeRepo) CreateTx(ctx context.Context, tx pgx.Tx, rec *model.Recipe) error {
	return createRecipe(ctx, tx, rec)
}

func createRecipe(ctx context.Context, q dbExecutor, rec *model.Recipe) error {
	// RETURNING 会覆盖出参结构体，标签由调用方另行写入，这里先留存再还原
	manual, ingredient := rec.Tags, rec.IngredientTags

	ingredients := marshalArray(rec.Ingredients)
	steps := marshalArray(rec.Steps)
	images := marshalArray(rec.Images)

	row := q.QueryRow(ctx,
		`INSERT INTO recipes (user_id, name, description, ingredients, steps, cook_time_minutes, difficulty, rating, images, made_at, is_liked)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING `+recipeCols,
		rec.UserID, rec.Name, rec.Description, ingredients, steps,
		nullableInt(rec.CookTimeMinutes), nullableString(&rec.Difficulty), nullableInt(rec.Rating),
		images, nullableString(rec.MadeAt), rec.IsLiked)
	if err := scanRecipeRow(ctx, row, rec); err != nil {
		return err
	}
	rec.Tags, rec.IngredientTags = manual, ingredient
	return nil
}

// GetByID 通过 ID 查找菜谱
func (r *RecipeRepo) GetByID(ctx context.Context, id string) (*model.Recipe, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+recipeCols+` FROM recipes WHERE id = $1`, id)
	rec, err := scanRecipe(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

// List 菜谱列表（分页 + 搜索 + 筛选）
// tagID 过滤菜谱级标签，ingredientTagID 过滤食材级标签
// made 为 "true"/"false"（其余值不过滤）用于「已做/未做」，liked 为 true 时仅返回「喜欢」
func (r *RecipeRepo) List(ctx context.Context, userID string, q *model.PageQuery, keyword, difficulty, tagID, ingredientTagID, sort string, isFavorite bool, made string, liked bool) ([]*model.Recipe, int64, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("user_id = $%d", len(args)))

	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if difficulty != "" {
		args = append(args, difficulty)
		where = append(where, fmt.Sprintf("difficulty = $%d", len(args)))
	}
	if tagID != "" {
		args = append(args, tagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM recipe_tags rt WHERE rt.recipe_id = recipes.id AND rt.tag_id = $%d::uuid)", len(args)))
	}
	if ingredientTagID != "" {
		args = append(args, ingredientTagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM recipe_ingredient_tags rit WHERE rit.recipe_id = recipes.id AND rit.tag_id = $%d::uuid)", len(args)))
	}
	if isFavorite {
		args = append(args, userID)
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM recipe_favorites rf WHERE rf.recipe_id = recipes.id AND rf.user_id = $%d)", len(args)))
	}
	switch made {
	case "true":
		where = append(where, "made_at IS NOT NULL")
	case "false":
		where = append(where, "made_at IS NULL")
	}
	if liked {
		where = append(where, "is_liked = true")
	}
	whereClause := strings.Join(where, " AND ")

	// 总数
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM recipes WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	switch sort {
	case "rating":
		orderBy = "rating DESC NULLS LAST, created_at DESC"
	case "cook_time":
		orderBy = "cook_time_minutes ASC NULLS LAST"
	}

	args = append(args, q.PageSize, q.Offset())
	sql := fmt.Sprintf(`SELECT %s FROM recipes WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		recipeCols, whereClause, orderBy, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []*model.Recipe{}
	for rows.Next() {
		rec, err := scanRecipe(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, rec)
	}
	return list, total, rows.Err()
}

// Update 更新菜谱
func (r *RecipeRepo) Update(ctx context.Context, rec *model.Recipe) error {
	return updateRecipe(ctx, r.pool, rec)
}

// UpdateTx 在事务内更新菜谱
func (r *RecipeRepo) UpdateTx(ctx context.Context, tx pgx.Tx, rec *model.Recipe) error {
	return updateRecipe(ctx, tx, rec)
}

func updateRecipe(ctx context.Context, q dbExecutor, rec *model.Recipe) error {
	// RETURNING 会覆盖出参结构体，标签由调用方另行写入，这里先留存再还原
	manual, ingredient := rec.Tags, rec.IngredientTags

	ingredients := marshalArray(rec.Ingredients)
	steps := marshalArray(rec.Steps)
	images := marshalArray(rec.Images)

	row := q.QueryRow(ctx,
		`UPDATE recipes SET
		   name=$2, description=$3, ingredients=$4, steps=$5,
		   cook_time_minutes=$6, difficulty=$7, rating=$8, images=$9,
		   made_at=$10, is_liked=$11, updated_at=now()
		 WHERE id=$1 RETURNING `+recipeCols,
		rec.ID, rec.Name, rec.Description, ingredients, steps,
		nullableInt(rec.CookTimeMinutes), nullableString(&rec.Difficulty), nullableInt(rec.Rating),
		images, nullableString(rec.MadeAt), rec.IsLiked)
	if err := scanRecipeRow(ctx, row, rec); err != nil {
		return err
	}
	rec.Tags, rec.IngredientTags = manual, ingredient
	return nil
}

// Delete 删除菜谱
func (r *RecipeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recipes WHERE id = $1`, id)
	return err
}

// SetMade 设置「做过日期」（nil 表示未做）
func (r *RecipeRepo) SetMade(ctx context.Context, id string, madeAt *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE recipes SET made_at = $2, updated_at = now() WHERE id = $1`,
		id, nullableString(madeAt))
	return err
}

// SetLiked 设置「喜欢」标记
func (r *RecipeRepo) SetLiked(ctx context.Context, id string, liked bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE recipes SET is_liked = $2, updated_at = now() WHERE id = $1`,
		id, liked)
	return err
}

func scanRecipeRow(ctx context.Context, row pgx.Row, out *model.Recipe) error {
	rec, err := scanRecipe(row)
	if err != nil {
		return err
	}
	*out = *rec
	return nil
}

func nullableInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

// Random 随机返回一条菜谱（可按标签/难度过滤），无记录时返回 nil
func (r *RecipeRepo) Random(ctx context.Context, userID, tagID, ingredientTagID, difficulty string) (*model.Recipe, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("user_id = $%d", len(args)))
	if difficulty != "" {
		args = append(args, difficulty)
		where = append(where, fmt.Sprintf("difficulty = $%d", len(args)))
	}
	if tagID != "" {
		args = append(args, tagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM recipe_tags rt WHERE rt.recipe_id = recipes.id AND rt.tag_id = $%d::uuid)", len(args)))
	}
	if ingredientTagID != "" {
		args = append(args, ingredientTagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM recipe_ingredient_tags rit WHERE rit.recipe_id = recipes.id AND rit.tag_id = $%d::uuid)", len(args)))
	}

	row := r.pool.QueryRow(ctx,
		`SELECT `+recipeCols+` FROM recipes WHERE `+strings.Join(where, " AND ")+` ORDER BY random() LIMIT 1`,
		args...)
	rec, err := scanRecipe(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

// ListAll 返回用户全部菜谱（导出用）
func (r *RecipeRepo) ListAll(ctx context.Context, userID string) ([]*model.Recipe, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+recipeCols+` FROM recipes WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Recipe{}
	for rows.Next() {
		rec, err := scanRecipe(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rec)
	}
	return list, rows.Err()
}

// SearchByName 按菜名模糊搜索（全局搜索用）
func (r *RecipeRepo) SearchByName(ctx context.Context, userID, keyword string, limit int) ([]*model.Recipe, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+recipeCols+` FROM recipes
		 WHERE user_id = $1 AND name ILIKE $2
		 ORDER BY created_at DESC LIMIT $3`,
		userID, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Recipe{}
	for rows.Next() {
		rec, err := scanRecipe(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rec)
	}
	return list, rows.Err()
}
