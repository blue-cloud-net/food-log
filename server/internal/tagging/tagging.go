package tagging

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Input 自动标签输入
type Input struct {
	Name            string
	Description     string
	Ingredients     []string // 食材名称列表
	CookTimeMinutes int
}

// Completer 词典未命中时由外部提供的 AI 补全能力（可为 nil）
type Completer interface {
	// CompleteTags 让 AI 从预设词表中选择标签并以字符串数组返回
	CompleteTags(ctx context.Context, prompt string) ([]string, error)
}

// Generate 生成自动标签并与手动标签合并去重。
// completer 为 nil 或调用失败时静默降级（仅返回词典结果）。
func Generate(ctx context.Context, in Input, manualTags []string, completer Completer) []string {
	auto := matchDictionary(in)

	// 词典对食材识别不足时，尝试 AI 兜底补全
	if completer != nil && needAI(auto) {
		if aiTags, err := completer.CompleteTags(ctx, buildPrompt(in, auto)); err == nil {
			auto = append(auto, aiTags...)
		}
	}

	return mergeTags(auto, manualTags)
}

// needAI 当没有命中任何荤素/食材类标签时，认为词典识别不足，需要 AI 兜底
func needAI(auto []string) bool {
	for _, t := range auto {
		if meatTags[t] || vegTags[t] {
			return false
		}
	}
	return true
}

// matchDictionary 纯词典匹配：食材关键词 + 荤素判定 + 场景/口味规则
func matchDictionary(in Input) []string {
	var tags []string
	add := func(t string) {
		for _, e := range tags {
			if e == t {
				return
			}
		}
		tags = append(tags, t)
	}

	// 1. 食材关键词匹配
	hasMeat := false
	hasVeg := false
	for _, name := range in.Ingredients {
		for _, rule := range ingredientRules {
			if !matched(name, rule) {
				continue
			}
			add(rule.Tag)
			if meatTags[rule.Tag] {
				hasMeat = true
			}
			if vegTags[rule.Tag] {
				hasVeg = true
			}
		}
	}

	// 2. 荤素判定（互斥；蛋归荤）
	if hasMeat {
		add("荤菜")
	} else if hasVeg {
		add("素菜")
	}

	// 3. 场景/口味规则
	if in.CookTimeMinutes > 0 && in.CookTimeMinutes <= 15 {
		add("快手")
	}
	title := in.Name + " " + in.Description
	if containsAny(title, soupKeywords) {
		add("汤")
	}
	if containsAny(in.Name, coldKeywords) {
		add("凉菜")
	}
	if containsAny(title, dessertKeywords) {
		add("甜点")
	}
	if containsAny(strings.Join(in.Ingredients, " "), spicyKeywords) {
		add("辣")
	}
	for _, name := range in.Ingredients {
		if containsAny(name, noodleKeywords) {
			add("面食")
			break
		}
	}

	return tags
}

// mergeTags 合并自动与手动标签并去重
func mergeTags(auto, manual []string) []string {
	seen := map[string]bool{}
	var out []string
	appendUnique := func(t string) {
		if t == "" || seen[t] {
			return
		}
		seen[t] = true
		out = append(out, t)
	}
	// 优先按词表顺序输出自动标签，保持展示稳定
	sortTags(auto)
	for _, t := range auto {
		appendUnique(t)
	}
	for _, t := range manual {
		appendUnique(strings.TrimSpace(t))
	}
	return out
}

// sortTags 按词表顺序排序（未收录的放最后）
func sortTags(tags []string) {
	order := make(map[string]int)
	idx := 0
	for _, cat := range TagCategories {
		for _, t := range cat.Tags {
			if _, ok := order[t]; !ok {
				order[t] = idx
				idx++
			}
		}
	}
	sort.SliceStable(tags, func(i, j int) bool {
		oi, iok := order[tags[i]]
		oj, jok := order[tags[j]]
		switch {
		case iok && jok:
			return oi < oj
		case iok:
			return true
		case jok:
			return false
		default:
			return tags[i] < tags[j]
		}
	})
}

// buildPrompt 构造 AI 兜底提示词
func buildPrompt(in Input, current []string) string {
	var sb strings.Builder
	sb.WriteString("你是一个美食标签助手。请根据菜谱信息，从以下预设标签中选择合适的标签（数量 2~6 个，只输出 JSON 字符串数组，不要输出其它内容）：\n")
	for _, cat := range TagCategories {
		sb.WriteString(fmt.Sprintf("%s：%s\n", cat.Name, strings.Join(cat.Tags, "、")))
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
		sb.WriteString("已识别标签（可补充，勿重复）：" + strings.Join(current, "、") + "\n")
	}
	return sb.String()
}
