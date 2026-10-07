package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/model"
	"foodlog/server/internal/repository"
)

var ErrNotFound = errors.New("资源不存在")

// RestaurantService 餐厅服务（含店内菜品）
type RestaurantService struct {
	pool           *pgxpool.Pool
	restaurantRepo *repository.RestaurantRepo
	dishRepo       *repository.DishRepo
	shopTagRepo    *repository.ShopTagRepo
	shopTagService *ShopTagService
}

func NewRestaurantService(
	pool *pgxpool.Pool,
	restaurantRepo *repository.RestaurantRepo,
	dishRepo *repository.DishRepo,
	shopTagRepo *repository.ShopTagRepo,
	shopTagService *ShopTagService,
) *RestaurantService {
	return &RestaurantService{
		pool:           pool,
		restaurantRepo: restaurantRepo,
		dishRepo:       dishRepo,
		shopTagRepo:    shopTagRepo,
		shopTagService: shopTagService,
	}
}

// prepareTags 校验并归一化餐厅标签 id
func (s *RestaurantService) prepareTags(ctx context.Context, userID string, rst *model.Restaurant) error {
	tags, err := s.shopTagService.ResolveManualTags(ctx, userID, repository.ShopTagDomainRestaurant, rst.Tags)
	if err != nil {
		return err
	}
	rst.Tags = tags
	return nil
}

// prepareDishTags 校验并归一化菜品标签 id
func (s *RestaurantService) prepareDishTags(ctx context.Context, userID string, d *model.Dish) error {
	tags, err := s.shopTagService.ResolveManualTags(ctx, userID, repository.ShopTagDomainDish, d.Tags)
	if err != nil {
		return err
	}
	d.Tags = tags
	return nil
}

// fillTags 批量填充餐厅标签 id
func (s *RestaurantService) fillTags(ctx context.Context, list []*model.Restaurant) {
	if len(list) == 0 {
		return
	}
	ids := make([]string, len(list))
	for i, r := range list {
		ids[i] = r.ID
	}
	byID, err := s.shopTagRepo.TagsFor(ctx, repository.ShopTagDomainRestaurant, ids)
	if err != nil {
		return
	}
	for _, r := range list {
		if v, ok := byID[r.ID]; ok {
			r.Tags = v
		} else {
			r.Tags = []string{}
		}
	}
}

// fillDishTags 批量填充菜品标签 id
func (s *RestaurantService) fillDishTags(ctx context.Context, list []*model.Dish) {
	if len(list) == 0 {
		return
	}
	ids := make([]string, len(list))
	for i, d := range list {
		ids[i] = d.ID
	}
	byID, err := s.shopTagRepo.TagsFor(ctx, repository.ShopTagDomainDish, ids)
	if err != nil {
		return
	}
	for _, d := range list {
		if v, ok := byID[d.ID]; ok {
			d.Tags = v
		} else {
			d.Tags = []string{}
		}
	}
}

