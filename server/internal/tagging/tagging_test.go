package tagging

import (
	"context"
	"errors"
	"testing"
)

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestGenerate_牛肉荤菜(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:            "土豆炖牛肉",
		Ingredients:     []string{"牛肉", "土豆"},
		CookTimeMinutes: 30,
	}, nil, nil)
	for _, want := range []string{"荤菜", "牛肉", "蔬菜"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "素菜") {
		t.Errorf("牛肉菜不应标为素菜: %v", got)
	}
}

func TestGenerate_素菜(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:        "麻婆豆腐",
		Ingredients: []string{"嫩豆腐", "豆瓣酱"},
	}, nil, nil)
	for _, want := range []string{"素菜", "豆制品", "辣"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "荤菜") {
		t.Errorf("纯素菜不应标为荤菜: %v", got)
	}
}

func TestGenerate_蛋算荤菜(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:        "番茄炒蛋",
		Ingredients: []string{"番茄", "鸡蛋"},
	}, nil, nil)
	for _, want := range []string{"荤菜", "蛋", "蔬菜"} {
		if !contains(got, want) {
			t.Errorf("缺少标签 %q，实际: %v", want, got)
		}
	}
	if contains(got, "素菜") {
		t.Errorf("含蛋菜不应标为素菜: %v", got)
	}
}

func TestGenerate_快手(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:            "清炒时蔬",
		Ingredients:     []string{"青菜"},
		CookTimeMinutes: 10,
	}, nil, nil)
	if !contains(got, "快手") {
		t.Errorf("10 分钟应标为快手: %v", got)
	}
}

func TestGenerate_鱼香排除(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:        "鱼香茄子",
		Ingredients: []string{"茄子"},
	}, nil, nil)
	if contains(got, "鱼") {
		t.Errorf("鱼香茄子不应标为鱼: %v", got)
	}
}

func TestGenerate_手动合并去重(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:        "土豆炖牛肉",
		Ingredients: []string{"牛肉"},
	}, []string{"牛肉", "我的最爱"}, nil)
	// 牛肉去重后只出现一次
	count := 0
	for _, v := range got {
		if v == "牛肉" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("牛肉应去重，实际出现 %d 次: %v", count, got)
	}
	if !contains(got, "我的最爱") {
		t.Errorf("手动标签应保留: %v", got)
	}
}

type mockCompleter struct {
	tags []string
	err  error
}

func (m mockCompleter) CompleteTags(_ context.Context, _ string) ([]string, error) {
	return m.tags, m.err
}

func TestGenerate_AI兜底(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name: "狮子头",
	}, nil, mockCompleter{tags: []string{"淮扬菜", "下饭菜"}})
	if !contains(got, "淮扬菜") {
		t.Errorf("词典未命中时应使用 AI 兜底: %v", got)
	}
}

func TestGenerate_AI失败静默降级(t *testing.T) {
	got := Generate(context.Background(), Input{
		Name:        "土豆炖牛肉",
		Ingredients: []string{"牛肉"},
	}, nil, mockCompleter{err: errors.New("AI 挂了")})
	if !contains(got, "牛肉") {
		t.Errorf("AI 失败时应保留词典结果: %v", got)
	}
}
