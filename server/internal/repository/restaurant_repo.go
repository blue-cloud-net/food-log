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

// restaurantSortCols 列表排序白名单：查询参数 → 排序列（避免拼接注入）
var restaurantSortCols = map[string]string{
	"recommend": "r.recommend_rating",
	"value":     "r.value_rating",
	"ambience":  "r.ambience_rating",
	"service":   "r.service_rating",
}

// restaurantCols 标签列不在此处：标签由 ShopTagRepo 单独填充，避免 SQL 膨胀。
// 带 r. 前缀，因此 INSERT / UPDATE ... RETURNING 必须写成 `restaurants AS r`。
const restaurantCols = `r.id, r.user_id, r.name, COALESCE(r.address,''),
	COALESCE(r.description,''), COALESCE(r.recommend_rating,0), COALESCE(r.value_rating,0),
	COALESCE(r.ambience_rating,0), COALESCE(r.service_rating,0), r.images, r.lat, r.lng,
	r.is_visited,
	(SELECT COUNT(*) FROM dishes d WHERE d.restaurant_id = r.id) AS dish_count,
	r.created_at, r.updated_at`

func scanRestaurant(row pgx.Row) (*model.Restaurant, error) {
	var rst model.Restaurant
	var images []byte
	err := row.Scan(&rst.ID, &rst.UserID, &rst.Name, &rst.Address,
		&rst.Description, &rst.RecommendRating, &rst.ValueRating,
		&rst.AmbienceRating, &rst.ServiceRating, &images, &rst.Lat, &rst.Lng,
		&rst.IsVisited, &rst.DishCount, &rst.CreatedAt, &rst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &rst.Images)
	// 标签由 ShopTagRepo 单独填充，先置空数组保证 JSON 输出 [] 而非 null
	rst.Tags = []string{}
	return &rst, nil
}

// Create 创建餐厅（标签关联由调用方在同一事务内另行写入）
func (r *RestaurantRepo) Create(ctx context.Context, rst *model.Restaurant) error {
	return createRestaurant(ctx, r.pool, rst)
}

// CreateTx 在事务内创建餐厅
func (r *RestaurantRepo) CreateTx(ctx context.Context, tx pgx.Tx, rst *model.Restaurant) error {
	return createRestaurant(ctx, tx, rst)
}

func createRestaurant(ctx context.Context, q dbExecutor, rst *model.Restaurant) error {
	images, _ := json.Marshal(rst.Images)
	// RETURNING 会覆盖出参结构体，标签由调用方另行写入，这里先留存再还原
	tags := rst.Tags
	row := q.QueryRow(ctx,
		`INSERT INTO restaurants AS r (user_id, name, address, description,
		   recommend_rating, value_rating, ambience_rating, service_rating, images, lat, lng, is_visited)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING `+restaurantCols,
		rst.UserID, rst.Name, rst.Address, rst.Description,
		nullableInt(rst.RecommendRating), nullableInt(rst.ValueRating),
		nullableInt(rst.AmbienceRating), nullableInt(rst.ServiceRating),
		images, rst.Lat, rst.Lng, rst.IsVisited)
	if err := scanRestaurantRow(ctx, row, rst); err != nil {
		return err
	}
	rst.Tags = tags
	return nil
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

// List 餐厅列表（分页 + 关键词 + 标签筛选 + 排序）
// visited 为 "true"/"false"（其余值不过滤）用于「已探店/未探店」
func (r *RestaurantRepo) List(ctx context.Context, userID string, q *model.PageQuery, keyword, tagID, sort, visited string) ([]*model.Restaurant, int64, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("r.user_id = $%d", len(args)))

	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf("(r.name ILIKE $%d OR r.address ILIKE $%d)", len(args), len(args)))
	}
	if tagID != "" {
		args = append(args, tagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM restaurant_tag_links rtl WHERE rtl.restaurant_id = r.id AND rtl.tag_id = $%d::uuid)", len(args)))
	}
	switch visited {
	case "true":
		where = append(where, "r.is_visited = true")
	case "false":
		where = append(where, "r.is_visited = false")
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM restaurants r WHERE `+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "r.created_at DESC"
	if col, ok := restaurantSortCols[sort]; ok {
		orderBy = fmt.Sprintf("(%s) DESC NULLS LAST, r.created_at DESC", col)
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

// Update 更新餐厅（标签关联由调用方在同一事务内另行写入）
func (r *RestaurantRepo) Update(ctx context.Context, rst *model.Restaurant) error {
	return updateRestaurant(ctx, r.pool, rst)
}

// UpdateTx 在事务内更新餐厅
func (r *RestaurantRepo) UpdateTx(ctx context.Context, tx pgx.Tx, rst *model.Restaurant) error {
	return updateRestaurant(ctx, tx, rst)
}

func updateRestaurant(ctx context.Context, q dbExecutor, rst *model.Restaurant) error {
	images, _ := json.Marshal(rst.Images)
	tags := rst.Tags
	row := q.QueryRow(ctx,
		`UPDATE restaurants AS r SET
		   name=$2, address=$3, description=$4,
		   recommend_rating=$5, value_rating=$6, ambience_rating=$7, service_rating=$8,
		   images=$9, lat=$10, lng=$11, is_visited=$12, updated_at=now()
		 WHERE r.id=$1 RETURNING `+restaurantCols,
		rst.ID, rst.Name, rst.Address, rst.Description,
		nullableInt(rst.RecommendRating), nullableInt(rst.ValueRating),
		nullableInt(rst.AmbienceRating), nullableInt(rst.ServiceRating),
		images, rst.Lat, rst.Lng, rst.IsVisited)
	if err := scanRestaurantRow(ctx, row, rst); err != nil {
		return err
	}
	rst.Tags = tags
	return nil
}

// Delete 删除餐厅（级联删除其下菜品与标签关联）
func (r *RestaurantRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM restaurants WHERE id = $1`, id)
	return err
}

// SetVisited 设置「已探店」标记
func (r *RestaurantRepo) SetVisited(ctx context.Context, id string, visited bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE restaurants SET is_visited = $2, updated_at = now() WHERE id = $1`,
		id, visited)
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

// SearchByName 按店名/地址/标签名模糊搜索（全局搜索用）
func (r *RestaurantRepo) SearchByName(ctx context.Context, userID, keyword string, limit int) ([]*model.Restaurant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+restaurantCols+` FROM restaurants r
		 WHERE r.user_id = $1 AND (
		   r.name ILIKE $2 OR r.address ILIKE $2
		   OR EXISTS (
		     SELECT 1 FROM restaurant_tag_links l
		     JOIN restaurant_tags t ON t.id = l.tag_id
		     WHERE l.restaurant_id = r.id AND t.name ILIKE $2
		   )
		 )
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
