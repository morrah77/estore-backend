package restapi

import (
	"context"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/order"
	"estore-backend/server/restapi/operations/orders"
	"fmt"
	"github.com/go-openapi/errors"
	"github.com/uptrace/bun"
	"log"
	"time"
)

func addOrder(params *orders.AddOrderParams, principal *models.Principal) (*models.Order, errors.Error) {
	isAdmin, err := isPrincipalAdmin(principal)
	if err != nil {
		return nil, err
	}
	var item = params.Body
	Logger.Debug("adding item %v\n")
	if item == nil {
		return nil, errors.New(400, "DB item cannot be nil!")
	}
	if item.Products == nil || len(item.Products) <= 0 {
		return nil, errors.New(400, "Order product list cannot be empty!")
	}

	// TODO consider using bun struct tags ",default:current_timestamp" for type time.Time
	dbModel := dbModels.NewOrderFrom(item)

	// admins can add orders for any user; if the userID is zero, then user ID comes from the principal
	// ordinarys user can add orders for themselves only; if the user ID is empty, then see above
	if !isAdmin && item.UserID > 0 && item.UserID != principal.User.ID {
		return nil, errors.New(403, "Attempt to create order for non-own user!")
	}

	if item.UserID < 1 {
		Logger.Debug("Setting user ID %d for order %s", principal.User.ID, dbModel)
		dbModel.UserID = principal.User.ID
	}
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateCreated = nowUnixEpoch
	dbModel.DateUpdated = nowUnixEpoch
	dbModel.Status = "created"

	query := db.NewInsert().Model(dbModel).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	res, sqlErr := query.Exec(context.Background())
	if sqlErr != nil {
		return nil, errors.New(500, "ERROR %v: Could not add order %s!\n", sqlErr, item)
	}

	id, sqlErr := res.LastInsertId()
	if sqlErr != nil {
		log.Printf("ERROR %v: Could not find last insert ID for order %s!", sqlErr, item)
		return nil, errors.New(500, "ERROR: Could not add order %s!\n", item)
	}

	dbModel.ID = id

	totalPrice, sqlErr := updateOrderedProductsIfNeeded(dbModel)
	// it looks like bun does not support DB transactions in general :(
	if sqlErr != nil {
		Logger.Error("Could not add order %d products %s!\nERROR: %s (%v)\n",
			id, dbModel.Products, err.Error(), err)
		delQuery := db.NewDelete().Model(dbModel).Where("id = ?", id)
		res, sqlDelErr := delQuery.Exec(context.Background())
		if sqlDelErr != nil {
			Logger.Error("ERROR %v: Could not add order %d products %s!\n"+
				"ERROR: %v: Could not delete order %s (result: %s)\n",
				sqlErr, id, dbModel.Products, sqlDelErr, dbModel, res)
		}
		return nil, errors.New(500, "ERROR: Could not add order %d products %s!",
			id, dbModel.Products)
	}
	if *totalPrice != *dbModel.TotalPrice {
		*dbModel.TotalPrice = *totalPrice
		updQuery := db.NewUpdate().Model(dbModel).ExcludeColumn("id").ExcludeColumn("date_created").
			ExcludeColumn("status").Column("total_price").Where("id = ?", dbModel.ID)
		_, sqlErr = updQuery.Exec(context.Background())
		if sqlErr != nil {
			Logger.Error("ERROR %v:\nCould not update order %d total price to %f\n", sqlErr, id, totalPrice)
			return nil, errors.New(500, "ERROR: Could not update order %d total price to %f", id, totalPrice)
		}
	}

	return dbModel.ToDTO(), nil
}

