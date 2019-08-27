// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// SubscribeHandlerFunc turns a function with the right signature into a subscribe handler
type SubscribeHandlerFunc func(SubscribeParams) middleware.Responder

// Handle executing the request and returning a response
func (fn SubscribeHandlerFunc) Handle(params SubscribeParams) middleware.Responder {
	return fn(params)
}

// SubscribeHandler interface for that can handle valid subscribe params
type SubscribeHandler interface {
	Handle(SubscribeParams) middleware.Responder
}

// NewSubscribe creates a new http.Handler for the subscribe operation
func NewSubscribe(ctx *middleware.Context, handler SubscribeHandler) *Subscribe {
	return &Subscribe{Context: ctx, Handler: handler}
}

/*Subscribe swagger:route GET /v1/subscribe/{name}/{room} subscribe

Subscribe to the chat room events

Subscribe to the chat room events

*/
type Subscribe struct {
	Context *middleware.Context
	Handler SubscribeHandler
}

func (o *Subscribe) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewSubscribeParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
