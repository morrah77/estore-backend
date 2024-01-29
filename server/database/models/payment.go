package models

import (
	"estore-backend/server/models"
)

type Payment struct {

	// amount
	// Required: true
	Amount *float64 `json:"amount"`

	// date created
	// Read Only: true
	DateCreated int64 `json:"dateCreated,omitempty"`

	// date updated
	DateUpdated int64 `json:"dateUpdated,omitempty"`

	// id
	// Read Only: true
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	OrderID int64

	Order *Order `json:"order,omitempty" bun:"rel:belongs-to,join:order_id=id"`

	// status
	Status string `json:"status,omitempty"`

	UserID int64

	User *User `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`

	CheckoutSessionID string `json:"checkout_session_id"`

	PaymentIntentId string `json:"payment_intent_id"`
}

func NewPaymentFrom(dto *models.Payment) *Payment {
	return &Payment{
		Amount:      dto.Amount,
		DateCreated: dto.DateCreated,
		DateUpdated: dto.DateUpdated,
		ID:          dto.ID,
		Order:       &Order{ID: *dto.OrderID},
		Status:      dto.Status,
		User:        &User{ID: dto.UserID},
	}
}

func (m *Payment) ToDTO() *models.Payment {
	return &models.Payment{
		Amount:      m.Amount,
		DateCreated: m.DateCreated,
		DateUpdated: m.DateUpdated,
		ID:          m.ID,
		OrderID:     &m.Order.ID,
		Status:      m.Status,
		UserID:      m.User.ID,
	}
}
