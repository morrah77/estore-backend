package models

import (
	"estore-backend/server/models"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type Order struct {

	// date created
	// Read Only: true
	DateCreated int64 `json:"dateCreated,omitempty"`

	// date updated
	DateUpdated int64 `json:"dateUpdated,omitempty"`

	// delivery info
	// Required: true
	// Min Length: 10
	DeliveryInfo *string `json:"deliveryInfo"`

	// id
	// Read Only: true
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	// products
	// Required: true
	Products []*OrderedProduct `json:"products" bun:"rel:has-many,join:id=order_id"`

	// status
	Status string `json:"status,omitempty"`

	// total price
	// Required: true
	TotalPrice *float64 `json:"totalPrice"`

	UserID int64

	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

var _ bun.BeforeCreateTableHook = (*Order)(nil)

func (m *Order) BeforeCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	query.ForeignKey(`("user_id") REFERENCES "users" ("id")`) // ON DELETE CASCADE`)
	return nil
}

func NewOrderFrom(dto *models.Order) *Order {
	return &Order{
		DateCreated:  dto.DateCreated,
		DateUpdated:  dto.DateUpdated,
		DeliveryInfo: dto.DeliveryInfo,
		ID:           dto.ID,
		Products:     OrderedProductsFromOrderedProductDTOs(dto.Products),
		Status:       dto.Status,
		TotalPrice:   dto.TotalPrice,
		UserID:       dto.UserID,
		User:         &User{ID: dto.UserID},
	}
}

func OrderedProductsFromOrderedProductDTOs(orderedProducts []*models.OrderedProduct) []*OrderedProduct {
	if orderedProducts == nil {
		return nil
	}
	result := make([]*OrderedProduct, len(orderedProducts))
	for i, product := range orderedProducts {
		result[i] = NewOrderedProductFrom(product)
	}
	return result
}

func (m *Order) ToDTO() *models.Order {
	return &models.Order{
		DateCreated:  m.DateCreated,
		DateUpdated:  m.DateUpdated,
		DeliveryInfo: m.DeliveryInfo,
		ID:           m.ID,
		Products:     OrderedProductsDTOsFromOrderedProducts(m.Products),
		Status:       m.Status,
		TotalPrice:   m.TotalPrice,
		UserID:       m.UserID,
	}
}

func OrderedProductsDTOsFromOrderedProducts(orderedProducts []*OrderedProduct) []*models.OrderedProduct {
	if orderedProducts == nil {
		return nil
	}
	result := make([]*models.OrderedProduct, len(orderedProducts))
	for i, product := range orderedProducts {
		result[i] = product.ToDTO()
	}
	return result
}

func UserIdFromDBUser(user *User) int64 {
	if user == nil {
		return 0
	}
	return user.ID
}
