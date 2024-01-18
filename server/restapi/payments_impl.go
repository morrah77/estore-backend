package restapi

import (
	"context"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/payments"
	"github.com/go-openapi/errors"
	"log"
	"time"
)

func addPayment(item *models.Payment) error {
	log.Printf("adding item %s\n%s\n%s", item, &item, *item)
	if item == nil {
		return errors.New(500, "DB item cannot be nil!")
	}

	dbModel := dbModels.NewPaymentFrom(item)
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateCreated = nowUnixEpoch
	dbModel.DateUpdated = nowUnixEpoch
	query := db.NewInsert().Model(dbModel).ExcludeColumn("id")
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not add payment %s!\n", err, item)
	}
	return nil
}

func updatePayment(id int64, item *models.Payment) error {
	if item == nil {
		return errors.New(500, "Empty order!")
	}
	dbModel := dbModels.NewPaymentFrom(item)
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateUpdated = nowUnixEpoch
	query := db.NewUpdate().Where("id = ?", id).
		Model(dbModel).ExcludeColumn("id").ExcludeColumn("date_created")
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not update payment %d!\n", err, id)
	}

	return nil
}

func deletePayment(id int64) error {
	query := db.NewDelete().TableExpr("orders").Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not delete payment %d!\n", err, id)
	}

	return nil
}

func getPayment(id int64) (result *models.Payment, err error) {
	dbModel := new(dbModels.Payment)

	query := db.NewSelect().Model(dbModel).Where("id = ?", id)
	log.Printf("Built the query %s\n", query)

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %v: Could not find payment %d!\n", err, id)
	}

	result = dbModel.ToDTO()
	return result, nil
}

func allPayments(params *payments.ListPaymentsParams) (result []*models.Payment, err error) {
	dbModel := make([]*dbModels.Payment, 0)

	query := db.NewSelect().Model(&dbModel)
	if params.Limit != nil {
		query.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		query.Offset(int(*params.Offset))
	}
	log.Printf("Built the query %s\n%s\n", query, query.QueryBuilder())

	err = query.Scan(context.Background())
	if err != nil {
		return nil, errors.New(500, "ERROR %v: Could not find payments matching %s!\n", err, params)
	}

	result = make([]*models.Payment, len(dbModel))
	for i, m := range dbModel {
		result[i] = m.ToDTO()
	}
	return result, nil
}
