package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// ErrInventoryInvalid 库存条目参数无效
var ErrInventoryInvalid = errors.New("库存食材参数无效")

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

// prepare 归一化并校验条目字段，同时确认存放位置对该用户可见
func (s *InventoryService) prepare(ctx context.Context, userID string, it *model.InventoryItem) error {
	it.Name = strings.TrimSpace(it.Name)
	if it.Name == "" || len([]rune(it.Name)) > 100 {
		return fmt.Errorf("%w: 食材名称长度需为 1-100", ErrInventoryInvalid)
	}
	it.Amount = strings.TrimSpace(it.Amount)
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
