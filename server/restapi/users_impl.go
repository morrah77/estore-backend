package restapi

import (
	"context"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/user"
	"estore-backend/server/restapi/operations/users"
	"github.com/go-openapi/errors"
	"github.com/uptrace/bun"
	"time"
)

func addUser(params *users.AddUserParams, principal *models.Principal) errors.Error {
	Logger.Debug("\naddUser: request params: %s\nprincipal: %s\n", params.HTTPRequest, principal)
	item := params.Body
	return addDBUser(item)
}

func addDBUser(item *models.User) errors.Error {
	Logger.Debug("\naddDBUser: item: %s\n", item)
	if item == nil {
		return errors.New(500, "DB item cannot be nil!")
	}

	dbModel := dbModels.NewUserFrom(item)
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateCreated = nowUnixEpoch
	dbModel.DateUpdated = nowUnixEpoch
	query := db.NewInsert().Model(dbModel).ExcludeColumn("id")
	Logger.Debug("Built the query %s\n", query)

	_, err := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not add user %s!\n", err, item)
	}
	return nil
}

func updateUser(params *user.EditUserParams, principal *models.Principal) errors.Error {
	Logger.Debug("\nupdateUser: request params: %s\nID:%d\nprincipal: %s\n", params.HTTPRequest, params.ID, principal)
	err := isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return err
	}

	item := params.Body
	if item == nil {
		return errors.New(500, "Empty user!")
	}
	dbModel := dbModels.NewUserFrom(item)
	nowUnixEpoch := time.Now().In(time.UTC).Unix()
	dbModel.DateUpdated = nowUnixEpoch
	query := db.NewUpdate().Where("id = ?", params.ID).
		Model(dbModel).ExcludeColumn("id").ExcludeColumn("date_created")
	Logger.Debug("Built the query %s\n", query)

	_, sqlErr := query.Exec(context.Background())
	if err != nil {
		return errors.New(500, "ERROR %v: Could not update user %d!\n", sqlErr, params.ID)
	}

	return nil
}

func deleteUser(params *user.DeleteUserParams, principal *models.Principal) errors.Error {
	Logger.Debug("\ndeleteUser: request params: %s\nID:%d\nprincipal: %s\n", params.HTTPRequest, params.ID, principal)
	err := isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return err
	}

	query := db.NewDelete().TableExpr("users").Where("id = ?", params.ID)
	Logger.Debug("Built the query %s\n", query)

	_, sqlErr := query.Exec(context.Background())
	if sqlErr != nil {
		return errors.New(500, "ERROR %v: Could not delete user %d!\n", sqlErr, params.ID)
	}

	return nil
}

func getUser(params *user.GetUserParams, principal *models.Principal) (result *models.User, err errors.Error) {
	Logger.Debug("\ngetUser: request params: %s\nID:%d\nprincipal: %s\n", params.HTTPRequest, params.ID, principal)
	err = isPrincipalOwnerOrAdmin(principal, params.ID)
	if err != nil {
		return nil, err
	}

	dbModel := new(dbModels.User)

	query := db.NewSelect().Model(dbModel).Where("id = ?", params.ID)
	Logger.Debug("Built the query %s\n", query)

	sqlErr := query.Scan(context.Background())
	if sqlErr != nil {
		return nil, errors.New(500, "ERROR %v: Could not find user %d!\n", sqlErr, params.ID)
	}

	result = dbModel.ToDTO()
	return result, nil
}

func getOwnUserInfo(params *user.GetOwnUserParams, principal *models.Principal) (result *models.User, err errors.Error) {
	Logger.Debug("\ngetOwnUserInfo: principal: %s\nprincipal.User.ID: %d\n", principal, principal.User.ID)
	err = isPrincipalOwnerOrAdmin(principal, principal.User.ID)
	if err != nil {
		return nil, err
	}

	dbModel := dbModels.User{
		ID:    principal.User.ID,
		Name:  principal.User.Name,
		Email: *principal.User.Email,
	}

	result = dbModel.ToDTO()
	return result, nil
}

func getUserByEmail(email string) (result *models.User, err errors.Error) {
	dbModel := new(dbModels.User)

	query := db.NewSelect().Model(dbModel).Where("email = ?", email)
	Logger.Debug("Built the query %s\n", query)

	sqlErr := query.Scan(context.Background())
	if sqlErr != nil {
		return nil, errors.New(500, "ERROR %v: Could not find user %d!\n", sqlErr, email)
	}

	result = dbModel.ToDTO()
	return result, nil
}

func allUsers(params *users.ListUsersParams, principal *models.Principal) (result []*models.User, err errors.Error) {
	Logger.Debug("\nallUsers:\nrequest %s\nprincipal %s\n", params.HTTPRequest, principal)

	dbModel := make([]*dbModels.User, 0)

	query := db.NewSelect().Model(&dbModel)
	if params.Limit != nil {
		query.Limit(int(*params.Limit))
	}
	if params.Offset != nil {
		query.Offset(int(*params.Offset))
	}
	if params.Search != nil && len(*params.Search) > 0 {
		query.Where("? LIKE ? OR ? LIKE ?", bun.Ident("name"), "%"+*params.Search+"%", bun.Ident("email"), "%"+*params.Search+"%")
	}
	Logger.Debug("Built the query %s\n%s\n", query, query.QueryBuilder())

	sqlErr := query.Scan(context.Background())
	if sqlErr != nil {
		return nil, errors.New(500, "ERROR %v: Could not find users matching %s!\n", sqlErr, params)
	}

	result = make([]*models.User, len(dbModel))
	for i, m := range dbModel {
		result[i] = m.ToDTO()
	}
	return result, nil
}
