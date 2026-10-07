package tagging

import (
	"sort"
	"strings"

	"foodlog/server/internal/model"
)

// Input 自动标签输入
type Input struct {
	Name            string
	Description     string
	Ingredients     []string // 食材名称列表
	CookTimeMinutes int
}

// Generate 依据数据库中的规则推导食材级标签 id 列表。
//
// 规则按 sort_order 升序求值：
//  1. 先执行 ingredient_* / text_keyword / cook_time_max 等普通规则，命中即收集标签 id
//  2. 再执行 group_mutex 规则（依赖步骤 1 的命中结果），同一 tag_group 只取首个命中
//     —— 用于表达「荤菜优先、素菜其次」这类互斥语义
//
// 返回结果已去重，顺序与规则求值顺序一致（展示排序由上层按 sort_order 处理）。
func Generate(in Input, rules []*model.TagRule) []string {
	ordered := sortedRules(rules)

	matched := map[string]bool{}
	out := []string{}
	add := func(id string) {
		if id == "" || matched[id] {
			return
		}
		matched[id] = true
		out = append(out, id)
	}

	// 1) 普通规则
	for _, rule := range ordered {
		if rule.Type == model.RuleGroupMutex {
			continue
		}
		if evalRule(in, rule) {
			add(rule.TagID)
		}
	}

	// 2) 互斥组规则
	groupApplied := map[string]bool{}
	for _, rule := range ordered {
		if rule.Type != model.RuleGroupMutex {
			continue
		}
		if rule.TagGroup != "" && groupApplied[rule.TagGroup] {
			continue
		}
		if intersects(rule.MemberTagIDs, matched) {
			add(rule.TagID)
			if rule.TagGroup != "" {
				groupApplied[rule.TagGroup] = true
			}
		}
	}

	return out
}

// HasCoreTags 判断自动标签结果中是否包含「互斥组成员标签」，用于决定是否需要 AI 兜底。
// 对应旧实现里「没有命中荤素/食材类标签则认为词典识别不足」的语义。
func HasCoreTags(tagIDs []string, rules []*model.TagRule) bool {
	core := map[string]bool{}
	for _, rule := range rules {
		if rule.Type != model.RuleGroupMutex {
			continue
		}
		for _, id := range rule.MemberTagIDs {
			core[id] = true
		}
	}
	for _, id := range tagIDs {
		if core[id] {
			return true
		}
	}
	return false
}

func sortedRules(rules []*model.TagRule) []*model.TagRule {
	ordered := make([]*model.TagRule, len(rules))
	copy(ordered, rules)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].SortOrder < ordered[j].SortOrder
	})
	return ordered
}

func evalRule(in Input, rule *model.TagRule) bool {
	switch rule.Type {
	case model.RuleIngredientKeyword:
		for _, name := range in.Ingredients {
			if matchIngredient(name, rule) {
				return true
			}
		}
		return false
	case model.RuleIngredientTextKeyword:
		return containsAny(strings.Join(in.Ingredients, " "), rule.Keywords)
	case model.RuleTextKeyword:
		return containsAny(textFor(in, rule.MatchField), rule.Keywords)
	case model.RuleCookTimeMax:
		return rule.MaxMinutes > 0 && in.CookTimeMinutes > 0 && in.CookTimeMinutes <= rule.MaxMinutes
	default:
		return false
	}
}

// textFor 返回 text_keyword 规则需要匹配的文本
func textFor(in Input, field string) string {
	switch model.TagTextMatchField(field) {
	case model.TagFieldDescription:
		return in.Description
	case model.TagFieldNameDescription:
		return in.Name + " " + in.Description
	default:
		return in.Name
	}
}

// matchIngredient 判断食材名是否命中规则（先看排除词）
func matchIngredient(name string, rule *model.TagRule) bool {
	for _, ex := range rule.ExcludeKeywords {
		if ex != "" && strings.Contains(name, ex) {
			return false
		}
	}
	return containsAny(name, rule.Keywords)
}

// containsAny 判断文本是否包含任一关键词
func containsAny(s string, keywords []string) bool {
	if s == "" {
		return false
	}
	for _, kw := range keywords {
		if kw != "" && strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func intersects(ids []string, set map[string]bool) bool {
	for _, id := range ids {
		if set[id] {
			return true
		}
	}
	return false
}
