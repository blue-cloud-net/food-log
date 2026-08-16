package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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
	cook_time_minutes, COALESCE(difficulty,''), rating, tags, images, created_at, updated_at`

func scanRecipe(row pgx.Row) (*model.Recipe, error) {
	var r model.Recipe
	var ingredients, steps, tags, images []byte
	var cookTime *int
	var rating *int
	err := row.Scan(&r.ID, &r.UserID, &r.Name, &r.Description,
		&ingredients, &steps, &cookTime, &r.Difficulty, &rating, &tags, &images,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(ingredients, &r.Ingredients)
	json.Unmarshal(steps, &r.Steps)
	json.Unmarshal(tags, &r.Tags)
	json.Unmarshal(images, &r.Images)
	if cookTime != nil {
		r.CookTimeMinutes = *cookTime
	}
	if rating != nil {
		r.Rating = *rating
	}
	return &r, nil
}

// Create 创建菜谱
func (r *RecipeRepo) Create(ctx context.Context, rec *model.Recipe) error {
	ingredients, _ := json.Marshal(rec.Ingredients)
	steps, _ := json.Marshal(rec.Steps)
	tags, _ := json.Marshal(rec.Tags)
	images, _ := json.Marshal(rec.Images)

	row := r.pool.QueryRow(ctx,
		`INSERT INTO recipes (user_id, name, description, ingredients, steps, cook_time_minutes, difficulty, rating, tags, images)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING `+recipeCols,
		rec.UserID, rec.Name, rec.Description, ingredients, steps,
		nullableInt(rec.CookTimeMinutes), rec.Difficulty, nullableInt(rec.Rating),
		tags, images)
	return scanRecipeRow(ctx, row, rec)
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
func (r *RecipeRepo) List(ctx context.Context, userID string, q *model.PageQuery, keyword, difficulty, tag, sort string) ([]*model.Recipe, int64, error) {
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
	if tag != "" {
		args = append(args, tag)
		where = append(where, fmt.Sprintf("tags @> $%d::jsonb", len(args)))
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
	ingredients, _ := json.Marshal(rec.Ingredients)
	steps, _ := json.Marshal(rec.Steps)
	tags, _ := json.Marshal(rec.Tags)
	images, _ := json.Marshal(rec.Images)

	row := r.pool.QueryRow(ctx,
		`UPDATE recipes SET
		   name=$2, description=$3, ingredients=$4, steps=$5,
		   cook_time_minutes=$6, difficulty=$7, rating=$8, tags=$9, images=$10
		 WHERE id=$1 RETURNING `+recipeCols,
		rec.ID, rec.Name, rec.Description, ingredients, steps,
		nullableInt(rec.CookTimeMinutes), rec.Difficulty, nullableInt(rec.Rating),
		tags, images)
	return scanRecipeRow(ctx, row, rec)
}

// Delete 删除菜谱
func (r *RecipeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recipes WHERE id = $1`, id)
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
