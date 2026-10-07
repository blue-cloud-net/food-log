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

// RestaurantRepo 餐厅数据访问
type RestaurantRepo struct {
	pool *pgxpool.Pool
}

func NewRestaurantRepo(pool *pgxpool.Pool) *RestaurantRepo {
	return &RestaurantRepo{pool: pool}
}

const restaurantCols = `r.id, r.user_id, r.name, COALESCE(r.address,''), COALESCE(r.cuisine_type,''),
	COALESCE(r.description,''), COALESCE(r.avg_rating,0), r.images, r.lat, r.lng,
	(SELECT COUNT(*) FROM dishes d WHERE d.restaurant_id = r.id) AS dish_count,
	r.created_at, r.updated_at`

func scanRestaurant(row pgx.Row) (*model.Restaurant, error) {
	var rst model.Restaurant
	var images []byte
	err := row.Scan(&rst.ID, &rst.UserID, &rst.Name, &rst.Address, &rst.CuisineType,
		&rst.Description, &rst.AvgRating, &images, &rst.Lat, &rst.Lng,
		&rst.DishCount, &rst.CreatedAt, &rst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &rst.Images)
	return &rst, nil
}

// Create 创建餐厅
func (r *RestaurantRepo) Create(ctx context.Context, rst *model.Restaurant) error {
	images, _ := json.Marshal(rst.Images)
	// restaurantCols 使用 r. 前缀，故 INSERT 目标表需同名别名
	row := r.pool.QueryRow(ctx,
		`INSERT INTO restaurants AS r (user_id, name, address, cuisine_type, description, avg_rating, images, lat, lng)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING `+restaurantCols,
		rst.UserID, rst.Name, rst.Address, rst.CuisineType, rst.Description,
		nullableFloat(rst.AvgRating), images, rst.Lat, rst.Lng)
	return scanRestaurantRow(ctx, row, rst)
}

// GetByID 通过 ID 查找餐厅
func (r *RestaurantRepo) GetByID(ctx context.Context, id string) (*model.Restaurant, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+restaurantCols+` FROM restaurants r WHERE r.id = $1`, id)
	rst, err := scanRestaurant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rst, err
}

// List 餐厅列表（分页 + 搜索 + 菜系筛选）
func (r *RestaurantRepo) List(ctx context.Context, userID string, q *model.PageQuery, keyword, cuisineType, sort string) ([]*model.Restaurant, int64, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("r.user_id = $%d", len(args)))

	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf("(r.name ILIKE $%d OR r.address ILIKE $%d)", len(args), len(args)))
	}
	if cuisineType != "" {
		args = append(args, cuisineType)
		where = append(where, fmt.Sprintf("r.cuisine_type = $%d", len(args)))
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM restaurants r WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "r.created_at DESC"
	if sort == "rating" {
		orderBy = "r.avg_rating DESC NULLS LAST, r.created_at DESC"
	}

	args = append(args, q.PageSize, q.Offset())
	sql := fmt.Sprintf(`SELECT %s FROM restaurants r WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		restaurantCols, whereClause, orderBy, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []*model.Restaurant{}
	for rows.Next() {
		rst, err := scanRestaurant(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, rst)
	}
	return list, total, rows.Err()
}

// Update 更新餐厅
func (r *RestaurantRepo) Update(ctx context.Context, rst *model.Restaurant) error {
	images, _ := json.Marshal(rst.Images)
	row := r.pool.QueryRow(ctx,
		`UPDATE restaurants SET
		   name=$2, address=$3, cuisine_type=$4, description=$5,
		   avg_rating=$6, images=$7, lat=$8, lng=$9
		 WHERE id=$1 RETURNING `+restaurantCols,
		rst.ID, rst.Name, rst.Address, rst.CuisineType, rst.Description,
		nullableFloat(rst.AvgRating), images, rst.Lat, rst.Lng)
	return scanRestaurantRow(ctx, row, rst)
}

// Delete 删除餐厅（级联删除菜品）
func (r *RestaurantRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM restaurants WHERE id = $1`, id)
	return err
}

// UpdateAvgRating 根据菜品评分重算餐厅平均分
func (r *RestaurantRepo) UpdateAvgRating(ctx context.Context, restaurantID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE restaurants SET avg_rating = (
			SELECT ROUND(AVG(rating)::numeric, 1) FROM dishes WHERE restaurant_id = $1 AND rating IS NOT NULL
		) WHERE id = $1`, restaurantID)
	return err
}

func scanRestaurantRow(ctx context.Context, row pgx.Row, out *model.Restaurant) error {
	rst, err := scanRestaurant(row)
	if err != nil {
		return err
	}
	*out = *rst
	return nil
}

func nullableFloat(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}

// ListAll 返回用户全部餐厅（导出用）
func (r *RestaurantRepo) ListAll(ctx context.Context, userID string) ([]*model.Restaurant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+restaurantCols+` FROM restaurants r WHERE r.user_id = $1 ORDER BY r.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Restaurant{}
	for rows.Next() {
		rst, err := scanRestaurant(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rst)
	}
	return list, rows.Err()
}

// SearchByName 按店名/地址/菜系模糊搜索（全局搜索用）
func (r *RestaurantRepo) SearchByName(ctx context.Context, userID, keyword string, limit int) ([]*model.Restaurant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+restaurantCols+` FROM restaurants r
		 WHERE r.user_id = $1 AND (r.name ILIKE $2 OR r.address ILIKE $2 OR r.cuisine_type ILIKE $2)
		 ORDER BY r.created_at DESC LIMIT $3`,
		userID, "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.Restaurant{}
	for rows.Next() {
		rst, err := scanRestaurant(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rst)
	}
	return list, rows.Err()
}
