package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// shopTagCacheTTL 探店词表进程内缓存有效期。探店标签没有自动规则，
// 词表只可能由本服务的写接口或 SQL 变更，30s 兜底足够。
const shopTagCacheTTL = 30 * time.Second

type shopTagCacheEntry struct {
	categories []*model.TagCategory
	expiresAt  time.Time
}

// ShopTagService 探店标签字典服务。
// 餐厅与菜品各有一套独立字典（repository.ShopTagDomain 区分），与菜谱标签体系完全隔离；
// 探店标签为纯手动（系统预设 + 用户自定义），没有 tag_rules 自动推导与 AI 兜底。
type ShopTagService struct {
	repo *repository.ShopTagRepo

	mu    sync.RWMutex
	cache map[string]*shopTagCacheEntry // key = userID + "|" + domain
}

func NewShopTagService(repo *repository.ShopTagRepo) *ShopTagService {
	return &ShopTagService{repo: repo, cache: map[string]*shopTagCacheEntry{}}
}

func shopTagCacheKey(userID string, domain repository.ShopTagDomain) string {
	return userID + "|" + string(domain)
}

// Categories 返回该用户在某域可见的标签分类（全局预设 + 本人自定义）
func (s *ShopTagService) Categories(ctx context.Context, userID string, domain repository.ShopTagDomain) ([]*model.TagCategory, error) {
	key := shopTagCacheKey(userID, domain)

	s.mu.RLock()
	entry := s.cache[key]
	s.mu.RUnlock()
	if entry != nil && time.Now().Before(entry.expiresAt) {
		return entry.categories, nil
	}

	cats, err := s.repo.ListVisibleCategories(ctx, userID, domain)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache[key] = &shopTagCacheEntry{categories: cats, expiresAt: time.Now().Add(shopTagCacheTTL)}
	s.mu.Unlock()

	return cats, nil
}

func (s *ShopTagService) invalidate(userID string, domain repository.ShopTagDomain) {
	s.mu.Lock()
	delete(s.cache, shopTagCacheKey(userID, domain))
	s.mu.Unlock()
}

// ===== 分类维护 =====

// CreateCategory 新建用户自定义分类
func (s *ShopTagService) CreateCategory(ctx context.Context, userID string, domain repository.ShopTagDomain, name, color string) (*model.TagCategory, error) {
	name, color, err := normalizeCategoryInput(name, color)
	if err != nil {
		return nil, err
	}
	taken, err := s.categoryNameTaken(ctx, userID, domain, name)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrTagDuplicate
	}

	c, err := s.repo.CreateCategory(ctx, userID, domain, name, color, 900)
	if err != nil {
		return nil, err
	}
	s.invalidate(userID, domain)
	return c, nil
}

// UpdateCategory 更新本人自定义分类
func (s *ShopTagService) UpdateCategory(ctx context.Context, userID string, domain repository.ShopTagDomain, id, name, color string, sortOrder int) (*model.TagCategory, error) {
	c, err := s.repo.GetCategory(ctx, domain, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if c.IsSystem {
		return nil, ErrSystemTagReadonly
	}
	if c.OwnerID != userID {
		return nil, ErrForbidden
	}

	name, color, err = normalizeCategoryInput(name, color)
	if err != nil {
		return nil, err
	}
	if name != c.Name {
		taken, err := s.categoryNameTaken(ctx, userID, domain, name)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, ErrTagDuplicate
		}
	}
	if err := s.repo.UpdateCategory(ctx, domain, id, name, color, sortOrder); err != nil {
		return nil, err
	}
	s.invalidate(userID, domain)
	return s.repo.GetCategory(ctx, domain, id)
}

// DeleteCategory 删除本人自定义分类（级联删除其下标签与实体关联）
func (s *ShopTagService) DeleteCategory(ctx context.Context, userID string, domain repository.ShopTagDomain, id string) error {
	c, err := s.repo.GetCategory(ctx, domain, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrNotFound
	}
	if c.IsSystem {
		return ErrSystemTagReadonly
	}
	if c.OwnerID != userID {
		return ErrForbidden
	}
	if err := s.repo.DeleteCategory(ctx, domain, id); err != nil {
		return err
	}
	s.invalidate(userID, domain)
	return nil
}

// EnsureCustomCategory 取（或建）该用户在某域的「自定义」兜底分类
func (s *ShopTagService) EnsureCustomCategory(ctx context.Context, userID string, domain repository.ShopTagDomain) (*model.TagCategory, error) {
	c, err := s.repo.FindOrCreateCategory(ctx, userID, domain, defaultCustomCategoryName, "info")
	if err != nil {
		return nil, err
	}
	s.invalidate(userID, domain)
	return c, nil
}

// ===== 标签维护 =====

