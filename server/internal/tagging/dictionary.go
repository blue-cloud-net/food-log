package tagging

import "strings"

// TagCategory 预设标签分类（供前端选择器与 GET /api/recipes/tags 使用）
type TagCategory struct {
	Name string   `json:"name"` // 分类名
	Tags []string `json:"tags"` // 该分类下所有标签
}

// TagCategories 预设词表
var TagCategories = []TagCategory{
	{Name: "荤素", Tags: []string{"荤菜", "素菜"}},
	{Name: "食材", Tags: []string{"牛肉", "猪肉", "鸡肉", "羊肉", "鸭肉", "鱼", "海鲜", "虾", "蟹", "蛋", "豆制品", "蔬菜", "菌菇", "主食"}},
	{Name: "场景", Tags: []string{"快手", "汤", "凉菜", "面食", "甜点"}},
	{Name: "时段", Tags: []string{"早餐", "午餐", "晚餐", "夜宵"}},
	{Name: "菜系", Tags: []string{"川菜", "粤菜", "湘菜", "鲁菜", "苏菜", "浙菜", "闽菜", "徽菜", "东北菜", "西北菜"}},
	{Name: "口味", Tags: []string{"辣", "清淡", "甜", "酸", "咸鲜"}},
}

// ingredientRule 食材关键词 → 标签规则
type ingredientRule struct {
	Tag      string
	Keywords []string
	Exclude  []string // 食材名包含这些词时跳过该规则（避免误判，如"鱼香茄子"）
}

// ingredientRules 食材标签规则（关键词尽量具体，避免泛词歧义）
var ingredientRules = []ingredientRule{
	{Tag: "牛肉", Keywords: []string{"牛肉", "牛腩", "牛腱", "牛排", "肥牛", "牛里脊"}},
	{Tag: "猪肉", Keywords: []string{"五花肉", "猪里脊", "猪蹄", "猪肝", "猪肚", "培根", "火腿", "香肠", "腊肉", "猪肉"}},
	{Tag: "羊肉", Keywords: []string{"羊肉", "羊排", "羊蝎子", "肥羊"}},
	{Tag: "鸡肉", Keywords: []string{"鸡腿", "鸡翅", "鸡胸", "鸡爪", "鸡肉", "三黄鸡", "土鸡", "乌鸡"}},
	{Tag: "鸭肉", Keywords: []string{"鸭腿", "鸭肉", "烤鸭", "盐水鸭"}},
	{Tag: "鱼", Keywords: []string{"三文鱼", "鲈鱼", "草鱼", "鲫鱼", "带鱼", "鳕鱼", "鲤鱼", "罗非鱼", "多宝鱼", "黄花鱼", "鱼丸", "鱼", "鱼片"},
		Exclude: []string{"鱼香"}},
	{Tag: "虾", Keywords: []string{"基围虾", "大虾", "虾仁", "虾皮", "小龙虾", "河虾", "虾滑", "虾"}},
	{Tag: "蟹", Keywords: []string{"大闸蟹", "梭子蟹", "螃蟹", "蟹"}},
	{Tag: "海鲜", Keywords: []string{"鱿鱼", "章鱼", "扇贝", "蛤蜊", "花甲", "生蚝", "牡蛎", "鲍鱼", "海参", "海螺", "蛏子", "青口", "海胆"}},
	{Tag: "蛋", Keywords: []string{"鹌鹑蛋", "皮蛋", "咸蛋", "鸭蛋", "鸡蛋", "蛋"}},
	{Tag: "豆制品", Keywords: []string{"油豆腐", "豆腐干", "豆腐", "豆干", "豆皮", "腐竹", "千张", "豆花", "豆浆", "素鸡"}},
	{Tag: "菌菇", Keywords: []string{"金针菇", "杏鲍菇", "茶树菇", "海鲜菇", "香菇", "蘑菇", "平菇", "木耳", "银耳"}},
	{Tag: "蔬菜", Keywords: []string{"西兰花", "花菜", "油麦菜", "空心菜", "生菜", "菠菜", "青菜", "小白菜", "白菜", "包菜", "卷心菜", "番茄", "西红柿", "土豆", "马铃薯", "胡萝卜", "白萝卜", "黄瓜", "茄子", "青椒", "彩椒", "尖椒", "洋葱", "冬瓜", "丝瓜", "苦瓜", "南瓜", "莲藕", "山药", "芋头", "芹菜", "韭菜", "豆角", "荷兰豆", "玉米", "竹笋", "莴笋", "苋菜", "蒜薹", "秋葵", "紫甘蓝", "芦笋", "西葫芦"}},
	{Tag: "主食", Keywords: []string{"米饭", "炒饭", "粥", "馒头", "包子", "饺子", "馄饨", "云吞", "面条", "挂面", "米粉", "河粉", "年糕", "意面", "饼"}},
}

// meatTags 判定"荤菜"的标签集合
var meatTags = map[string]bool{
	"牛肉": true, "猪肉": true, "鸡肉": true, "羊肉": true, "鸭肉": true,
	"鱼": true, "虾": true, "蟹": true, "海鲜": true, "蛋": true,
}

// vegTags 判定"素菜"的标签集合
var vegTags = map[string]bool{
	"豆制品": true, "菌菇": true, "蔬菜": true,
}

// 口味/场景规则（由菜名或食材推导）
var spicyKeywords = []string{"辣椒", "小米辣", "尖椒", "螺丝椒", "花椒", "藤椒", "豆瓣", "剁椒", "泡椒", "辣椒粉", "辣椒面", "油泼辣子"}

var soupKeywords = []string{"汤"}

var coldKeywords = []string{"凉拌", "凉菜"}

var dessertKeywords = []string{"蛋糕", "甜点", "布丁", "冰淇淋", "雪糕", "糖水", "曲奇", "饼干"}

var noodleKeywords = []string{"面条", "挂面", "米粉", "河粉", "意面", "凉面", "炒面", "拉面", "馒头", "包子", "饺子", "馄饨", "云吞", "饼"}

// matched 判断食材名是否命中规则（考虑排除词）
func matched(name string, rule ingredientRule) bool {
	for _, ex := range rule.Exclude {
		if strings.Contains(name, ex) {
			return false
		}
	}
	for _, kw := range rule.Keywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// containsAny 判断文本是否包含任一关键词
func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
