package models

import (
	"estore-backend/server/models"
)

type Category struct {

	// description
	// Min Length: 1
	Description string `json:"description,omitempty"`

	// id
	// Read Only: true
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	// title
	// Required: true
	// Min Length: 1
	Title *string `json:"title"`
}

func NewCategoryFrom(dto *models.Category) *Category {
	return &Category{
		Description: dto.Description,
		ID:          dto.ID,
		Title:       dto.Title,
	}
}

func (m *Category) ToDTO() *models.Category {
	return &models.Category{
		Description: m.Description,
		ID:          m.ID,
		Title:       m.Title,
	}
}
