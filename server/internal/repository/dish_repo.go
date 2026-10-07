package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// dishCols 标签列不在此处：标签由 ShopTagRepo 单独填充。
// rating 用 COALESCE 兜底：写入端空值存 NULL，直接 Scan 到 int 会报错。
const dishCols = `id, restaurant_id, user_id, name, COALESCE(description,''), price,
	COALESCE(rating,0), images, eaten_at, is_liked, created_at`

// dishListCols 在 dishCols 基础上追加所属餐厅名（标量子查询，避免 JOIN 时的列名歧义）
const dishListCols = dishCols + `,
	COALESCE((SELECT r.name FROM restaurants r WHERE r.id = dishes.restaurant_id), '')`

func scanDish(row pgx.Row) (*model.Dish, error) {
	var d model.Dish
	var images []byte
	// eaten_at 是 DATE 列：pgx 无法把 date 扫进 **string，必须经 *time.Time 中转再格式化为 YYYY-MM-DD
	var eatenAt *time.Time
	err := row.Scan(&d.ID, &d.RestaurantID, &d.UserID, &d.Name, &d.Description,
		&d.Price, &d.Rating, &images, &eatenAt, &d.IsLiked, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &d.Images)
	if eatenAt != nil {
		s := eatenAt.Format("2006-01-02")
		d.EatenAt = &s
	}
	// 标签由 ShopTagRepo 单独填充，先置空数组保证 JSON 输出 [] 而非 null
	d.Tags = []string{}
	return &d, nil
}

// scanDishWithRestaurant 扫描 dishListCols（末列为所属餐厅名）
func scanDishWithRestaurant(row pgx.Row) (*model.Dish, error) {
	var d model.Dish
	var images []byte
	var eatenAt *time.Time
	err := row.Scan(&d.ID, &d.RestaurantID, &d.UserID, &d.Name, &d.Description,
		&d.Price, &d.Rating, &images, &eatenAt, &d.IsLiked, &d.CreatedAt, &d.RestaurantName)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &d.Images)
	if eatenAt != nil {
		s := eatenAt.Format("2006-01-02")
		d.EatenAt = &s
	}
	d.Tags = []string{}
	return &d, nil
}

// Create 创建菜品（标签关联由调用方在同一事务内另行写入）
func (r *DishRepo) Create(ctx context.Context, d *model.Dish) error {
	return createDish(ctx, r.pool, d)
}

// CreateTx 在事务内创建菜品
func (r *DishRepo) CreateTx(ctx context.Context, tx pgx.Tx, d *model.Dish) error {
	return createDish(ctx, tx, d)
}

func createDish(ctx context.Context, q dbExecutor, d *model.Dish) error {
	images, _ := json.Marshal(d.Images)
	// RETURNING 会覆盖出参结构体，标签由调用方另行写入，这里先留存再还原
	tags := d.Tags
	row := q.QueryRow(ctx,
		`INSERT INTO dishes (restaurant_id, user_id, name, description, price, rating, images, eaten_at, is_liked)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING `+dishCols,
		d.RestaurantID, d.UserID, d.Name, d.Description, d.Price,
		nullableInt(d.Rating), images, nullableString(d.EatenAt), d.IsLiked)
	if err := scanDishRow(ctx, row, d); err != nil {
		return err
	}
	d.Tags = tags
	return nil
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

// ListByRestaurant 某餐厅的全部菜品，可按菜品标签筛选（tagID 为空则不过滤）
func (r *DishRepo) ListByRestaurant(ctx context.Context, restaurantID, tagID string) ([]*model.Dish, error) {
	sql := `SELECT ` + dishCols + ` FROM dishes WHERE restaurant_id = $1`
	args := []any{restaurantID}
	if tagID != "" {
		args = append(args, tagID)
		sql += fmt.Sprintf(
			` AND EXISTS (SELECT 1 FROM dish_tag_links dtl WHERE dtl.dish_id = dishes.id AND dtl.tag_id = $%d::uuid)`, len(args))
	}
	sql += ` ORDER BY eaten_at DESC NULLS LAST, created_at DESC`

	rows, err := r.pool.Query(ctx, sql, args...)
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

// Update 更新菜品（标签关联由调用方在同一事务内另行写入）
func (r *DishRepo) Update(ctx context.Context, d *model.Dish) error {
	return updateDish(ctx, r.pool, d)
}

// UpdateTx 在事务内更新菜品
func (r *DishRepo) UpdateTx(ctx context.Context, tx pgx.Tx, d *model.Dish) error {
	return updateDish(ctx, tx, d)
}

func updateDish(ctx context.Context, q dbExecutor, d *model.Dish) error {
	images, _ := json.Marshal(d.Images)
	tags := d.Tags
	row := q.QueryRow(ctx,
		`UPDATE dishes SET
		   name=$2, description=$3, price=$4, rating=$5, images=$6, eaten_at=$7, is_liked=$8
		 WHERE id=$1 RETURNING `+dishCols,
		d.ID, d.Name, d.Description, d.Price, nullableInt(d.Rating), images, nullableString(d.EatenAt), d.IsLiked)
	if err := scanDishRow(ctx, row, d); err != nil {
		return err
	}
	d.Tags = tags
	return nil
}

// List 菜品列表（分页 + 喜欢筛选 + 关键词 + 标签 + 所属餐厅），附带所属餐厅名
// liked 为 true 时仅返回「喜欢」的菜品
func (r *DishRepo) List(ctx context.Context, userID string, q *model.PageQuery, liked bool, keyword, tagID, restaurantID string) ([]*model.Dish, int64, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("user_id = $%d", len(args)))
	if liked {
		where = append(where, "is_liked = true")
	}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if restaurantID != "" {
		args = append(args, restaurantID)
		where = append(where, fmt.Sprintf("restaurant_id = $%d", len(args)))
	}
	if tagID != "" {
		args = append(args, tagID)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM dish_tag_links dtl WHERE dtl.dish_id = dishes.id AND dtl.tag_id = $%d::uuid)", len(args)))
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM dishes WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.PageSize, q.Offset())
	sql := fmt.Sprintf(`SELECT %s FROM dishes WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		dishListCols, whereClause, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []*model.Dish{}
	for rows.Next() {
		d, err := scanDishWithRestaurant(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, d)
	}
	return list, total, rows.Err()
}

// SetLiked 设置菜品「喜欢」标记
func (r *DishRepo) SetLiked(ctx context.Context, id string, liked bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE dishes SET is_liked = $2 WHERE id = $1`, id, liked)
	return err
}

// Delete 删除菜品（级联删除标签关联）
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

// SearchByName 按菜名或菜品标签名模糊搜索（全局搜索用）
func (r *DishRepo) SearchByName(ctx context.Context, userID, keyword string, limit int) ([]*model.Dish, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dishCols+` FROM dishes
		 WHERE user_id = $1 AND (
		   name ILIKE $2
		   OR EXISTS (
		     SELECT 1 FROM dish_tag_links l
		     JOIN dish_tags t ON t.id = l.tag_id
		     WHERE l.dish_id = dishes.id AND t.name ILIKE $2
		   )
		 )
		 ORDER BY created_at DESC LIMIT $3`,
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
