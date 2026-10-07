package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

// 存放位置相关业务错误
var (
	ErrLocationInvalid   = errors.New("存放位置参数无效")
	ErrLocationDuplicate = errors.New("同名存放位置已存在")
	ErrLocationReadonly  = errors.New("系统预设存放位置不可修改")
	ErrLocationInUse     = errors.New("该存放位置下仍有库存食材，无法删除")
)

// validStorageAreas 存放位置大类白名单
var validStorageAreas = map[string]bool{
	string(model.StorageAreaFridge):  true,
	string(model.StorageAreaOutside): true,
}

// StorageLocationService 存放位置字典服务：词表读取 + 用户自定义位置维护
type StorageLocationService struct {
	repo *repository.StorageLocationRepo
}

func NewStorageLocationService(repo *repository.StorageLocationRepo) *StorageLocationService {
	return &StorageLocationService{repo: repo}
}

// List 返回该用户可见的存放位置（全局预设 + 本人自定义）
func (s *StorageLocationService) List(ctx context.Context, userID string) ([]*model.StorageLocation, error) {
	return s.repo.ListVisible(ctx, userID)
}

// Create 新建用户自定义存放位置
func (s *StorageLocationService) Create(ctx context.Context, userID, area, name string) (*model.StorageLocation, error) {
	area, name, err := normalizeLocationInput(area, name)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUniqueName(ctx, userID, area, name, ""); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, userID, area, name, 900)
}

// Update 更新本人自定义存放位置
func (s *StorageLocationService) Update(ctx context.Context, userID, id, area, name string, sortOrder int) (*model.StorageLocation, error) {
	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, ErrNotFound
	}
	if loc.IsSystem {
		return nil, ErrLocationReadonly
	}
	if loc.OwnerID != userID {
		return nil, ErrForbidden
	}

	area, name, err = normalizeLocationInput(area, name)
	if err != nil {
		return nil, err
	}
	if err := s.ensureUniqueName(ctx, userID, area, name, id); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, id, area, name, sortOrder); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete 删除本人自定义存放位置；位置下仍有库存条目时拒绝删除
func (s *StorageLocationService) Delete(ctx context.Context, userID, id string) error {
	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if loc == nil {
		return ErrNotFound
	}
	if loc.IsSystem {
		return ErrLocationReadonly
	}
	if loc.OwnerID != userID {
		return ErrForbidden
	}

	n, err := s.repo.CountItems(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w（还有 %d 项）", ErrLocationInUse, n)
	}
	return s.repo.Delete(ctx, id)
}

// ResolveLocationID 校验位置存在且对该用户可见。
// 他人私有位置按「不存在」处理，避免泄露。
func (s *StorageLocationService) ResolveLocationID(ctx context.Context, userID, id string) (*model.StorageLocation, error) {
	if strings.TrimSpace(id) == "" || !uuidRe.MatchString(id) {
		return nil, ErrLocationInvalid
	}
	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, ErrNotFound
	}
	if loc.OwnerID != "" && loc.OwnerID != userID {
		return nil, ErrNotFound
	}
	return loc, nil
}

func (s *StorageLocationService) ensureUniqueName(ctx context.Context, userID, area, name, excludeID string) error {
	taken, err := s.repo.NameTaken(ctx, userID, area, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return ErrLocationDuplicate
	}
	return nil
}

func normalizeLocationInput(area, name string) (string, string, error) {
	area = strings.TrimSpace(area)
	name = strings.TrimSpace(name)
	if !validStorageAreas[area] {
		return "", "", fmt.Errorf("%w: area 必须为 fridge 或 outside", ErrLocationInvalid)
	}
	if name == "" || len([]rune(name)) > 50 {
		return "", "", fmt.Errorf("%w: 名称长度需为 1-50", ErrLocationInvalid)
	}
	return area, name, nil
}