func updateOrder(params *order.EditOrderParams, principal *models.Principal) errors.Error {
	err := isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return err
	}

	var item = params.Body
	Logger.Debug("Updating item %s\n", item)
	if item == nil {
		return errors.New(400, "DB item cannot be nil!")
	}
	if item.Products == nil || len(item.Products) <= 0 {
		return errors.New(400, "Order product list cannot be empty!")
	}

	dbModel := dbModels.NewOrderFrom(item)
	if dbModel.ID <= 0 {
		dbModel.ID = params.ID
	}

	isAdmin, err := isPrincipalAdmin(principal)
	if err != nil {
		return err
	}
	if !isAdmin && item.UserID > 0 && item.UserID != principal.User.ID {
		return errors.New(403, "Attempt to update order of non-own user!")
	}
	if item.UserID < 1 {
		Logger.Debug("Setting user ID %d for order %s", principal.User.ID, dbModel)
		dbModel.User = dbModels.NewUserFrom(principal.User)
		dbModel.UserID = principal.User.ID
	}

	totalPrice, sqlErr := updateOrderedProductsIfNeeded(dbModel)
	if sqlErr != nil {
		Logger.Error("ERROR %v: Could not update order %d products!\n", sqlErr, params.ID)
		return errors.New(500, "ERROR: Could not update order %d products!", params.ID)
	}
	Logger.Debug("Calculated total price: %f", *totalPrice)
	*dbModel.TotalPrice = *totalPrice
	Logger.Debug("Assigned new order total price: %f", *dbModel.TotalPrice)

	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateUpdated = nowUnixEpoch
	Logger.Debug("Will update order record %s", *dbModel)

	query := db.NewUpdate().Where("id = ?", params.ID).
		Model(dbModel).ExcludeColumn("id").
		ExcludeColumn("date_created").
		ExcludeColumn("status")
	log.Printf("Built the query %s\n", query)

	_, sqlErr = query.Exec(context.Background())
	if sqlErr != nil {
		Logger.Error("ERROR %v: Could not update order %d!\n", sqlErr, params.ID)
		return errors.New(500, "ERROR: Could not update order %d!", params.ID)
	}

	return nil
}

func updateDBOrder(dbModel *dbModels.Order) (*dbModels.Order, errors.Error) {
	if dbModel == nil {
		return nil, errors.New(500, "Empty order!")
	}
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateUpdated = nowUnixEpoch
	query := db.NewUpdate().Where("id = ?", dbModel.ID).
		Model(dbModel).ExcludeColumn("id").ExcludeColumn("date_created")
	Logger.Debug("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		Logger.Error("ERROR %v (%s): Could not update order %d!\n", err, err, dbModel.ID)
		return nil, errors.New(500, "Could not update order %d!", dbModel.ID)
	}

	return dbModel, nil
}

func deleteOrder(params *order.DeleteOrderParams, principal *models.Principal) errors.Error {
	err := isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return err
	}

	// either SQLite DB driver do not actually support cascade deletion by foreign keys
	// despite of declared PRAGMA foreign_keys,
	// or bun ORM forms foreign key expression wrong, but the FK constraint with ON DELETE CASCADE
	// presents in the DB table create expression but do not work...
	query := db.NewDelete().TableExpr("ordered_products").Where("order_id = ?", params.ID)
	Logger.Debug("Built the query %s\n", query)
	_, sqlErr := query.Exec(context.Background())
	if sqlErr != nil {
		Logger.Error("ERROR %v: Could not delete order %d products!\n", sqlErr, params.ID)
	}

	query = db.NewDelete().TableExpr("orders").Where("id = ?", params.ID)
	Logger.Debug("Built the query %s\n", query)

	_, sqlErr = query.Exec(context.Background())
	if sqlErr != nil {
		Logger.Error("ERROR %v: Could not delete order %d!\n", sqlErr, params.ID)
		return errors.New(500, "ERROR: Could not delete order %d!", params.ID)
	}

	return nil
}

func getOrder(params *order.GetOrderParams, principal *models.Principal) (result *models.Order, err errors.Error) {
	err = isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return nil, err
	}
	isAdmin, err := isPrincipalAdmin(principal)
	if err != nil {
		return nil, err
	}

	dbModel, err := getOrderFromDB(params.ID, isAdmin, principal.User.ID)
	if err != nil {
		return nil, err
	}

	result = dbModel.ToDTO()
	return result, nil
}

func getOrderFromDB(orderId int64, isAdmin bool, userId int64) (*dbModels.Order, errors.Error) {
	dbModel := new(dbModels.Order)

	query := db.NewSelect().Model(dbModel)
	query = queryOrder(query, isAdmin, userId)
	query.Where("id = ?", orderId)
	Logger.Debug("Built the query %s\n", query)

	sqlErr := query.Scan(context.Background())
	if sqlErr != nil {
		return nil, errors.New(500, "ERROR %v: Could not find order %d!\n", sqlErr, orderId)
	}

	Logger.Debug("Fetched an order %s\n", dbModel)
	return dbModel, nil
}

