package models

type ProductToCategory struct {
	ProductID  int64     `bun:",pk"`
	Product    *Product  `bun:"rel:belongs-to,join:product_id=id"`
	CategoryID int64     `bun:",pk"`
	Category   *Category `bun:"rel:belongs-to,join:category_id=id"`
}
