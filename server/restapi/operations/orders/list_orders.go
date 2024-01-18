// Code generated by go-swagger; DO NOT EDIT.

package orders

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"estore-backend/server/models"
)

// ListOrdersHandlerFunc turns a function with the right signature into a list orders handler
type ListOrdersHandlerFunc func(ListOrdersParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn ListOrdersHandlerFunc) Handle(params ListOrdersParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// ListOrdersHandler interface for that can handle valid list orders params
type ListOrdersHandler interface {
	Handle(ListOrdersParams, *models.Principal) middleware.Responder
}

// NewListOrders creates a new http.Handler for the list orders operation
func NewListOrders(ctx *middleware.Context, handler ListOrdersHandler) *ListOrders {
	return &ListOrders{Context: ctx, Handler: handler}
}

/*
	ListOrders swagger:route GET /orders orders listOrders

List orders
*/
type ListOrders struct {
	Context *middleware.Context
	Handler ListOrdersHandler
}

func (o *ListOrders) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewListOrdersParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *models.Principal
	if uprinc != nil {
		principal = uprinc.(*models.Principal) // this is really a models.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
