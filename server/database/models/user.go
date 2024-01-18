package models

import (
	"estore-backend/server/models"
)

// User user
//
// swagger:model user
type User struct {

	// date created
	// Read Only: true
	DateCreated int64 `json:"dateCreated,omitempty"`

	// date updated
	DateUpdated int64 `json:"dateUpdated,omitempty"`

	// email
	// Required: true
	Email string `json:"email"`

	// id
	// Read Only: true
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	// name
	Name string `json:"name,omitempty"`
}

func NewUserFrom(dto *models.User) *User {
	return &User{
		DateCreated: dto.DateCreated,
		DateUpdated: dto.DateUpdated,
		Email:       *dto.Email,
		ID:          dto.ID,
		Name:        dto.Name,
	}
}

func (m *User) ToDTO() *models.User {
	return &models.User{
		DateCreated: m.DateCreated,
		DateUpdated: m.DateUpdated,
		Email:       &m.Email,
		ID:          m.ID,
		Name:        m.Name,
	}
}
