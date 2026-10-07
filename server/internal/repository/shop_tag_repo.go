package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// ShopTagDomain 探店标签域：餐厅与菜品各有一套独立的字典表
type ShopTagDomain string

const (
	ShopTagDomainRestaurant ShopTagDomain = "restaurant"
	ShopTagDomainDish       ShopTagDomain = "dish"
)

// shopTagTables 某个域的三张表名与关联表实体列名。
// 表名 / 列名无法参数化，因此只允许来自下方白名单，杜绝拼接注入。
type shopTagTables struct {
	categories string
	tags       string
	links      string
	entityCol  string
}

var shopTagTableMap = map[ShopTagDomain]shopTagTables{
	ShopTagDomainRestaurant: {
		categories: "restaurant_tag_categories",
		tags:       "restaurant_tags",
		links:      "restaurant_tag_links",
		entityCol:  "restaurant_id",
	},
	ShopTagDomainDish: {
		categories: "dish_tag_categories",
		tags:       "dish_tags",
		links:      "dish_tag_links",
		entityCol:  "dish_id",
	},
}

func shopTagTablesFor(domain ShopTagDomain) (shopTagTables, error) {
	t, ok := shopTagTableMap[domain]
	if !ok {
		return shopTagTables{}, fmt.Errorf("未知的探店标签域: %q", domain)
	}
	return t, nil
}

// ShopTagRepo 探店标签字典（餐厅 / 菜品）数据访问。
// 可见性规则与菜谱标签一致：owner_id IS NULL 为全局预设，owner_id = 用户为私有自定义。
// 返回值复用 model.Tag / model.TagCategory 结构（纯 DTO，不与具体表绑定）。
type ShopTagRepo struct {
	pool *pgxpool.Pool
}

func NewShopTagRepo(pool *pgxpool.Pool) *ShopTagRepo {
	return &ShopTagRepo{pool: pool}
}

// ListVisibleCategories 返回该用户在某域可见的分类（全局预设 + 本人自定义）及其下标签
func (r *ShopTagRepo) ListVisibleCategories(ctx context.Context, userID string, domain ShopTagDomain) ([]*model.TagCategory, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT c.id, COALESCE(c.owner_id::text,''), c.name, c.color, c.sort_order, c.is_system,
		       COALESCE(t.id::text,''), COALESCE(t.category_id::text,''), COALESCE(t.owner_id::text,''),
		       COALESCE(t.name,''), COALESCE(t.mutex_group,''), COALESCE(t.sort_order,0), COALESCE(t.is_system,false)
		FROM `+tb.categories+` c
		LEFT JOIN `+tb.tags+` t ON t.category_id = c.id AND t.is_active
		WHERE c.owner_id IS NULL OR c.owner_id = $1::uuid
		ORDER BY c.is_system DESC, c.sort_order, c.created_at, c.id,
		         t.sort_order, t.created_at, t.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.TagCategory{}
	var cur *model.TagCategory
	for rows.Next() {
		var (
			catID, catOwner, catName, catColor       string
			catSort                                  int
			catSystem                                bool
			tagID, tagCategoryID, tagOwner, tagMutex string
			tagName                                  string
			tagSort                                  int
			tagSystem                                bool
		)
		if err := rows.Scan(&catID, &catOwner, &catName, &catColor, &catSort, &catSystem,
			&tagID, &tagCategoryID, &tagOwner, &tagName, &tagMutex, &tagSort, &tagSystem); err != nil {
			return nil, err
		}
		if cur == nil || cur.ID != catID {
			cur = &model.TagCategory{
				ID:        catID,
				OwnerID:   catOwner,
				Name:      catName,
				Color:     catColor,
				SortOrder: catSort,
				IsSystem:  catSystem,
				Tags:      []model.Tag{},
			}
			list = append(list, cur)
		}
		if tagID != "" {
			cur.Tags = append(cur.Tags, model.Tag{
				ID:         tagID,
				CategoryID: tagCategoryID,
				OwnerID:    tagOwner,
				Name:       tagName,
				MutexGroup: tagMutex,
				SortOrder:  tagSort,
				IsSystem:   tagSystem,
			})
		}
	}
	return list, rows.Err()
}

