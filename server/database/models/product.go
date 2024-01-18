package models

import (
	"estore-backend/server/models"
)

type Product struct {

	// category ids
	Categories []Category `json:"categories" bun:"m2m:product_to_categories,join:Product=Category"`

	// description
	// Required: true
	// Min Length: 1
	Description *string `json:"description"`

	// id
	// Read Only: true
	ID int64 `json:"id,omitempty" bun:",pk,autoincrement,unique"`

	// images
	Images []string `json:"images"`

	// number in stock
	NumberInStock int64 `json:"numberInStock,omitempty"`

	// price
	Price float64 `json:"price,omitempty"`

	// title
	// Required: true
	// Min Length: 1
	Title *string `json:"title"`
}

func NewProductFrom(dto *models.Product) *Product {
	return &Product{
		Categories:    CategoriesFrom(dto.CategoryIds),
		Description:   dto.Description,
		ID:            dto.ID,
		Images:        dto.Images,
		NumberInStock: dto.NumberInStock,
		Price:         dto.Price,
		Title:         dto.Title,
	}
}

func CategoriesFrom(categoryIds []int64) []Category {
	if categoryIds == nil {
		return nil
	}
	result := make([]Category, len(categoryIds))
	for i, id := range categoryIds {
		result[i] = Category{ID: id}
	}
	return result
}

func (m *Product) ToDTO() *models.Product {
	return &models.Product{
		CategoryIds:   CategoryIdsFrom(m.Categories),
		Description:   m.Description,
		ID:            m.ID,
		Images:        m.Images,
		NumberInStock: m.NumberInStock,
		Price:         m.Price,
		Title:         m.Title,
	}
}

func CategoryIdsFrom(categories []Category) []int64 {
	if categories == nil {
		return nil
	}
	result := make([]int64, len(categories))
	for i, c := range categories {
		result[i] = c.ID
	}
	return result
}
