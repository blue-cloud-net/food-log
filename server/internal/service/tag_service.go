package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"foodlog/server/internal/ai"
	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/tagging"
)

// 标签相关业务错误
var (
	ErrTagInvalid        = errors.New("包含无效的标签")
	ErrTagDuplicate      = errors.New("同名标签或分类已存在")
	ErrSystemTagReadonly = errors.New("系统预设标签不可修改")
)

// 自定义分类兜底名称（TagSelector 自定义输入默认归入该分类）
const defaultCustomCategoryName = "自定义"

// tagCacheTTL 词表进程内缓存有效期。全局词表仅由 SQL 维护，
// 用户自定义词表在本服务内写入时会主动失效，30s 兜底足够。
const tagCacheTTL = 30 * time.Second

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type tagCacheEntry struct {
	categories []*model.TagCategory
	rules      []*model.TagRule
	expiresAt  time.Time
}

// TagService 标签字典服务：词表读取、自定义分类/标签维护、自动标签推导
type TagService struct {
	tagRepo    *repository.TagRepo
	aiProvider ai.Provider // 可为 nil（禁用 AI 兜底）

	mu    sync.RWMutex
	cache map[string]*tagCacheEntry
}

func NewTagService(tagRepo *repository.TagRepo, aiProvider ai.Provider) *TagService {
	return &TagService{tagRepo: tagRepo, aiProvider: aiProvider, cache: map[string]*tagCacheEntry{}}
}

// Categories 返回该用户可见的标签分类（全局预设 + 本人自定义）
func (s *TagService) Categories(ctx context.Context, userID string) ([]*model.TagCategory, error) {
	cats, _, err := s.dictionary(ctx, userID)
	return cats, err
}

// dictionary 读取（并缓存）用户可见的分类与启用规则
func (s *TagService) dictionary(ctx context.Context, userID string) ([]*model.TagCategory, []*model.TagRule, error) {
	s.mu.RLock()
	entry := s.cache[userID]
	s.mu.RUnlock()
	if entry != nil && time.Now().Before(entry.expiresAt) {
		return entry.categories, entry.rules, nil
	}

	cats, err := s.tagRepo.ListVisibleCategories(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	rules, err := s.tagRepo.ListActiveRules(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	s.cache[userID] = &tagCacheEntry{categories: cats, rules: rules, expiresAt: time.Now().Add(tagCacheTTL)}
	s.mu.Unlock()

	return cats, rules, nil
}

func (s *TagService) invalidate(userID string) {
	s.mu.Lock()
	delete(s.cache, userID)
	s.mu.Unlock()
}

// ===== 分类维护 =====

// CreateCategory 新建用户自定义分类
func (s *TagService) CreateCategory(ctx context.Context, userID, name, color string) (*model.TagCategory, error) {
	name, color, err := normalizeCategoryInput(name, color)
	if err != nil {
		return nil, err
	}
	taken, err := s.categoryNameTaken(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrTagDuplicate
	}

	c, err := s.tagRepo.CreateCategory(ctx, userID, name, color, 900)
	if err != nil {
		return nil, err
	}
	s.invalidate(userID)
	return c, nil
}

// UpdateCategory 更新本人自定义分类
func (s *TagService) UpdateCategory(ctx context.Context, userID, id, name, color string, sortOrder int) (*model.TagCategory, error) {
	c, err := s.tagRepo.GetCategory(ctx, id)
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
		taken, err := s.categoryNameTaken(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, ErrTagDuplicate
		}
	}
	if err := s.tagRepo.UpdateCategory(ctx, id, name, color, sortOrder); err != nil {
		return nil, err
	}
	s.invalidate(userID)
	return s.tagRepo.GetCategory(ctx, id)
}

// DeleteCategory 删除本人自定义分类（级联删除其下标签与菜谱关联）
func (s *TagService) DeleteCategory(ctx context.Context, userID, id string) error {
	c, err := s.tagRepo.GetCategory(ctx, id)
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
	if err := s.tagRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	s.invalidate(userID)
	return nil
}

// EnsureCustomCategory 取（或建）该用户的「自定义」兜底分类
func (s *TagService) EnsureCustomCategory(ctx context.Context, userID string) (*model.TagCategory, error) {
	c, err := s.tagRepo.FindOrCreateCategory(ctx, userID, defaultCustomCategoryName, "info")
	if err != nil {
		return nil, err
	}
	s.invalidate(userID)
	return c, nil
}

// ===== 标签维护 =====

// CreateTag 在指定分类下新建用户自定义标签
func (s *TagService) CreateTag(ctx context.Context, userID, categoryID, name, mutexGroup string) (*model.Tag, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return nil, err
	}
	cat, err := s.tagRepo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrNotFound
	}
	if cat.OwnerID != "" && cat.OwnerID != userID {
		return nil, ErrForbidden
	}
	if err := s.ensureUniqueTagName(ctx, userID, name, ""); err != nil {
		return nil, err
	}

	t, err := s.tagRepo.CreateTag(ctx, userID, categoryID, name, mutexGroup, 900)
	if err != nil {
		return nil, err
	}
	s.invalidate(userID)
	return t, nil
}