// ResolveVisibleIDs 返回给定标签 id 中「当前用户可见且启用」的集合
func (r *ShopTagRepo) ResolveVisibleIDs(ctx context.Context, userID string, domain ShopTagDomain, ids []string) (map[string]bool, error) {
	set := map[string]bool{}
	if len(ids) == 0 {
		return set, nil
	}
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text FROM `+tb.tags+`
		WHERE id = ANY($1::uuid[]) AND is_active
		  AND (owner_id IS NULL OR owner_id = $2::uuid)`, ids, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		set[id] = true
	}
	return set, rows.Err()
}

// GetTag 按 id 取标签（不含可见性过滤），不存在返回 nil
func (r *ShopTagRepo) GetTag(ctx context.Context, domain ShopTagDomain, id string) (*model.Tag, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	return scanTag(r.pool.QueryRow(ctx,
		`SELECT `+tagCols+` FROM `+tb.tags+` WHERE id = $1::uuid AND is_active`, id))
}

// GetCategory 按 id 取分类，不存在返回 nil
func (r *ShopTagRepo) GetCategory(ctx context.Context, domain ShopTagDomain, id string) (*model.TagCategory, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	return scanCategory(r.pool.QueryRow(ctx,
		`SELECT `+categoryCols+` FROM `+tb.categories+` WHERE id = $1::uuid`, id))
}

// CreateCategory 新建用户自定义分类
func (r *ShopTagRepo) CreateCategory(ctx context.Context, userID string, domain ShopTagDomain, name, color string, sortOrder int) (*model.TagCategory, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	return scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO `+tb.categories+` (owner_id, name, color, sort_order, is_system)
		 VALUES ($1::uuid, $2, $3, $4, false)
		 RETURNING `+categoryCols, userID, name, color, sortOrder))
}

// UpdateCategory 更新本人分类
func (r *ShopTagRepo) UpdateCategory(ctx context.Context, domain ShopTagDomain, id, name, color string, sortOrder int) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE `+tb.categories+` SET name=$2, color=$3, sort_order=$4 WHERE id=$1::uuid`,
		id, name, color, sortOrder)
	return err
}

// DeleteCategory 删除本人分类（级联删除其下标签与关联）
func (r *ShopTagRepo) DeleteCategory(ctx context.Context, domain ShopTagDomain, id string) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM `+tb.categories+` WHERE id = $1::uuid`, id)
	return err
}

// OwnCategoryNameTaken 该用户是否已有同名自定义分类（excludeID 非空时排除自身）
func (r *ShopTagRepo) OwnCategoryNameTaken(ctx context.Context, userID string, domain ShopTagDomain, name, excludeID string) (bool, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return false, err
	}
	var exists bool
	err = r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM `+tb.categories+`
			WHERE owner_id = $1::uuid AND name = $2
			  AND (NULLIF($3,'')::uuid IS NULL OR id <> NULLIF($3,'')::uuid))`,
		userID, name, excludeID).Scan(&exists)
	return exists, err
}

// OwnTagNameTaken 该用户是否已有同名自定义标签（excludeID 非空时排除自身）
func (r *ShopTagRepo) OwnTagNameTaken(ctx context.Context, userID string, domain ShopTagDomain, name, excludeID string) (bool, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return false, err
	}
	var exists bool
	err = r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM `+tb.tags+`
			WHERE owner_id = $1::uuid AND name = $2 AND is_active
			  AND (NULLIF($3,'')::uuid IS NULL OR id <> NULLIF($3,'')::uuid))`,
		userID, name, excludeID).Scan(&exists)
	return exists, err
}

// FindOrCreateCategory 按名称查找本人分类，不存在则创建
func (r *ShopTagRepo) FindOrCreateCategory(ctx context.Context, userID string, domain ShopTagDomain, name, color string) (*model.TagCategory, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	c, err := scanCategory(r.pool.QueryRow(ctx,
		`SELECT `+categoryCols+` FROM `+tb.categories+` WHERE owner_id = $1::uuid AND name = $2`, userID, name))
	if err != nil || c != nil {
		return c, err
	}
	return scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO `+tb.categories+` (owner_id, name, color, sort_order, is_system)
		 VALUES ($1::uuid, $2, $3, 900, false)
		 ON CONFLICT (owner_id, name) DO UPDATE SET updated_at = now()
		 RETURNING `+categoryCols, userID, name, color))
}

// CreateTag 新建用户自定义标签
func (r *ShopTagRepo) CreateTag(ctx context.Context, userID string, domain ShopTagDomain, categoryID, name, mutexGroup string, sortOrder int) (*model.Tag, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	return scanTag(r.pool.QueryRow(ctx,
		`INSERT INTO `+tb.tags+` (category_id, owner_id, name, mutex_group, sort_order, is_system)
		 VALUES ($1::uuid, $2::uuid, $3, NULLIF($4,''), $5, false)
		 RETURNING `+tagCols, categoryID, userID, name, mutexGroup, sortOrder))
}

// UpdateTag 更新本人标签
func (r *ShopTagRepo) UpdateTag(ctx context.Context, domain ShopTagDomain, id, name, mutexGroup string, sortOrder int) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE `+tb.tags+` SET name=$2, mutex_group=NULLIF($3,''), sort_order=$4 WHERE id=$1::uuid`,
		id, name, mutexGroup, sortOrder)
	return err
}

