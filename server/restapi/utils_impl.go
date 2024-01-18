package restapi

import (
	"estore-backend/server/models"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"net/http"
)

func isPrincipalOwnerOrAdmin(principal *models.Principal, id int64) errors.Error {
	Logger.Debug("\nisPrincipalOwnerOrAdmin: principal from context: %s\n", principal)
	if principal.User == nil {
		return errors.New(403, "ERROR : Current user data do not contain registered user data while accessing data of the user user %d! Denying...\n", id)
	}
	if principal.User.ID != id && principal.ACLRole != "admin" {
		return errors.New(403, "ERROR : Current user data id %d is not equal to requested user data ID or current user role '%s' is not 'admin' while accessing data of the user user %d! Denying...\n",
			principal.User.ID, principal.ACLRole, id)
	}
	return nil
}

func isPrincipalAdmin(principal *models.Principal) (bool, errors.Error) {
	if principal.User == nil || principal.User.ID < 1 {
		return false, errors.New(403, "Unregistered users are forbidden!")
	}
	if principal.ACLRole == "admin" {
		return true, nil
	}
	return false, nil
}

func isContextPrincipalOwnerOrAdmin(request *http.Request, id int64) errors.Error {
	currentUserInfo := middleware.SecurityPrincipalFrom(request)
	Logger.Debug("\nisContextPrincipalOwnerOrAdmin: interface principal from context: %s\n", currentUserInfo)
	info, ok := currentUserInfo.(*models.Principal)
	if !ok {
		return errors.New(500, "ERROR : Could not fetch current user data while accessing data of the user user %d! Denying...\n", id)
	}
	Logger.Debug("\nisContextPrincipalOwnerOrAdmin: Type-asserted principal from context: %s\n", info)
	if info.User == nil {
		return errors.New(403, "ERROR : Current user data do not contain registered user data while accessing data of the user user %d! Denying...\n", id)
	}
	if info.User.ID != id && info.ACLRole != "admin" {
		return errors.New(403, "ERROR : Current user data id %d is not equal to requested user data ID or current user role '%s' is not 'admin' while accessing data of the user user %d! Denying...\n", info.User.ID, info.ACLRole, id)
	}
	return nil
}