// UpdateTag 更新本人自定义标签
func (s *TagService) UpdateTag(ctx context.Context, userID, id, name, mutexGroup string, sortOrder int) (*model.Tag, error) {
	t, err := s.tagRepo.GetTag(ctx, id)
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
	if err := s.ensureUniqueTagName(ctx, userID, name, id); err != nil {
		return nil, err
	}
	if err := s.tagRepo.UpdateTag(ctx, id, name, mutexGroup, sortOrder); err != nil {
		return nil, err
	}
	s.invalidate(userID)
	return s.tagRepo.GetTag(ctx, id)
}

// DeleteTag 删除本人自定义标签（级联解除菜谱关联）
func (s *TagService) DeleteTag(ctx context.Context, userID, id string) error {
	t, err := s.tagRepo.GetTag(ctx, id)
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
	if err := s.tagRepo.DeleteTag(ctx, id); err != nil {
		return err
	}
	s.invalidate(userID)
	return nil
}

func (s *TagService) ensureUniqueTagName(ctx context.Context, userID, name, excludeID string) error {
	taken, err := s.tagRepo.OwnTagNameTaken(ctx, userID, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return ErrTagDuplicate
	}
	return nil
}

func (s *TagService) categoryNameTaken(ctx context.Context, userID, name string) (bool, error) {
	return s.tagRepo.OwnCategoryNameTaken(ctx, userID, name, "")
}

// ===== 菜谱标签解析与自动推导 =====

// ResolveManualTags 校验并归一化用户提交的菜谱级标签 id
func (s *TagService) ResolveManualTags(ctx context.Context, userID string, ids []string) ([]string, error) {
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

	visible, err := s.tagRepo.ResolveVisibleIDs(ctx, userID, out)
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

// ComputeIngredientTags 推导食材级标签 id：规则匹配 +（必要时）AI 兜底，
// 并剔除已手动选中的标签，避免同一标签在两处重复出现。
// 若手动已选中某互斥组（如荤素）的标签，自动结果中同组标签一并剔除，避免出现互相矛盾的展示。
func (s *TagService) ComputeIngredientTags(ctx context.Context, userID string, in tagging.Input, manualIDs []string) []string {
	cats, rules, err := s.dictionary(ctx, userID)
	if err != nil {
		return []string{}
	}

	ids := tagging.Generate(in, rules)
	if !tagging.HasCoreTags(ids, rules) {
		ids = append(ids, s.aiSuggest(ctx, cats, in, ids)...)
	}

	manual := make(map[string]bool, len(manualIDs))
	for _, id := range manualIDs {
		manual[id] = true
	}
	manualGroups, groupOf := mutexIndex(cats, manual)

	out := []string{}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || manual[id] || seen[id] {
			continue
		}
		if g := groupOf[id]; g != "" && manualGroups[g] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// mutexIndex 返回「手动已占用的互斥组集合」与「标签 id → 互斥组」映射
func mutexIndex(cats []*model.TagCategory, manual map[string]bool) (map[string]bool, map[string]string) {
	manualGroups := map[string]bool{}
	groupOf := map[string]string{}
	for _, c := range cats {
		for _, t := range c.Tags {
			if t.MutexGroup == "" {
				continue
			}
			groupOf[t.ID] = t.MutexGroup
			if manual[t.ID] {
				manualGroups[t.MutexGroup] = true
			}
		}
	}
	return manualGroups, groupOf
}

// aiSuggest 规则识别不足时让 AI 从词表中选择标签；AI 未配置或失败时静默降级
func (s *TagService) aiSuggest(ctx context.Context, cats []*model.TagCategory, in tagging.Input, current []string) []string {
	if s.aiProvider == nil {
		return nil
	}
	names, err := s.aiProvider.CompleteTags(ctx, buildTagPrompt(cats, in, current), "")
	if err != nil {
		return nil
	}
	return matchTagNames(cats, names)
}

// RecognizeTagNames 把 AI 图片识别返回的标签名转成 id；词表未收录的名称
// 会在该用户「自定义」分类下创建自定义标签（识别是显式用户操作，可接受词表增长）。
func (s *TagService) RecognizeTagNames(ctx context.Context, userID string, names []string) ([]string, error) {
	nameMap, err := s.nameIndex(ctx, userID)
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
			custom, err = s.EnsureCustomCategory(ctx, userID)
			if err != nil {
				return nil, err
			}
		}
		t, err := s.tagRepo.FindOrCreateTag(ctx, userID, custom.ID, name)
		if err != nil {
			return nil, err
		}
		out = append(out, t.ID)
	}
	return out, nil
}

