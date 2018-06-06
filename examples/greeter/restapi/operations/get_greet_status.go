// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetGreetStatusHandlerFunc turns a function with the right signature into a get greet status handler
type GetGreetStatusHandlerFunc func(GetGreetStatusParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetGreetStatusHandlerFunc) Handle(params GetGreetStatusParams) middleware.Responder {
	return fn(params)
}

// GetGreetStatusHandler interface for that can handle valid get greet status params
type GetGreetStatusHandler interface {
	Handle(GetGreetStatusParams) middleware.Responder
}

// NewGetGreetStatus creates a new http.Handler for the get greet status operation
func NewGetGreetStatus(ctx *middleware.Context, handler GetGreetStatusHandler) *GetGreetStatus {
	return &GetGreetStatus{Context: ctx, Handler: handler}
}

/*GetGreetStatus swagger:route GET /v1/greet/{name} getGreetStatus

Show when the person was greeted last

Show when the person was last greeted

*/
type GetGreetStatus struct {
	Context *middleware.Context
	Handler GetGreetStatusHandler
}

func (o *GetGreetStatus) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetGreetStatusParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}