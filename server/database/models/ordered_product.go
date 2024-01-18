package models

import (
	"estore-backend/server/models"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type OrderedProduct struct {
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	// Read Only: true
	OrderID int64  `json:"orderId,omitempty"`
	Order   *Order `bun:"rel:belongs-to,join:order_id=id"`

	// product Id
	// Required: true
	ProductID *int64 `json:"productId"`

	// Read Only: true
	InStock *bool `json:"inStock,omitempty" bun:",scanonly"`

	// Read Only: true
	ProductName string `json:"productName,omitempty" bun:",scanonly"`

	// quantity
	// Required: true
	Quantity *int64 `json:"quantity"`

	// total price
	// Required: true
	TotalPrice *float64 `json:"totalPrice"`
}

var _ bun.BeforeCreateTableHook = (*OrderedProduct)(nil)

func (m *OrderedProduct) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE`)
	return nil
}

func NewOrderedProductFrom(dto *models.OrderedProduct) *OrderedProduct {
	return &OrderedProduct{
		ID:         0,
		OrderID:    dto.OrderID,
		ProductID:  dto.ProductID,
		Quantity:   dto.Quantity,
		TotalPrice: dto.TotalPrice,
	}
}

func (m *OrderedProduct) ToDTO() *models.OrderedProduct {
	return &models.OrderedProduct{
		OrderID:     m.OrderID,
		ProductID:   m.ProductID,
		ProductName: m.ProductName,
		InStock:     m.InStock,
		Quantity:    m.Quantity,
		TotalPrice:  m.TotalPrice,
	}
}
