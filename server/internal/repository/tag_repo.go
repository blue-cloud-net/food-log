package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
)

// TagRepo 标签字典（分类 / 标签 / 自动规则 / 菜谱关联）数据访问。
// 可见性规则：owner_id IS NULL 为全局预设，owner_id = 用户为私有自定义。
type TagRepo struct {
	pool *pgxpool.Pool
}

func NewTagRepo(pool *pgxpool.Pool) *TagRepo {
	return &TagRepo{pool: pool}
}

const tagCols = `id, category_id, COALESCE(owner_id::text,''), name, COALESCE(mutex_group,''), sort_order, is_system`

const categoryCols = `id, COALESCE(owner_id::text,''), name, color, sort_order, is_system`

func scanTag(row pgx.Row) (*model.Tag, error) {
	var t model.Tag
	err := row.Scan(&t.ID, &t.CategoryID, &t.OwnerID, &t.Name, &t.MutexGroup, &t.SortOrder, &t.IsSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func scanCategory(row pgx.Row) (*model.TagCategory, error) {
	var c model.TagCategory
	err := row.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Color, &c.SortOrder, &c.IsSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Tags = []model.Tag{}
	return &c, nil
}

// ListVisibleCategories 返回该用户可见的分类（全局预设 + 本人自定义）及其下标签
func (r *TagRepo) ListVisibleCategories(ctx context.Context, userID string) ([]*model.TagCategory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, COALESCE(c.owner_id::text,''), c.name, c.color, c.sort_order, c.is_system,
		       COALESCE(t.id::text,''), COALESCE(t.category_id::text,''), COALESCE(t.owner_id::text,''),
		       COALESCE(t.name,''), COALESCE(t.mutex_group,''), COALESCE(t.sort_order,0), COALESCE(t.is_system,false)
		FROM tag_categories c
		LEFT JOIN tags t ON t.category_id = c.id AND t.is_active
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
			catID, catOwner, catName, catColor      string
			catSort                                 int
			catSystem                               bool
			tagID, tagCategoryID, tagOwner, tagName string
			tagMutex                                string
			tagSort                                 int
			tagSystem                               bool
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

// ListActiveRules 返回该用户可见的启用中自动标签规则（按 sort_order 升序）
func (r *TagRepo) ListActiveRules(ctx context.Context, userID string) ([]*model.TagRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tag_id, rule_type, COALESCE(match_field,''), keywords, exclude_keywords,
		       COALESCE(max_minutes,0), COALESCE(tag_group,''), member_tag_ids::text[], sort_order
		FROM tag_rules
		WHERE is_active AND (owner_id IS NULL OR owner_id = $1::uuid)
		ORDER BY sort_order, created_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*model.TagRule{}
	for rows.Next() {
		var rule model.TagRule
		if err := rows.Scan(&rule.ID, &rule.TagID, &rule.Type, &rule.MatchField,
			&rule.Keywords, &rule.ExcludeKeywords, &rule.MaxMinutes, &rule.TagGroup,
			&rule.MemberTagIDs, &rule.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, &rule)
	}
	return list, rows.Err()
}

// ResolveVisibleIDs 返回给定标签 id 中「当前用户可见且启用」的集合
func (r *TagRepo) ResolveVisibleIDs(ctx context.Context, userID string, ids []string) (map[string]bool, error) {
	set := map[string]bool{}
	if len(ids) == 0 {
		return set, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text FROM tags
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
func (r *TagRepo) GetTag(ctx context.Context, id string) (*model.Tag, error) {
	return scanTag(r.pool.QueryRow(ctx, `SELECT `+tagCols+` FROM tags WHERE id = $1::uuid AND is_active`, id))
}

// GetCategory 按 id 取分类，不存在返回 nil
func (r *TagRepo) GetCategory(ctx context.Context, id string) (*model.TagCategory, error) {
	return scanCategory(r.pool.QueryRow(ctx, `SELECT `+categoryCols+` FROM tag_categories WHERE id = $1::uuid`, id))
}

// CreateCategory 新建用户自定义分类
func (r *TagRepo) CreateCategory(ctx context.Context, userID, name, color string, sortOrder int) (*model.TagCategory, error) {
	return scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO tag_categories (owner_id, name, color, sort_order, is_system)
		 VALUES ($1::uuid, $2, $3, $4, false)
		 RETURNING `+categoryCols, userID, name, color, sortOrder))
}

// UpdateCategory 更新本人分类
func (r *TagRepo) UpdateCategory(ctx context.Context, id, name, color string, sortOrder int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tag_categories SET name=$2, color=$3, sort_order=$4 WHERE id=$1::uuid`,
		id, name, color, sortOrder)
	return err
}

// DeleteCategory 删除本人分类（级联删除其下标签与关联）
func (r *TagRepo) DeleteCategory(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tag_categories WHERE id = $1::uuid`, id)
	return err
}

// OwnCategoryNameTaken 该用户是否已有同名自定义分类（excludeID 非空时排除自身）
func (r *TagRepo) OwnCategoryNameTaken(ctx context.Context, userID, name, excludeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tag_categories
			WHERE owner_id = $1::uuid AND name = $2
			  AND (NULLIF($3,'')::uuid IS NULL OR id <> NULLIF($3,'')::uuid))`,
		userID, name, excludeID).Scan(&exists)
	return exists, err
}

// OwnTagNameTaken 该用户是否已有同名自定义标签（excludeID 非空时排除自身）
func (r *TagRepo) OwnTagNameTaken(ctx context.Context, userID, name, excludeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tags
			WHERE owner_id = $1::uuid AND name = $2 AND is_active
			  AND (NULLIF($3,'')::uuid IS NULL OR id <> NULLIF($3,'')::uuid))`,
		userID, name, excludeID).Scan(&exists)
	return exists, err
}

// FindOrCreateCategory 按名称查找本人分类，不存在则创建
func (r *TagRepo) FindOrCreateCategory(ctx context.Context, userID, name, color string) (*model.TagCategory, error) {
	c, err := scanCategory(r.pool.QueryRow(ctx,
		`SELECT `+categoryCols+` FROM tag_categories WHERE owner_id = $1::uuid AND name = $2`, userID, name))
	if err != nil || c != nil {
		return c, err
	}
	return scanCategory(r.pool.QueryRow(ctx,
		`INSERT INTO tag_categories (owner_id, name, color, sort_order, is_system)
		 VALUES ($1::uuid, $2, $3, 900, false)
		 ON CONFLICT (owner_id, name) DO UPDATE SET updated_at = now()
		 RETURNING `+categoryCols, userID, name, color))
}

// CreateTag 新建用户自定义标签
func (r *TagRepo) CreateTag(ctx context.Context, userID, categoryID, name, mutexGroup string, sortOrder int) (*model.Tag, error) {
	return scanTag(r.pool.QueryRow(ctx,
		`INSERT INTO tags (category_id, owner_id, name, mutex_group, sort_order, is_system)
		 VALUES ($1::uuid, $2::uuid, $3, NULLIF($4,''), $5, false)
		 RETURNING `+tagCols, categoryID, userID, name, mutexGroup, sortOrder))
}

// UpdateTag 更新本人标签
func (r *TagRepo) UpdateTag(ctx context.Context, id, name, mutexGroup string, sortOrder int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tags SET name=$2, mutex_group=NULLIF($3,''), sort_order=$4 WHERE id=$1::uuid`,
		id, name, mutexGroup, sortOrder)
	return err
}

// DeleteTag 删除本人标签（级联删除菜谱关联）
func (r *TagRepo) DeleteTag(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id = $1::uuid`, id)
	return err
}

// FindOrCreateTag 按名称查找本人标签，不存在则在指定分类下创建
func (r *TagRepo) FindOrCreateTag(ctx context.Context, userID, categoryID, name string) (*model.Tag, error) {
	t, err := scanTag(r.pool.QueryRow(ctx,
		`SELECT `+tagCols+` FROM tags WHERE owner_id = $1::uuid AND name = $2`, userID, name))
	if err != nil || t != nil {
		return t, err
	}
	return scanTag(r.pool.QueryRow(ctx,
		`INSERT INTO tags (category_id, owner_id, name, sort_order, is_system)
		 VALUES ($1::uuid, $2::uuid, $3, 900, false)
		 ON CONFLICT (owner_id, name) DO UPDATE SET is_active = true, updated_at = now()
		 RETURNING `+tagCols, categoryID, userID, name))
}

// TagsForRecipes 批量查询多个菜谱的关联标签 id（菜谱级 / 食材级），避免 N+1
func (r *TagRepo) TagsForRecipes(ctx context.Context, recipeIDs []string) (map[string][]string, map[string][]string, error) {
	manual := map[string][]string{}
	ingredient := map[string][]string{}
	if len(recipeIDs) == 0 {
		return manual, ingredient, nil
	}

	const orderBy = ` ORDER BY c.is_system DESC, c.sort_order, t.sort_order, t.created_at, t.id`
	if err := r.collectLinks(ctx, `
		SELECT rt.recipe_id::text, rt.tag_id::text
		FROM recipe_tags rt
		JOIN tags t ON t.id = rt.tag_id
		JOIN tag_categories c ON c.id = t.category_id
		WHERE rt.recipe_id = ANY($1::uuid[])`+orderBy, recipeIDs, manual); err != nil {
		return nil, nil, err
	}
	if err := r.collectLinks(ctx, `
		SELECT rit.recipe_id::text, rit.tag_id::text
		FROM recipe_ingredient_tags rit
		JOIN tags t ON t.id = rit.tag_id
		JOIN tag_categories c ON c.id = t.category_id
		WHERE rit.recipe_id = ANY($1::uuid[])`+orderBy, recipeIDs, ingredient); err != nil {
		return nil, nil, err
	}
	return manual, ingredient, nil
}

func (r *TagRepo) collectLinks(ctx context.Context, sql string, recipeIDs []string, out map[string][]string) error {
	rows, err := r.pool.Query(ctx, sql, recipeIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var recipeID, tagID string
		if err := rows.Scan(&recipeID, &tagID); err != nil {
			return err
		}
		out[recipeID] = append(out[recipeID], tagID)
	}
	return rows.Err()
}

// NameByIDs 返回标签 id → 名称映射（用于导出等展示场景）
func (r *TagRepo) NamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text, name FROM tags WHERE id = ANY($1::uuid[])`, ids)
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

// ReplaceRecipeTags 整体替换菜谱级标签关联（非事务，供导入等简单场景使用）
func (r *TagRepo) ReplaceRecipeTags(ctx context.Context, recipeID string, tagIDs []string) error {
	return replaceLinks(ctx, r.pool, "recipe_tags", recipeID, tagIDs)
}

// ReplaceRecipeIngredientTags 整体替换食材级标签关联（非事务）
func (r *TagRepo) ReplaceRecipeIngredientTags(ctx context.Context, recipeID string, tagIDs []string) error {
	return replaceLinks(ctx, r.pool, "recipe_ingredient_tags", recipeID, tagIDs)
}

// ReplaceRecipeTagsTx 在事务内整体替换菜谱级标签关联
func (r *TagRepo) ReplaceRecipeTagsTx(ctx context.Context, tx pgx.Tx, recipeID string, tagIDs []string) error {
	return replaceLinks(ctx, tx, "recipe_tags", recipeID, tagIDs)
}

// ReplaceRecipeIngredientTagsTx 在事务内整体替换食材级标签关联
func (r *TagRepo) ReplaceRecipeIngredientTagsTx(ctx context.Context, tx pgx.Tx, recipeID string, tagIDs []string) error {
	return replaceLinks(ctx, tx, "recipe_ingredient_tags", recipeID, tagIDs)
}

func replaceLinks(ctx context.Context, q dbExecutor, table, recipeID string, tagIDs []string) error {
	if _, err := q.Exec(ctx, `DELETE FROM `+table+` WHERE recipe_id = $1::uuid`, recipeID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	_, err := q.Exec(ctx,
		`INSERT INTO `+table+` (recipe_id, tag_id)
		 SELECT $1::uuid, u FROM unnest($2::uuid[]) AS u
		 ON CONFLICT DO NOTHING`, recipeID, tagIDs)
	return err
}
