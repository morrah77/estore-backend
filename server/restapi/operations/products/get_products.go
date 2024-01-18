// Code generated by go-swagger; DO NOT EDIT.

package products

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetProductsHandlerFunc turns a function with the right signature into a get products handler
type GetProductsHandlerFunc func(GetProductsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetProductsHandlerFunc) Handle(params GetProductsParams) middleware.Responder {
	return fn(params)
}

// GetProductsHandler interface for that can handle valid get products params
type GetProductsHandler interface {
	Handle(GetProductsParams) middleware.Responder
}

// NewGetProducts creates a new http.Handler for the get products operation
func NewGetProducts(ctx *middleware.Context, handler GetProductsHandler) *GetProducts {
	return &GetProducts{Context: ctx, Handler: handler}
}

/*
	GetProducts swagger:route GET /products products getProducts

List products
*/
type GetProducts struct {
	Context *middleware.Context
	Handler GetProductsHandler
}

func (o *GetProducts) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetProductsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
