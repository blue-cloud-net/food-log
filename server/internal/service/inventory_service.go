package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// ErrInventoryInvalid 库存条目参数无效
var ErrInventoryInvalid = errors.New("库存食材参数无效")

// maxBatchDeleteInventory 单次批量删除的条目上限
const maxBatchDeleteInventory = 200

// maxInventoryQuantity 单个条目允许的最大数量（与 NUMERIC(10,2) 量级匹配）
const maxInventoryQuantity = 99999999

// ConsumeResult 消耗结果
type ConsumeResult struct {
	Remaining float64              // 剩余数量（已移出时为 0）
	Removed   bool                 // 是否已减到 0 并移出库存
	Item      *model.InventoryItem // 剩余条目（已移出时为 nil）
}

// InventoryService 库存食材服务
type InventoryService struct {
	repo            *repository.InventoryRepo
	locationService *StorageLocationService
}

func NewInventoryService(repo *repository.InventoryRepo, locationService *StorageLocationService) *InventoryService {
	return &InventoryService{repo: repo, locationService: locationService}
}

// List 库存列表（筛选 + 排序）
func (s *InventoryService) List(ctx context.Context, userID string, f repository.InventoryFilter) ([]*model.InventoryItem, error) {
	return s.repo.List(ctx, userID, f)
}

// Get 获取库存条目（校验归属）
func (s *InventoryService) Get(ctx context.Context, userID, id string) (*model.InventoryItem, error) {
	it, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, ErrNotFound
	}
	if it.UserID != userID {
		return nil, ErrForbidden
	}
	return it, nil
}

// Create 创建库存条目
func (s *InventoryService) Create(ctx context.Context, userID string, it *model.InventoryItem) (*model.InventoryItem, error) {
	if err := s.prepare(ctx, userID, it); err != nil {
		return nil, err
	}
	it.UserID = userID
	if err := s.repo.Create(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

// Update 更新库存条目（校验归属）
func (s *InventoryService) Update(ctx context.Context, userID, id string, it *model.InventoryItem) (*model.InventoryItem, error) {
	existing, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := s.prepare(ctx, userID, it); err != nil {
		return nil, err
	}
	it.ID = existing.ID
	it.UserID = existing.UserID
	if err := s.repo.Update(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

// Delete 删除库存条目（校验归属）
func (s *InventoryService) Delete(ctx context.Context, userID, id string) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// BatchDelete 批量删除库存条目（仅限本人；id 去重后一次性删除，返回删除数量）
func (s *InventoryService) BatchDelete(ctx context.Context, userID string, ids []string) (int64, error) {
	uniq := make([]string, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		if !uuidRe.MatchString(id) {
			return 0, fmt.Errorf("%w: 条目 id 无效: %s", ErrInventoryInvalid, id)
		}
		seen[id] = true
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return 0, fmt.Errorf("%w: 请选择要删除的条目", ErrInventoryInvalid)
	}
	if len(uniq) > maxBatchDeleteInventory {
		return 0, fmt.Errorf("%w: 一次最多删除 %d 条", ErrInventoryInvalid, maxBatchDeleteInventory)
	}
	return s.repo.DeleteMany(ctx, userID, uniq)
}

// Consume 消耗一定数量（吃掉 / 用掉）；数量减到 0 时自动移出库存。
// 归属校验与「不存在」区分同 Get：不存在 → ErrNotFound，非本人 → ErrForbidden。
func (s *InventoryService) Consume(ctx context.Context, userID, id string, amount float64) (*ConsumeResult, error) {
	it, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: 消耗数量需大于 0", ErrInventoryInvalid)
	}
	amount = math.Round(amount*100) / 100
	if amount > it.Quantity {
		return nil, fmt.Errorf("%w: 消耗数量不能超过剩余数量（%v）", ErrInventoryInvalid, it.Quantity)
	}

	remaining, removed, ok, err := s.repo.Consume(ctx, userID, id, amount)
	if err != nil {
		return nil, err
	}
	if !ok {
		// 并发下数量已被其它请求改小
		return nil, fmt.Errorf("%w: 数量已变化，请刷新后重试", ErrInventoryInvalid)
	}

	res := &ConsumeResult{Remaining: remaining, Removed: removed}
	if !removed {
		if res.Item, err = s.repo.GetByID(ctx, id); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// prepare 归一化并校验条目字段，同时确认存放位置对该用户可见
func (s *InventoryService) prepare(ctx context.Context, userID string, it *model.InventoryItem) error {
	it.Name = strings.TrimSpace(it.Name)
	if it.Name == "" || len([]rune(it.Name)) > 100 {
		return fmt.Errorf("%w: 食材名称长度需为 1-100", ErrInventoryInvalid)
	}
	it.Quantity = math.Round(it.Quantity*100) / 100
	if it.Quantity <= 0 {
		return fmt.Errorf("%w: 数量需大于 0", ErrInventoryInvalid)
	}
	if it.Quantity > maxInventoryQuantity {
		return fmt.Errorf("%w: 数量过大", ErrInventoryInvalid)
	}
	it.Unit = strings.TrimSpace(it.Unit)
	it.Category = strings.TrimSpace(it.Category)
	it.Note = strings.TrimSpace(it.Note)
	if len([]rune(it.Note)) > 500 {
		return fmt.Errorf("%w: 备注最长 500 字", ErrInventoryInvalid)
	}
	if len(it.Images) == 0 {
		it.Images = []string{}
	}

	if it.ExpireAt != nil {
		v := strings.TrimSpace(*it.ExpireAt)
		switch {
		case v == "":
			it.ExpireAt = nil
		default:
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return fmt.Errorf("%w: 过期日期需为 YYYY-MM-DD", ErrInventoryInvalid)
			}
			it.ExpireAt = &v
		}
	}

	if _, err := s.locationService.ResolveLocationID(ctx, userID, it.LocationID); err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrLocationInvalid) {
			return fmt.Errorf("%w: 存放位置不存在", ErrInventoryInvalid)
		}
		return err
	}
	return nil
}
