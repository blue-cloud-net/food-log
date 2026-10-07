package model

import "time"

// StorageArea 存放位置所属大类
type StorageArea string

const (
	// StorageAreaFridge 冰箱内
	StorageAreaFridge StorageArea = "fridge"
	// StorageAreaOutside 外面（常温存放）
	StorageAreaOutside StorageArea = "outside"
)

// StorageLocation 存放位置。OwnerID 为空表示全局预设位置，否则为该用户私有自定义位置。
type StorageLocation struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id,omitempty"`
	Area      string `json:"area"` // fridge / outside
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsSystem  bool   `json:"is_system"`
}

// InventoryItem 库存食材条目
type InventoryItem struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	LocationID string    `json:"location_id"`
	Name       string    `json:"name"`
	Amount     string    `json:"amount"`
	Unit       string    `json:"unit"`
	Category   string    `json:"category"`  // 食材分类（自由文本，前端给预设建议）
	ExpireAt   *string   `json:"expire_at"` // 过期日期（YYYY-MM-DD，nil = 未设置）
	Note       string    `json:"note"`
	Images     []string  `json:"images"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