func allOrders(params *orders.ListOrdersParams, principal *models.Principal) (result []*models.Order, err errors.Error) {
	isAdmin, err := isPrincipalAdmin(principal)
	if err != nil {
		return nil, err
	}
	dbModel := make([]*dbModels.Order, 0)

	query := db.NewSelect().Model(&dbModel)
	query = queryOrder(query, isAdmin, principal.User.ID)
	if params.Limit != nil {
		query.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		query.Offset(int(*params.Offset))
	}
	if params.OrderBy != nil {
		orderBy := "ASC"
		if params.Order != nil {
			orderBy = *params.Order
		}
		query.Order(fmt.Sprintf("%s %s", *params.OrderBy, orderBy))
	}
	Logger.Debug("Built the query %s\n", query)

	sqlErr := query.Scan(context.Background())
	if sqlErr != nil {
		Logger.Error("ERROR %v: Could not find orders matching %s!\n", sqlErr, params)
		return nil, errors.New(500, "ERROR: Could not find orders matching %s!", params)
	}

	Logger.Debug("Fetched orders %s\n", dbModel)
	result = make([]*models.Order, len(dbModel))
	for i, m := range dbModel {
		Logger.Debug("Processing order %s, user ID %d\n", m, m.UserID)
		result[i] = m.ToDTO()
		Logger.Debug("Converted to DTO %s, user ID %d\n", result[i], result[i].UserID)
	}
	return result, nil
}

func queryOrder(query *bun.SelectQuery, isAdmin bool, userID int64) *bun.SelectQuery {
	query.Relation("Products",
		func(q *bun.SelectQuery) *bun.SelectQuery {
			bunQuery := q.Join("JOIN products as p on ordered_product.product_id = p.id ").
				ColumnExpr("ordered_product.*, p.title as product_name, (p.number_in_stock is not null) as in_stock")
			Logger.Debug("Built the query %s\n", bunQuery)
			return bunQuery
		})
	if !isAdmin {
		query.Where("user_id = ?", userID)
	}
	return query
}

func updateOrderedProductsIfNeeded(order *dbModels.Order) (*float64, error) {
	Logger.Debug("Updating ordered products for order %s\n%s\n", order, *order)
	delQuery := db.NewDelete().Table("ordered_products").Where("order_id = ?", order.ID)
	_, err := delQuery.Exec(context.Background())
	if err != nil {
		Logger.Error("Could not delete order %d products!\n", order.ID)
	}
	productIDs := make([]int64, len(order.Products))
	for i, product := range order.Products {
		if product.OrderID != order.ID {
			product.OrderID = order.ID
		}
		productIDs[i] = *product.ProductID
	}
	var actualProducts []*dbModels.Product = make([]*dbModels.Product, 0)
	selQuery := db.NewSelect().Model(&actualProducts).Where("id in (?)", bun.In(productIDs))
	Logger.Debug("Built the query %s\n", selQuery)
	err = selQuery.Scan(context.Background())
	if err != nil {
		Logger.Error("ERROR %v:\nCould not find products for order %s!", err, order)
		return nil, errors.New(500, "Could not find products for order %s!", order)
	}
	Logger.Debug("Actual products: %s\n", actualProducts)
	var isFound bool
	var totalPrice float64 = 0
	var productTotalPrice *float64 = nil
	for _, product := range order.Products {
		isFound = false
		for _, actualProduct := range actualProducts {
			if *product.ProductID == actualProduct.ID {
				productTotalPrice = CalculateProductTotalPrice(actualProduct, product.Quantity)
				product.TotalPrice = productTotalPrice
				Logger.Debug("Found product %d (%s); total price is %f, assigned total price is %f\n",
					actualProduct.ID, *productTotalPrice, actualProduct.Price)
				isFound = true
				totalPrice += *productTotalPrice
				break
			}
		}
		if !isFound {
			Logger.Debug("Not found product for %d", product.ProductID)
			product.TotalPrice = nil
		}
	}

	query := db.NewInsert().Model(&order.Products).ExcludeColumn("id")
	Logger.Debug("Built the query %s\n", query)

	_, err = query.Exec(context.Background())
	if err != nil {
		Logger.Error("ERROR %v: Could not update order %d products %s!", err, order.ID, order.Products)
		return nil, errors.New(500,
			"ERROR %s: Could not update order %d products %s!", err.Error(), order.ID, order.Products)
	} else {
		Logger.Debug("Updated order %d products %s", order.ID, order.Products)
	}
	Logger.Debug("Returning order total price: %f", totalPrice)
	return &totalPrice, nil
}

func CalculateProductTotalPrice(actualProduct *dbModels.Product, quantity *int64) *float64 {
	res := actualProduct.Price * float64(*quantity)
	return &res
}
