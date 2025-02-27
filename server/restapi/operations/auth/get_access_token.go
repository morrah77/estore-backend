// Code generated by go-swagger; DO NOT EDIT.

package auth

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetAccessTokenHandlerFunc turns a function with the right signature into a get access token handler
type GetAccessTokenHandlerFunc func(GetAccessTokenParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAccessTokenHandlerFunc) Handle(params GetAccessTokenParams) middleware.Responder {
	return fn(params)
}

// GetAccessTokenHandler interface for that can handle valid get access token params
type GetAccessTokenHandler interface {
	Handle(GetAccessTokenParams) middleware.Responder
}

// NewGetAccessToken creates a new http.Handler for the get access token operation
func NewGetAccessToken(ctx *middleware.Context, handler GetAccessTokenHandler) *GetAccessToken {
	return &GetAccessToken{Context: ctx, Handler: handler}
}

/*
	GetAccessToken swagger:route GET /auth/cb auth getAccessToken

Obtain access token
*/
type GetAccessToken struct {
	Context *middleware.Context
	Handler GetAccessTokenHandler
}

func (o *GetAccessToken) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetAccessTokenParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
