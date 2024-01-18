package restapi

import (
	"context"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/products"
	"github.com/go-openapi/errors"
	"github.com/uptrace/bun"
	"log"
)

func addProduct(item *models.Product) error {
	log.Printf("adding item %s\n%s\n%s", item, &item, *item)
	if item == nil {
		return errors.New(500, "DB item cannot be nil!")
	}

	product := dbModels.NewProductFrom(item)
	query := db.NewInsert().Model(product).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	res, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %s: Could not add product %s!", err.Error(), item)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("ERROR %v: Could not find last insert ID for product %s!", err, item)
		return nil
	}

	err = updateProductCategoriesIfNeeded(id, product)
	if err != nil {
		log.Printf("ERROR %v: Could not update product %d categories %s!", err.Error(), id, product.Categories)
	}

	return nil
}

func updateProduct(id int64, item *models.Product) error {
	if item == nil {
		return errors.New(500, "Empty product!")
	}
	product := dbModels.NewProductFrom(item)
	query := db.NewUpdate().Where("id = ?", id).Model(product).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %s: Could not update product %d!", err.Error(), id)
	}

	err = updateProductCategoriesIfNeeded(id, product)
	if err != nil {
		log.Printf("ERROR %v: Could not update product %d categories %s!", err.Error(), product.Categories)
	}

	return nil
}

func updateProductCategoriesIfNeeded(id int64, product *dbModels.Product) error {
	if product.Categories == nil {
		delQery := db.NewDelete().TableExpr("product_to_categories").Where("product_id = ?", id)
		log.Printf("Built the query %s\n", delQery)

		_, err := delQery.Exec(context.Background())
		if err != nil {
			log.Printf("ERROR %v: Could not clean product %d categories %s!", err, id, product.Categories)
			return errors.New(500,
				"ERROR %s: Could not clean product %d categories %s!", err.Error(), id, product.Categories)
		}
		return nil
	}
	if len(product.Categories) > 0 {
		pcModel := make([]dbModels.ProductToCategory, len(product.Categories))
		for i, c := range product.Categories {
			pcModel[i] = dbModels.ProductToCategory{
				ProductID:  id,
				Product:    product,
				CategoryID: c.ID,
				Category:   &c,
			}
		}
		catQuery := db.NewInsert().Model(&pcModel)
		log.Printf("Built the query %s\n", catQuery)

		_, err := catQuery.Exec(context.Background())
		if err != nil {
			log.Printf("ERROR %v: Could not update product %d categories %s!", err, id, product.Categories)
			return errors.New(500,
				"ERROR %s: Could not update product %d categories %s!", err.Error(), id, product.Categories)
		} else {
			log.Printf("Updated product %d categories: %s", id, product.Categories)
		}
	}
	return nil
}

func deleteProduct(id int64) error {
	query := db.NewDelete().TableExpr("products").Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %s: Could not delete product %d!", err.Error(), id)
	}

	return nil
}

func getProduct(id int64) (result *models.Product, err error) {
	product := new(dbModels.Product)

	query := db.NewSelect().Model(product).Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %s: Could not find product %d!", err.Error(), id)
	}

	result = product.ToDTO()
	return result, nil
}

func allProducts(params *products.GetProductsParams) (result []*models.Product, err error) {
	queryResult := make([]*dbModels.Product, 0)

	query := db.NewSelect().Model(&queryResult).Relation("Categories", func(q *bun.SelectQuery) *bun.SelectQuery {
		bunQuery := q.Column("id")
		log.Printf("Built the query %s\n%s\n", bunQuery, bunQuery.QueryBuilder())
		return bunQuery
	})

	if params.CategoryIds != nil && len(params.CategoryIds) > 0 {
		query.Distinct().Join("JOIN product_to_categories as pk on product.id = pk.product_id").
			Where("pk.product_id is not null").
			Where("pk.category_id in (?)", bun.In(params.CategoryIds))
	}

	if params.Limit != nil {
		query.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		query.Offset(int(*params.Offset))
	}
	if params.Search != nil && len(*params.Search) > 0 {
		//  OR price LIKE %s
		query.Where("? LIKE ? OR ? LIKE ?", bun.Ident("title"), "%"+*params.Search+"%", bun.Ident("description"), "%"+*params.Search+"%")
	}
	log.Printf("Built the query %s\n%s\n", query, query.QueryBuilder())

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %s: Could not find product matching %s!", err.Error(), params)
	}

	result = make([]*models.Product, len(queryResult))
	for i, m := range queryResult {
		result[i] = m.ToDTO()
	}
	return result, nil
}
