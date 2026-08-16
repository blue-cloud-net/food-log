package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FavoriteRepo 菜谱收藏数据访问
type FavoriteRepo struct {
	pool *pgxpool.Pool
}

func NewFavoriteRepo(pool *pgxpool.Pool) *FavoriteRepo {
	return &FavoriteRepo{pool: pool}
}

// Add 收藏（幂等）
func (r *FavoriteRepo) Add(ctx context.Context, userID, recipeID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO recipe_favorites (user_id, recipe_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, recipe_id) DO NOTHING`,
		userID, recipeID)
	return err
}

// Remove 取消收藏
func (r *FavoriteRepo) Remove(ctx context.Context, userID, recipeID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM recipe_favorites WHERE user_id = $1 AND recipe_id = $2`,
		userID, recipeID)
	return err
}

// IsFavorited 是否已收藏
func (r *FavoriteRepo) IsFavorited(ctx context.Context, userID, recipeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM recipe_favorites WHERE user_id = $1 AND recipe_id = $2)`,
		userID, recipeID).Scan(&exists)
	return exists, err
}

// ListIDs 当前用户收藏的菜谱 ID 列表
func (r *FavoriteRepo) ListIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT recipe_id FROM recipe_favorites WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
