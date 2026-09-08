package persistent

import (
	"context"
	"errors"

	"github.com/railzwaylabs/github-actions-samples/internal/order/domain"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	ListOrder(context.Context, *domain.Order) ([]*domain.Order, error)
	GetOrder(context.Context, string) (*domain.Order, error)
	CreateOrder(context.Context, *domain.Order) error
	BatchCreateOrder(context.Context, []*domain.Order) error

	CreateOrderItem(context.Context, *domain.OrderItems) error
	UpdateOrderItem(context.Context, string, *domain.OrderItems) error
	DeleteOrderItem(context.Context, string) error
}

type repository struct {
	DB *gorm.DB
}

type Params struct {
	fx.In

	DB *gorm.DB
}

func NewRepository(p Params) Repository {
	return &repository{
		DB: p.DB,
	}
}

func (r *repository) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	var order domain.Order
	err := r.DB.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) ListOrder(ctx context.Context, filter *domain.Order) ([]*domain.Order, error) {
	orders := make([]*domain.Order, 0)
	query := r.DB.WithContext(ctx).Preload("Items")
	if filter != nil {
		query = query.Where(filter)
	}
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *repository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return r.DB.WithContext(ctx).Create(order).Error
}

func (r *repository) BatchCreateOrder(ctx context.Context, orders []*domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).Create(&orders).Error
}

func (r *repository) CreateOrderItem(ctx context.Context, item *domain.OrderItems) error {
	return r.DB.WithContext(ctx).Create(item).Error
}

func (r *repository) UpdateOrderItem(ctx context.Context, id string, item *domain.OrderItems) error {
	return r.DB.WithContext(ctx).Model(&domain.OrderItems{}).Where("id = ?", id).Updates(item).Error
}

func (r *repository) DeleteOrderItem(ctx context.Context, id string) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&domain.OrderItems{}).Error
}