// CreateTag 在指定分类下新建用户自定义标签
func (s *ShopTagService) CreateTag(ctx context.Context, userID string, domain repository.ShopTagDomain, categoryID, name, mutexGroup string) (*model.Tag, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return nil, err
	}
	cat, err := s.repo.GetCategory(ctx, domain, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrNotFound
	}
	if cat.OwnerID != "" && cat.OwnerID != userID {
		return nil, ErrForbidden
	}
	if err := s.ensureUniqueTagName(ctx, userID, domain, name, ""); err != nil {
		return nil, err
	}

	t, err := s.repo.CreateTag(ctx, userID, domain, categoryID, name, mutexGroup, 900)
	if err != nil {
		return nil, err
	}
	s.invalidate(userID, domain)
	return t, nil
}

// UpdateTag 更新本人自定义标签
func (s *ShopTagService) UpdateTag(ctx context.Context, userID string, domain repository.ShopTagDomain, id, name, mutexGroup string, sortOrder int) (*model.Tag, error) {
	t, err := s.repo.GetTag(ctx, domain, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrNotFound
	}
	if t.IsSystem {
		return nil, ErrSystemTagReadonly
	}
	if t.OwnerID != userID {
		return nil, ErrForbidden
	}

	name, err = normalizeTagName(name)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUniqueTagName(ctx, userID, domain, name, id); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateTag(ctx, domain, id, name, mutexGroup, sortOrder); err != nil {
		return nil, err
	}
	s.invalidate(userID, domain)
	return s.repo.GetTag(ctx, domain, id)
}

// DeleteTag 删除本人自定义标签（级联解除实体关联）
func (s *ShopTagService) DeleteTag(ctx context.Context, userID string, domain repository.ShopTagDomain, id string) error {
	t, err := s.repo.GetTag(ctx, domain, id)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrNotFound
	}
	if t.IsSystem {
		return ErrSystemTagReadonly
	}
	if t.OwnerID != userID {
		return ErrForbidden
	}
	if err := s.repo.DeleteTag(ctx, domain, id); err != nil {
		return err
	}
	s.invalidate(userID, domain)
	return nil
}

func (s *ShopTagService) ensureUniqueTagName(ctx context.Context, userID string, domain repository.ShopTagDomain, name, excludeID string) error {
	taken, err := s.repo.OwnTagNameTaken(ctx, userID, domain, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return ErrTagDuplicate
	}
	return nil
}

func (s *ShopTagService) categoryNameTaken(ctx context.Context, userID string, domain repository.ShopTagDomain, name string) (bool, error) {
	return s.repo.OwnCategoryNameTaken(ctx, userID, domain, name, "")
}

// ===== 标签 id 解析 =====

// ResolveManualTags 校验并归一化用户提交的标签 id（去空、去重、可见性校验）
func (s *ShopTagService) ResolveManualTags(ctx context.Context, userID string, domain repository.ShopTagDomain, ids []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	invalid := []string{}
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		if !uuidRe.MatchString(id) {
			invalid = append(invalid, id)
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrTagInvalid, strings.Join(invalid, ", "))
	}
	if len(out) == 0 {
		return out, nil
	}

	visible, err := s.repo.ResolveVisibleIDs(ctx, userID, domain, out)
	if err != nil {
		return nil, err
	}
	for _, id := range out {
		if !visible[id] {
			invalid = append(invalid, id)
		}
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrTagInvalid, strings.Join(invalid, ", "))
	}
	return out, nil
}

// RecognizeTagNames 把标签名转成 id；词表未收录的名称会在该用户「自定义」
// 分类下创建自定义标签（用于数据导入，与菜谱标签的处理一致）。
func (s *ShopTagService) RecognizeTagNames(ctx context.Context, userID string, domain repository.ShopTagDomain, names []string) ([]string, error) {
	nameMap, err := s.nameIndex(ctx, userID, domain)
	if err != nil {
		return nil, err
	}

	var custom *model.TagCategory
	out := []string{}
	seen := map[string]bool{}
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		if id, ok := nameMap[name]; ok {
			out = append(out, id)
			continue
		}
		if custom == nil {
			custom, err = s.EnsureCustomCategory(ctx, userID, domain)
			if err != nil {
				return nil, err
			}
		}
		t, err := s.repo.FindOrCreateTag(ctx, userID, domain, custom.ID, name)
		if err != nil {
			return nil, err
		}
		out = append(out, t.ID)
	}
	return out, nil
}

// nameIndex 构建「标签名 → id」索引（同名时优先全局预设）
func (s *ShopTagService) nameIndex(ctx context.Context, userID string, domain repository.ShopTagDomain) (map[string]string, error) {
	cats, err := s.repo.ListVisibleCategories(ctx, userID, domain)
	if err != nil {
		return nil, err
	}
	idx := map[string]string{}
	for _, c := range cats {
		for _, t := range c.Tags {
			if _, exists := idx[t.Name]; !exists || t.OwnerID == "" {
				idx[t.Name] = t.ID
			}
		}
	}
	return idx, nil
}
