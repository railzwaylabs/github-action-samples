package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/railzwaylabs/github-actions-samples/internal/order/domain"
	"github.com/railzwaylabs/github-actions-samples/internal/order/infrastructure/persistent"
)

var (
	ErrInvalidOrderID = errors.New("invalid order ID")
	ErrOrderNotFound  = errors.New("order not found")
)

type CreateOrderInput struct {
	CustomerID string
	Items      []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int64
	Price     int64
}

type Service interface {
	List(context.Context) ([]*domain.Order, error)
	Get(context.Context, string) (*domain.Order, error)
	Create(context.Context, CreateOrderInput) (*domain.Order, error)
}

type service struct{ orders persistent.Repository }

func NewService(orders persistent.Repository) Service { return &service{orders: orders} }

func (s *service) List(ctx context.Context) ([]*domain.Order, error) {
	return s.orders.ListOrder(ctx, nil)
}

func (s *service) Get(ctx context.Context, id string) (*domain.Order, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidOrderID
	}
	order, err := s.orders.GetOrder(ctx, id)
	if errors.Is(err, persistent.ErrNotFound) {
		return nil, ErrOrderNotFound
	}
	return order, err
}

func (s *service) Create(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if strings.TrimSpace(input.CustomerID) == "" {
		return nil, errors.New("customer_id is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}
	orderID, err := randomUUID()
	if err != nil {
		return nil, fmt.Errorf("generate order ID: %w", err)
	}
	now := time.Now().UTC()
	order := &domain.Order{ID: orderID, OrderID: orderID, CustomerID: strings.TrimSpace(input.CustomerID), Status: 1, CreatedAt: now, UpdatedAt: now}
	order.Items = make([]*domain.OrderItems, 0, len(input.Items))
	for index, inputItem := range input.Items {
		if strings.TrimSpace(inputItem.ProductID) == "" || inputItem.Quantity <= 0 || inputItem.Price < 0 {
			return nil, fmt.Errorf("item %d must have product_id, positive quantity, and non-negative price", index)
		}
		if inputItem.Price > 0 && inputItem.Quantity > math.MaxInt64/inputItem.Price {
			return nil, fmt.Errorf("item %d subtotal exceeds int64", index)
		}
		itemID, idErr := randomUUID()
		if idErr != nil {
			return nil, fmt.Errorf("generate item ID: %w", idErr)
		}
		order.Items = append(order.Items, &domain.OrderItems{
			ID: itemID, OrderID: orderID, ProductID: strings.TrimSpace(inputItem.ProductID),
			Quantity: inputItem.Quantity, Price: inputItem.Price, Subtotal: inputItem.Quantity * inputItem.Price,
			CreatedAt: now, UpdatedAt: now,
		})
	}
	if err := s.orders.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return order, nil
}

func randomUUID() (string, error) {
	value, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return value.String(), nil
}
