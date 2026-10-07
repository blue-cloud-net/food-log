// Package migrations 内嵌并暴露 SQL 迁移文件
//
// 迁移文件按文件名升序执行，命名约定：
//
//	001_schema.sql            结构初始化（全部表 / 索引 / 触发器）
//	002_catalog.sql           初始数据（菜谱全局预设标签字典），开发与生产都执行
//	004_shop_tags.sql         探店标签表结构 + 餐厅评分维度变更
//	005_shop_tag_catalog.sql  初始数据（探店全局预设标签字典），开发与生产都执行
//	009_demo_seed.sql         演示数据，仅 APP_ENV=development 执行
//	010_recipe_state_flags.sql 状态字段（已做日期 / 喜欢 / 已探店 / 喜欢菜品）
//
// 注意：演示数据引用了结构类迁移新增的列与表，因此文件名必须排在它们之后
// （故为 009 而非早期的 003）。
package migrations

import (
	"embed"
	"sort"
)

//go:embed *.sql
var FS embed.FS

// DemoSeedFile 演示数据文件名（仅开发模式执行）
const DemoSeedFile = "009_demo_seed.sql"

// File 迁移文件
type File struct {
	Name string
	SQL  string
}

// List 返回按文件名升序排列的迁移文件
func List() ([]File, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	files := make([]File, 0, len(names))
	for _, name := range names {
		content, err := FS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		files = append(files, File{Name: name, SQL: string(content)})
	}
	return files, nil
}
