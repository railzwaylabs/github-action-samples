package domain

import "time"

type Order struct {
	ID        string     `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	CustomerID string `gorm:"column:customer_id" json:"customer_id"`
	OrderID    string `gorm:"column:order_id;type:uuid" json:"order_id"`

	Status int8 `gorm:"column:status" json:"status"`

	Items []*OrderItems `gorm:"foreignKey:OrderID;references:ID" json:"items"`
}

func (m *Order) TableName() string {
	return "orders"
}

type OrderItems struct {
	ID        string     `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	OrderID   string `gorm:"column:order_id;type:uuid" json:"order_id"`
	ProductID string `gorm:"column:product_id" json:"product_id"`
	Quantity  int64  `gorm:"column:quantity" json:"quantity"`
	Price     int64  `gorm:"column:price" json:"price"`

	Subtotal int64 `gorm:"column:subtotal" json:"subtotal"`
}

func (m *OrderItems) TableName() string {
	return "order_items"
}
