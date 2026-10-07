package tagging

import (
	"testing"

	"foodlog/server/internal/model"
)

// testRules 内存规则，与 004_tags.sql 种子结构一致（id 用可读占位符）
func testRules() []*model.TagRule {
	return []*model.TagRule{
		{ID: "r1", TagID: "beef", Type: model.RuleIngredientKeyword,
			Keywords: []string{"牛肉", "牛腩", "牛腱"}, SortOrder: 10},
		{ID: "r2", TagID: "egg", Type: model.RuleIngredientKeyword,
			Keywords: []string{"鸡蛋", "蛋"}, SortOrder: 20},
		{ID: "r3", TagID: "fish", Type: model.RuleIngredientKeyword,
			Keywords: []string{"鱼", "鱼片"}, ExcludeKeywords: []string{"鱼香"}, SortOrder: 30},
		{ID: "r4", TagID: "tofu", Type: model.RuleIngredientKeyword,
			Keywords: []string{"豆腐", "豆干"}, SortOrder: 40},
		{ID: "r5", TagID: "vegetable", Type: model.RuleIngredientKeyword,
			Keywords: []string{"青菜", "土豆", "番茄", "茄子"}, SortOrder: 50},
		{ID: "r6", TagID: "spicy", Type: model.RuleIngredientTextKeyword,
			Keywords: []string{"豆瓣", "辣椒", "花椒"}, SortOrder: 60},
		{ID: "r7", TagID: "noodle", Type: model.RuleIngredientTextKeyword,
			Keywords: []string{"面条", "饺子", "饼"}, SortOrder: 70},
		{ID: "r8", TagID: "soup", Type: model.RuleTextKeyword, MatchField: "name_description",
			Keywords: []string{"汤"}, SortOrder: 80},
		{ID: "r9", TagID: "cold", Type: model.RuleTextKeyword, MatchField: "name",
			Keywords: []string{"凉拌", "凉菜"}, SortOrder: 90},
		{ID: "r10", TagID: "quick", Type: model.RuleCookTimeMax,
			MaxMinutes: 15, SortOrder: 100},
		{ID: "r11", TagID: "diet-meat", Type: model.RuleGroupMutex, TagGroup: "diet",
			MemberTagIDs: []string{"beef", "egg", "fish"}, SortOrder: 210},
		{ID: "r12", TagID: "diet-veg", Type: model.RuleGroupMutex, TagGroup: "diet",
			MemberTagIDs: []string{"tofu", "vegetable"}, SortOrder: 220},
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestGenerate_牛肉土豆判为荤菜(t *testing.T) {
	got := Generate(Input{
		Name:            "土豆炖牛肉",
		Ingredients:     []string{"牛肉", "土豆"},
		CookTimeMinutes: 30,
	}, testRules())

	for _, want := range []string{"diet-meat", "beef", "vegetable"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "diet-veg") {
		t.Errorf("有肉类时不应判为素菜: %v", got)
	}
}

func TestGenerate_纯豆制品判为素菜(t *testing.T) {
	got := Generate(Input{
		Name:        "麻婆豆腐",
		Ingredients: []string{"嫩豆腐", "豆瓣酱"},
	}, testRules())

	for _, want := range []string{"diet-veg", "tofu", "spicy"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "diet-meat") {
		t.Errorf("纯素菜不应判为荤菜: %v", got)
	}
}

func TestGenerate_蛋归为荤菜(t *testing.T) {
	got := Generate(Input{
		Name:        "番茄炒蛋",
		Ingredients: []string{"番茄", "鸡蛋"},
	}, testRules())

	for _, want := range []string{"diet-meat", "egg", "vegetable"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "diet-veg") {
		t.Errorf("含蛋菜不应判为素菜: %v", got)
	}
}

func TestGenerate_快手按耗时(t *testing.T) {
	got := Generate(Input{
		Name:            "清炒时蔬",
		Ingredients:     []string{"青菜"},
		CookTimeMinutes: 10,
	}, testRules())
	if !contains(got, "quick") {
		t.Errorf("10 分钟应命中快手: %v", got)
	}

	got = Generate(Input{
		Name:            "慢炖牛腩",
		Ingredients:     []string{"牛腩"},
		CookTimeMinutes: 90,
	}, testRules())
	if contains(got, "quick") {
		t.Errorf("90 分钟不应命中快手: %v", got)
	}
}

func TestGenerate_排除词生效(t *testing.T) {
	got := Generate(Input{
		Name:        "鱼香茄子",
		Ingredients: []string{"鱼香茄子", "茄子"},
	}, testRules())
	if contains(got, "fish") {
		t.Errorf("鱼香茄子不应命中鱼标签: %v", got)
	}
	if !contains(got, "vegetable") {
		t.Errorf("茄子应命中蔬菜标签: %v", got)
	}
}

func TestGenerate_结果不重复(t *testing.T) {
	got := Generate(Input{
		Name:        "牛肉面",
		Ingredients: []string{"牛肉", "面条"},
	}, testRules())
	count := 0
	for _, v := range got {
		if v == "beef" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("同一标签应只出现一次，实际 %d 次: %v", count, got)
	}
}

func TestGenerate_文本规则匹配菜名与描述(t *testing.T) {
	got := Generate(Input{Name: "番茄鸡蛋汤"}, testRules())
	if !contains(got, "soup") {
		t.Errorf("菜名含「汤」应命中汤标签: %v", got)
	}

	got = Generate(Input{Name: "凉拌黄瓜"}, testRules())
	if !contains(got, "cold") {
		t.Errorf("菜名含「凉拌」应命中凉菜标签: %v", got)
	}
}

func TestGenerate_无规则返回空(t *testing.T) {
	got := Generate(Input{Name: "神秘料理", Ingredients: []string{"未知"}}, nil)
	if len(got) != 0 {
		t.Errorf("无规则时应返回空: %v", got)
	}
}

func TestGenerate_规则顺序无关(t *testing.T) {
	rules := testRules()
	// 打乱顺序后 group_mutex 仍应在普通规则之后求值
	shuffled := []*model.TagRule{rules[10], rules[0], rules[11], rules[4]}
	got := Generate(Input{Name: "土豆炖牛肉", Ingredients: []string{"牛肉", "土豆"}}, shuffled)
	if !contains(got, "diet-meat") {
		t.Errorf("乱序规则下荤菜判定应仍然生效: %v", got)
	}
	if contains(got, "diet-veg") {
		t.Errorf("乱序规则下不应判为素菜: %v", got)
	}
}

func TestHasCoreTags(t *testing.T) {
	rules := testRules()
	if !HasCoreTags([]string{"quick", "beef"}, rules) {
		t.Error("包含互斥组成员 beef 时应返回 true")
	}
	if HasCoreTags([]string{"quick", "soup"}, rules) {
		t.Error("不含互斥组成员时应返回 false")
	}
	if HasCoreTags(nil, rules) {
		t.Error("空结果应返回 false")
	}
}
