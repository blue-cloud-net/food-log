package model

// TagRuleType 自动标签规则类型
type TagRuleType string

const (
	// RuleIngredientKeyword 逐个食材名匹配关键词（含排除词）
	RuleIngredientKeyword TagRuleType = "ingredient_keyword"
	// RuleIngredientTextKeyword 食材拼接文本匹配关键词
	RuleIngredientTextKeyword TagRuleType = "ingredient_text_keyword"
	// RuleTextKeyword 按 match_field 指定文本匹配关键词
	RuleTextKeyword TagRuleType = "text_keyword"
	// RuleCookTimeMax 耗时小于等于 max_minutes
	RuleCookTimeMax TagRuleType = "cook_time_max"
	// RuleGroupMutex 已命中标签落在成员集合内；同 tag_group 只取首个命中
	RuleGroupMutex TagRuleType = "group_mutex"
)

// TagTextMatchField text_keyword 规则的匹配字段
type TagTextMatchField string

const (
	TagFieldName            TagTextMatchField = "name"
	TagFieldDescription     TagTextMatchField = "description"
	TagFieldNameDescription TagTextMatchField = "name_description"
)

// Tag 标签。OwnerID 为空表示全局预设标签，否则为该用户私有自定义标签。
type Tag struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	OwnerID    string `json:"owner_id,omitempty"`
	Name       string `json:"name"`
	MutexGroup string `json:"mutex_group,omitempty"`
	SortOrder  int    `json:"sort_order"`
	IsSystem   bool   `json:"is_system"`
}

// TagCategory 标签分类，含其下标签
type TagCategory struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id,omitempty"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsSystem  bool   `json:"is_system"`
	Tags      []Tag  `json:"tags"`
}

// TagRule 自动标签规则
type TagRule struct {
	ID              string      `json:"id"`
	TagID           string      `json:"tag_id"`
	Type            TagRuleType `json:"rule_type"`
	MatchField      string      `json:"match_field,omitempty"`
	Keywords        []string    `json:"keywords"`
	ExcludeKeywords []string    `json:"exclude_keywords"`
	MaxMinutes      int         `json:"max_minutes,omitempty"`
	TagGroup        string      `json:"tag_group,omitempty"`
	MemberTagIDs    []string    `json:"member_tag_ids"`
	SortOrder       int         `json:"sort_order"`
}
