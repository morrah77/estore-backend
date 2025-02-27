// Code generated by go-swagger; DO NOT EDIT.

package products

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"estore-backend/server/models"
)

// AddProductHandlerFunc turns a function with the right signature into a add product handler
type AddProductHandlerFunc func(AddProductParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn AddProductHandlerFunc) Handle(params AddProductParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// AddProductHandler interface for that can handle valid add product params
type AddProductHandler interface {
	Handle(AddProductParams, *models.Principal) middleware.Responder
}

// NewAddProduct creates a new http.Handler for the add product operation
func NewAddProduct(ctx *middleware.Context, handler AddProductHandler) *AddProduct {
	return &AddProduct{Context: ctx, Handler: handler}
}

/*
	AddProduct swagger:route POST /products products addProduct

Add product
*/
type AddProduct struct {
	Context *middleware.Context
	Handler AddProductHandler
}

func (o *AddProduct) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewAddProductParams()
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
