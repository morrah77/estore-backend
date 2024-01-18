package restapi

import (
	"context"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/categories"
	"github.com/go-openapi/errors"
	"github.com/uptrace/bun"
	"log"
)

func addCategory(item *models.Category) error {
	log.Printf("adding item %s\n%s\n%s", item, &item, *item)
	if item == nil {
		return errors.New(500, "DB item cannot be nil!")
	}

	dbModel := dbModels.NewCategoryFrom(item)
	query := db.NewInsert().Model(dbModel).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not add category %s!\n", err, item)
	}
	return nil
}

func updateCategory(id int64, item *models.Category) error {
	if item == nil {
		return errors.New(500, "Empty category!")
	}
	dbModel := dbModels.NewCategoryFrom(item)
	query := db.NewUpdate().Where("id = ?", id).Model(dbModel).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not update category %d!\n", err, id)
	}

	return nil
}

func deleteCategory(id int64) error {
	query := db.NewDelete().TableExpr("categories").Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not delete category %d!\n", err, id)
	}

	return nil
}

func getCategory(id int64) (result *models.Category, err error) {
	dbModel := new(dbModels.Category)

	query := db.NewSelect().Model(dbModel).Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %v: Could not find category %d!\n", err, id)
	}

	result = dbModel.ToDTO()
	return result, nil
}

func allCategories(params *categories.ListCategoriesParams) (result []*models.Category, err error) {
	dbModel := make([]*dbModels.Category, 0)

	query := db.NewSelect().Model(&dbModel)
	if params.Limit != nil {
		query.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		query.Offset(int(*params.Offset))
	}
	if params.Search != nil && len(*params.Search) > 0 {
		query.Where("? LIKE ? OR ? LIKE ?", bun.Ident("title"), "%"+*params.Search+"%", bun.Ident("description"), "%"+*params.Search+"%")
	}
	log.Printf("Built the query %s\n%s\n", query, query.QueryBuilder())

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %v: Could not find categories matching %s!\n", err, params)
	}

	result = make([]*models.Category, len(dbModel))
	for i, m := range dbModel {
		result[i] = m.ToDTO()
	}
	return result, nil
}