// nameIndex 构建「标签名 → id」索引（同名时优先全局预设）
func (s *TagService) nameIndex(ctx context.Context, userID string) (map[string]string, error) {
	cats, err := s.tagRepo.ListVisibleCategories(ctx, userID)
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

func matchTagNames(cats []*model.TagCategory, names []string) []string {
	idx := map[string]string{}
	for _, c := range cats {
		for _, t := range c.Tags {
			if _, exists := idx[t.Name]; !exists || t.OwnerID == "" {
				idx[t.Name] = t.ID
			}
		}
	}
	out := []string{}
	for _, raw := range names {
		if id, ok := idx[strings.TrimSpace(raw)]; ok {
			out = append(out, id)
		}
	}
	return out
}

// ===== 辅助 =====

func normalizeTagName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: 标签名不能为空", ErrTagInvalid)
	}
	if len([]rune(name)) > 50 {
		return "", fmt.Errorf("%w: 标签名不能超过 50 个字符", ErrTagInvalid)
	}
	return name, nil
}

func normalizeCategoryInput(name, color string) (string, string, error) {
	name, err := normalizeTagName(name)
	if err != nil {
		return "", "", err
	}
	color = strings.TrimSpace(color)
	switch color {
	case "primary", "success", "warning", "danger", "info":
	case "":
		color = "info"
	default:
		return "", "", fmt.Errorf("%w: 颜色仅支持 primary/success/warning/danger/info", ErrTagInvalid)
	}
	return name, color, nil
}

// buildTagPrompt 用数据库词表构造 AI 兜底提示词
func buildTagPrompt(cats []*model.TagCategory, in tagging.Input, current []string) string {
	var sb strings.Builder
	sb.WriteString("你是一个美食标签助手。请根据菜谱信息，从以下预设标签中选择合适的标签（数量 2~6 个，只输出 JSON 字符串数组，不要输出其它内容）：\n")
	for _, c := range cats {
		if len(c.Tags) == 0 {
			continue
		}
		names := make([]string, 0, len(c.Tags))
		for _, t := range c.Tags {
			names = append(names, t.Name)
		}
		sb.WriteString(fmt.Sprintf("%s：%s\n", c.Name, strings.Join(names, "、")))
	}
	sb.WriteString("\n菜名：" + in.Name + "\n")
	if in.Description != "" {
		sb.WriteString("描述：" + in.Description + "\n")
	}
	if len(in.Ingredients) > 0 {
		sb.WriteString("食材：" + strings.Join(in.Ingredients, "、") + "\n")
	}
	if in.CookTimeMinutes > 0 {
		sb.WriteString(fmt.Sprintf("耗时：%d 分钟\n", in.CookTimeMinutes))
	}
	if len(current) > 0 {
		sb.WriteString("已识别标签数量：" + fmt.Sprint(len(current)) + "\n")
	}
	return sb.String()
}
