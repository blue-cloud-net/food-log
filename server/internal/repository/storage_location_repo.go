package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// StorageLocationRepo 存放位置字典数据访问。
// 可见性规则：owner_id IS NULL 为全局预设，owner_id = 用户为私有自定义。
type StorageLocationRepo struct {
	pool *pgxpool.Pool
}

func NewStorageLocationRepo(pool *pgxpool.Pool) *StorageLocationRepo {
	return &StorageLocationRepo{pool: pool}
}

const storageLocationCols = `id, COALESCE(owner_id::text,''), area, name, sort_order, is_system`

func scanStorageLocation(row pgx.Row) (*model.StorageLocation, error) {
	var l model.StorageLocation
	err := row.Scan(&l.ID, &l.OwnerID, &l.Area, &l.Name, &l.SortOrder, &l.IsSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// ListVisible 返回该用户可见的存放位置（全局预设 + 本人自定义）
func (r *StorageLocationRepo) ListVisible(ctx context.Context, userID string) ([]*model.StorageLocation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+storageLocationCols+` FROM storage_locations
		 WHERE owner_id IS NULL OR owner_id = $1::uuid
		 ORDER BY area, sort_order, created_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.StorageLocation{}
	for rows.Next() {
		l, err := scanStorageLocation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

// GetByID 按 id 取位置（不含可见性过滤），不存在返回 nil
func (r *StorageLocationRepo) GetByID(ctx context.Context, id string) (*model.StorageLocation, error) {
	return scanStorageLocation(r.pool.QueryRow(ctx,
		`SELECT `+storageLocationCols+` FROM storage_locations WHERE id = $1::uuid`, id))
}

// Create 新建用户自定义位置
func (r *StorageLocationRepo) Create(ctx context.Context, userID, area, name string, sortOrder int) (*model.StorageLocation, error) {
	return scanStorageLocation(r.pool.QueryRow(ctx,
		`INSERT INTO storage_locations (owner_id, area, name, sort_order, is_system)
		 VALUES ($1::uuid, $2, $3, $4, false)
		 RETURNING `+storageLocationCols, userID, area, name, sortOrder))
}

// Update 更新本人自定义位置
func (r *StorageLocationRepo) Update(ctx context.Context, id, area, name string, sortOrder int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE storage_locations SET area=$2, name=$3, sort_order=$4 WHERE id=$1::uuid`,
		id, area, name, sortOrder)
	return err
}

// Delete 删除本人自定义位置
func (r *StorageLocationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM storage_locations WHERE id = $1::uuid`, id)
	return err
}

// NameTaken 该用户在同一大类下是否已有同名位置（含全局预设名，避免与预设重名）。
// excludeID 非空时排除自身（用于更新场景）。
func (r *StorageLocationRepo) NameTaken(ctx context.Context, userID, area, name, excludeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM storage_locations
			WHERE area = $2 AND name = $3
			  AND (owner_id IS NULL OR owner_id = $1::uuid)
			  AND (NULLIF($4,'')::uuid IS NULL OR id <> NULLIF($4,'')::uuid))`,
		userID, area, name, excludeID).Scan(&exists)
	return exists, err
}

// CountItems 统计引用该位置的库存条目数
func (r *StorageLocationRepo) CountItems(ctx context.Context, locationID string) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_items WHERE location_id = $1::uuid`, locationID).Scan(&n)
	return n, err
}