// Create 创建餐厅（餐厅行 + 标签关联在同一事务内写入）
func (s *RestaurantService) Create(ctx context.Context, userID string, rst *model.Restaurant) (*model.Restaurant, error) {
	rst.UserID = userID
	if err := s.prepareTags(ctx, userID, rst); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.restaurantRepo.CreateTx(ctx, tx, rst); err != nil {
		return nil, err
	}
	if err := s.shopTagRepo.ReplaceTagsTx(ctx, tx, repository.ShopTagDomainRestaurant, rst.ID, rst.Tags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return rst, nil
}

// Get 获取餐厅（校验归属，并填充标签）
func (s *RestaurantService) Get(ctx context.Context, userID, id string) (*model.Restaurant, error) {
	rst, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rst == nil {
		return nil, ErrNotFound
	}
	if rst.UserID != userID {
		return nil, ErrForbidden
	}
	s.fillTags(ctx, []*model.Restaurant{rst})
	return rst, nil
}

// GetDetail 获取餐厅详情（含菜品列表，可按菜品标签过滤）
func (s *RestaurantService) GetDetail(ctx context.Context, userID, id, dishTagID string) (*model.RestaurantDetail, error) {
	rst, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	dishes, err := s.dishRepo.ListByRestaurant(ctx, id, dishTagID)
	if err != nil {
		return nil, err
	}
	s.fillDishTags(ctx, dishes)

	items := make([]model.Dish, len(dishes))
	for i, d := range dishes {
		items[i] = *d
	}
	return &model.RestaurantDetail{Restaurant: *rst, Dishes: items}, nil
}

// List 餐厅列表（分页 + 关键词 + 标签筛选 + 排序）
func (s *RestaurantService) List(ctx context.Context, userID string, q *model.PageQuery, keyword, tagID, sort string) (*model.Paginated, error) {
	list, total, err := s.restaurantRepo.List(ctx, userID, q, keyword, tagID, sort)
	if err != nil {
		return nil, err
	}
	s.fillTags(ctx, list)

	items := make([]any, len(list))
	for i, r := range list {
		items[i] = r
	}
	return &model.Paginated{List: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Update 更新餐厅（校验归属，餐厅行 + 标签关联在同一事务内写入）
func (s *RestaurantService) Update(ctx context.Context, userID, id string, rst *model.Restaurant) (*model.Restaurant, error) {
	existing, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	rst.ID = existing.ID
	rst.UserID = existing.UserID
	if err := s.prepareTags(ctx, userID, rst); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.restaurantRepo.UpdateTx(ctx, tx, rst); err != nil {
		return nil, err
	}
	if err := s.shopTagRepo.ReplaceTagsTx(ctx, tx, repository.ShopTagDomainRestaurant, rst.ID, rst.Tags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return rst, nil
}

// Delete 删除餐厅（校验归属，级联删除其下菜品与标签关联）
func (s *RestaurantService) Delete(ctx context.Context, userID, id string) error {
	if _, err := s.Get(ctx, userID, id); err != nil {
		return err
	}
	return s.restaurantRepo.Delete(ctx, id)
}

// AddDish 添加菜品（校验餐厅归属，菜品行 + 标签关联在同一事务内写入）
func (s *RestaurantService) AddDish(ctx context.Context, userID, restaurantID string, d *model.Dish) (*model.Dish, error) {
	if _, err := s.Get(ctx, userID, restaurantID); err != nil {
		return nil, err
	}
	d.RestaurantID = restaurantID
	d.UserID = userID
	if err := s.prepareDishTags(ctx, userID, d); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.dishRepo.CreateTx(ctx, tx, d); err != nil {
		return nil, err
	}
	if err := s.shopTagRepo.ReplaceTagsTx(ctx, tx, repository.ShopTagDomainDish, d.ID, d.Tags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

// UpdateDish 更新菜品（校验归属，菜品行 + 标签关联在同一事务内写入）
func (s *RestaurantService) UpdateDish(ctx context.Context, userID, dishID string, d *model.Dish) (*model.Dish, error) {
	existing, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.UserID != userID {
		return nil, ErrForbidden
	}
	d.ID = existing.ID
	d.RestaurantID = existing.RestaurantID
	d.UserID = existing.UserID
	if err := s.prepareDishTags(ctx, userID, d); err != nil {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // 提交成功后回滚为无操作

	if err := s.dishRepo.UpdateTx(ctx, tx, d); err != nil {
		return nil, err
	}
	if err := s.shopTagRepo.ReplaceTagsTx(ctx, tx, repository.ShopTagDomainDish, d.ID, d.Tags); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

// DeleteDish 删除菜品（校验归属，级联删除标签关联）
func (s *RestaurantService) DeleteDish(ctx context.Context, userID, dishID string) error {
	existing, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if existing.UserID != userID {
		return ErrForbidden
	}
	return s.dishRepo.Delete(ctx, dishID)
}
