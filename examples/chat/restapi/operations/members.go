// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// MembersHandlerFunc turns a function with the right signature into a members handler
type MembersHandlerFunc func(MembersParams) middleware.Responder

// Handle executing the request and returning a response
func (fn MembersHandlerFunc) Handle(params MembersParams) middleware.Responder {
	return fn(params)
}

// MembersHandler interface for that can handle valid members params
type MembersHandler interface {
	Handle(MembersParams) middleware.Responder
}

// NewMembers creates a new http.Handler for the members operation
func NewMembers(ctx *middleware.Context, handler MembersHandler) *Members {
	return &Members{Context: ctx, Handler: handler}
}

/*Members swagger:route GET /v1/members members

Get the list of chat members

Get the list of chat memebers

*/
type Members struct {
	Context *middleware.Context
	Handler MembersHandler
}

func (o *Members) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewMembersParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
