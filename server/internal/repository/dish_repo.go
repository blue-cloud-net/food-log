package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// DishRepo 菜品数据访问
type DishRepo struct {
	pool *pgxpool.Pool
}

func NewDishRepo(pool *pgxpool.Pool) *DishRepo {
	return &DishRepo{pool: pool}
}

const dishCols = `id, restaurant_id, user_id, name, COALESCE(description,''), price, rating, images, eaten_at, created_at`

func scanDish(row pgx.Row) (*model.Dish, error) {
	var d model.Dish
	var images []byte
	var eatenAt *time.Time
	err := row.Scan(&d.ID, &d.RestaurantID, &d.UserID, &d.Name, &d.Description,
		&d.Price, &d.Rating, &images, &eatenAt, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &d.Images)
	if eatenAt != nil {
		formatted := eatenAt.Format("2006-01-02")
		d.EatenAt = &formatted
	}
	return &d, nil
}

// Create 创建菜品
func (r *DishRepo) Create(ctx context.Context, d *model.Dish) error {
	images, _ := json.Marshal(d.Images)
	row := r.pool.QueryRow(ctx,
		`INSERT INTO dishes (restaurant_id, user_id, name, description, price, rating, images, eaten_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING `+dishCols,
		d.RestaurantID, d.UserID, d.Name, d.Description, d.Price,
		nullableInt(d.Rating), images, nullableString(d.EatenAt))
	return scanDishRow(ctx, row, d)
}

// GetByID 通过 ID 查找菜品
func (r *DishRepo) GetByID(ctx context.Context, id string) (*model.Dish, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+dishCols+` FROM dishes WHERE id = $1`, id)
	d, err := scanDish(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

// ListByRestaurant 某餐厅的全部菜品
func (r *DishRepo) ListByRestaurant(ctx context.Context, restaurantID string) ([]*model.Dish, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dishCols+` FROM dishes WHERE restaurant_id = $1 ORDER BY eaten_at DESC NULLS LAST, created_at DESC`,
		restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Dish{}
	for rows.Next() {
		d, err := scanDish(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// Update 更新菜品
func (r *DishRepo) Update(ctx context.Context, d *model.Dish) error {
	images, _ := json.Marshal(d.Images)
	row := r.pool.QueryRow(ctx,
		`UPDATE dishes SET
		   name=$2, description=$3, price=$4, rating=$5, images=$6, eaten_at=$7
		 WHERE id=$1 RETURNING `+dishCols,
		d.ID, d.Name, d.Description, d.Price, nullableInt(d.Rating), images, nullableString(d.EatenAt))
	return scanDishRow(ctx, row, d)
}

// Delete 删除菜品
func (r *DishRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM dishes WHERE id = $1`, id)
	return err
}

func scanDishRow(ctx context.Context, row pgx.Row, out *model.Dish) error {
	d, err := scanDish(row)
	if err != nil {
		return err
	}
	*out = *d
	return nil
}

func nullableString(v *string) any {
	if v == nil || *v == "" {
		return nil
	}
	return *v
}

// ListAll 返回用户全部菜品（导出用）
func (r *DishRepo) ListAll(ctx context.Context, userID string) ([]*model.Dish, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dishCols+` FROM dishes WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Dish{}
	for rows.Next() {
		d, err := scanDish(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// SearchByName 按菜名模糊搜索（全局搜索用）
func (r *DishRepo) SearchByName(ctx context.Context, userID, keyword string, limit int) ([]*model.Dish, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dishCols+` FROM dishes WHERE user_id = $1 AND name ILIKE $2 ORDER BY created_at DESC LIMIT $3`,
		userID, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Dish{}
	for rows.Next() {
		d, err := scanDish(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}
