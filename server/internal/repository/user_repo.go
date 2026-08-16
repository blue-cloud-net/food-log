package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// UserRepo 用户数据访问
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, avatar_url)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		u.Username, u.Email, u.PasswordHash, u.AvatarURL,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetByUsername 通过用户名查找用户
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.getBy(ctx, "username = $1", username)
}

// GetByEmail 通过邮箱查找用户
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.getBy(ctx, "email = $1", email)
}

// GetByID 通过 ID 查找用户
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	return r.getBy(ctx, "id = $1", id)
}

func (r *UserRepo) getBy(ctx context.Context, where string, arg any) (*model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, COALESCE(avatar_url,''), created_at, updated_at
		 FROM users WHERE `+where,
		arg,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Update 更新用户信息
func (r *UserRepo) Update(ctx context.Context, u *model.User) error {
	return r.pool.QueryRow(ctx,
		`UPDATE users
		 SET username = $2, email = $3, avatar_url = $4,
		     password_hash = COALESCE($5, password_hash)
		 WHERE id = $1
		 RETURNING created_at, updated_at`,
		u.ID, u.Username, u.Email, u.AvatarURL, u.PasswordHash,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
}

// GetStats 获取用户统计
func (r *UserRepo) GetStats(ctx context.Context, userID string) (*model.UserStats, error) {
	var stats model.UserStats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM recipes    WHERE user_id = $1) AS recipe_count,
			(SELECT COUNT(*) FROM restaurants WHERE user_id = $1) AS restaurant_count,
			(SELECT COUNT(*) FROM dishes      WHERE user_id = $1) AS dish_count
	`, userID).Scan(&stats.RecipeCount, &stats.RestaurantCount, &stats.DishCount)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UsernameExists 检查用户名是否存在
func (r *UserRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

// EmailExists 检查邮箱是否存在
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
