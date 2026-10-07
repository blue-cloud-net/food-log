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

// InventoryRepo 库存食材数据访问
type InventoryRepo struct {
	pool *pgxpool.Pool
}

func NewInventoryRepo(pool *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{pool: pool}
}

// inventoryCols 带 i. 前缀，因此 INSERT / UPDATE ... RETURNING 需写成 `inventory_items AS i`
const inventoryCols = `i.id, i.user_id, i.location_id, i.name, i.quantity, i.unit,
	i.category, i.expire_at, i.note, i.images, i.created_at, i.updated_at`

// InventoryFilter 库存列表筛选条件
type InventoryFilter struct {
	Keyword            string // 名称 / 备注模糊搜索
	LocationID         string // 指定存放位置
	Area               string // 存放位置大类：fridge / outside
	ExpiringWithinDays int    // >0 时仅返回 N 天内到期（含已过期）的条目
	Sort               string // expire（默认）/ created / name
}

func scanInventoryItem(row pgx.Row) (*model.InventoryItem, error) {
	var it model.InventoryItem
	var images []byte
	var expireAt *time.Time
	err := row.Scan(&it.ID, &it.UserID, &it.LocationID, &it.Name, &it.Quantity, &it.Unit,
		&it.Category, &expireAt, &it.Note, &images, &it.CreatedAt, &it.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(images, &it.Images)
	if it.Images == nil {
		it.Images = []string{}
	}
	if expireAt != nil {
		s := expireAt.Format("2006-01-02")
		it.ExpireAt = &s
	}
	return &it, nil
}

// List 库存列表（筛选 + 排序）。条目量级小，不做分页，分组在前端完成。
func (r *InventoryRepo) List(ctx context.Context, userID string, f InventoryFilter) ([]*model.InventoryItem, error) {
	var where []string
	var args []any
	args = append(args, userID)
	where = append(where, fmt.Sprintf("i.user_id = $%d", len(args)))

	if f.Keyword != "" {
		args = append(args, "%"+f.Keyword+"%")
		where = append(where, fmt.Sprintf("(i.name ILIKE $%d OR i.note ILIKE $%d)", len(args), len(args)))
	}
	if f.LocationID != "" {
		args = append(args, f.LocationID)
		where = append(where, fmt.Sprintf("i.location_id = $%d::uuid", len(args)))
	}
	if f.Area != "" {
		args = append(args, f.Area)
		where = append(where, fmt.Sprintf(
			"i.location_id IN (SELECT id FROM storage_locations WHERE area = $%d)", len(args)))
	}
	if f.ExpiringWithinDays > 0 {
		args = append(args, f.ExpiringWithinDays)
		where = append(where, fmt.Sprintf(
			"i.expire_at IS NOT NULL AND i.expire_at <= CURRENT_DATE + $%d::int", len(args)))
	}

	orderBy := "i.expire_at ASC NULLS LAST, i.created_at DESC"
	switch f.Sort {
	case "created":
		orderBy = "i.created_at DESC"
	case "name":
		orderBy = "i.name ASC"
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(
		`SELECT %s FROM inventory_items i WHERE %s ORDER BY %s`,
		inventoryCols, strings.Join(where, " AND "), orderBy), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.InventoryItem{}
	for rows.Next() {
		it, err := scanInventoryItem(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, rows.Err()
}

// GetByID 通过 ID 查找库存条目，不存在返回 nil
func (r *InventoryRepo) GetByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	return scanInventoryItem(r.pool.QueryRow(ctx,
		`SELECT `+inventoryCols+` FROM inventory_items i WHERE i.id = $1::uuid`, id))
}

// Create 创建库存条目
func (r *InventoryRepo) Create(ctx context.Context, it *model.InventoryItem) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO inventory_items AS i (user_id, location_id, name, quantity, unit, category, expire_at, note, images)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+inventoryCols,
		it.UserID, it.LocationID, it.Name, it.Quantity, it.Unit, it.Category,
		nullableString(it.ExpireAt), it.Note, marshalArray(it.Images))
	return scanInventoryItemRow(row, it)
}

// Update 更新库存条目
func (r *InventoryRepo) Update(ctx context.Context, it *model.InventoryItem) error {
	row := r.pool.QueryRow(ctx,
		`UPDATE inventory_items AS i SET
		   location_id=$2::uuid, name=$3, quantity=$4, unit=$5, category=$6,
		   expire_at=$7, note=$8, images=$9, updated_at=now()
		 WHERE i.id=$1::uuid RETURNING `+inventoryCols,
		it.ID, it.LocationID, it.Name, it.Quantity, it.Unit, it.Category,
		nullableString(it.ExpireAt), it.Note, marshalArray(it.Images))
	return scanInventoryItemRow(row, it)
}

// Delete 删除库存条目
func (r *InventoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1::uuid`, id)
	return err
}

// DeleteMany 批量删除该用户的库存条目，返回实际删除数量
func (r *InventoryRepo) DeleteMany(ctx context.Context, userID string, ids []string) (int64, error) {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM inventory_items WHERE user_id = $1::uuid AND id = ANY($2::uuid[])`, userID, ids)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// Consume 消耗 amount 数量：仅在条目属于该用户且剩余数量不少于 amount 时生效。
// 返回剩余数量；ok = false 表示条件未命中（条目不存在 / 不属于该用户 / 数量不足）。
// 剩余数量减到 0 时直接删除该条目，removed = true。
func (r *InventoryRepo) Consume(ctx context.Context, userID, id string, amount float64) (remaining float64, removed bool, ok bool, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, false, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err = tx.QueryRow(ctx,
		`UPDATE inventory_items SET quantity = quantity - $3, updated_at = now()
		 WHERE id = $1::uuid AND user_id = $2::uuid AND quantity >= $3
		 RETURNING quantity`, id, userID, amount).Scan(&remaining); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, false, nil
		}
		return 0, false, false, err
	}

	if remaining <= 0 {
		if _, err = tx.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1::uuid`, id); err != nil {
			return 0, false, false, err
		}
		removed = true
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, false, false, err
	}
	return remaining, removed, true, nil
}

func scanInventoryItemRow(row pgx.Row, out *model.InventoryItem) error {
	it, err := scanInventoryItem(row)
	if err != nil {
		return err
	}
	*out = *it
	return nil
}