// DeleteTag 删除本人标签（级联删除实体关联）
func (r *ShopTagRepo) DeleteTag(ctx context.Context, domain ShopTagDomain, id string) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM `+tb.tags+` WHERE id = $1::uuid`, id)
	return err
}

// FindOrCreateTag 按名称查找本人标签，不存在则在指定分类下创建
func (r *ShopTagRepo) FindOrCreateTag(ctx context.Context, userID string, domain ShopTagDomain, categoryID, name string) (*model.Tag, error) {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}
	t, err := scanTag(r.pool.QueryRow(ctx,
		`SELECT `+tagCols+` FROM `+tb.tags+` WHERE owner_id = $1::uuid AND name = $2`, userID, name))
	if err != nil || t != nil {
		return t, err
	}
	return scanTag(r.pool.QueryRow(ctx,
		`INSERT INTO `+tb.tags+` (category_id, owner_id, name, sort_order, is_system)
		 VALUES ($1::uuid, $2::uuid, $3, 900, false)
		 ON CONFLICT (owner_id, name) DO UPDATE SET is_active = true, updated_at = now()
		 RETURNING `+tagCols, categoryID, userID, name))
}

// TagsFor 批量查询多个实体（餐厅 / 菜品）的关联标签 id，避免 N+1
func (r *ShopTagRepo) TagsFor(ctx context.Context, domain ShopTagDomain, entityIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(entityIDs) == 0 {
		return out, nil
	}
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT l.`+tb.entityCol+`::text, l.tag_id::text
		FROM `+tb.links+` l
		JOIN `+tb.tags+` t ON t.id = l.tag_id
		JOIN `+tb.categories+` c ON c.id = t.category_id
		WHERE l.`+tb.entityCol+` = ANY($1::uuid[])
		ORDER BY c.is_system DESC, c.sort_order, t.sort_order, t.created_at, t.id`, entityIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID, tagID string
		if err := rows.Scan(&entityID, &tagID); err != nil {
			return nil, err
		}
		out[entityID] = append(out[entityID], tagID)
	}
	return out, rows.Err()
}

// NamesByIDs 返回标签 id → 名称映射（用于导出等展示场景）
func (r *ShopTagRepo) NamesByIDs(ctx context.Context, domain ShopTagDomain, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `SELECT id::text, name FROM `+tb.tags+` WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = name
	}
	return out, rows.Err()
}

// ReplaceTags 整体替换实体的标签关联（非事务，供导入等简单场景使用）
func (r *ShopTagRepo) ReplaceTags(ctx context.Context, domain ShopTagDomain, entityID string, tagIDs []string) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	return replaceShopLinks(ctx, r.pool, tb, entityID, tagIDs)
}

// ReplaceTagsTx 在事务内整体替换实体的标签关联
func (r *ShopTagRepo) ReplaceTagsTx(ctx context.Context, tx pgx.Tx, domain ShopTagDomain, entityID string, tagIDs []string) error {
	tb, err := shopTagTablesFor(domain)
	if err != nil {
		return err
	}
	return replaceShopLinks(ctx, tx, tb, entityID, tagIDs)
}

func replaceShopLinks(ctx context.Context, q dbExecutor, tb shopTagTables, entityID string, tagIDs []string) error {
	if _, err := q.Exec(ctx, `DELETE FROM `+tb.links+` WHERE `+tb.entityCol+` = $1::uuid`, entityID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	_, err := q.Exec(ctx,
		`INSERT INTO `+tb.links+` (`+tb.entityCol+`, tag_id)
		 SELECT $1::uuid, u FROM unnest($2::uuid[]) AS u
		 ON CONFLICT DO NOTHING`, entityID, tagIDs)
	return err
}
