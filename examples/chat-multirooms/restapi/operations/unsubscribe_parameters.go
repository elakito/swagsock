// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"
)

// NewUnsubscribeParams creates a new UnsubscribeParams object
// no default values defined in spec.
func NewUnsubscribeParams() UnsubscribeParams {

	return UnsubscribeParams{}
}

// UnsubscribeParams contains all the bound params for the unsubscribe operation
// typically these are obtained from a http.Request
//
// swagger:parameters unsubscribe
type UnsubscribeParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*subscription id
	  Required: true
	  In: path
	*/
	Sid string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUnsubscribeParams() beforehand.
func (o *UnsubscribeParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rSid, rhkSid, _ := route.Params.GetOK("sid")
	if err := o.bindSid(rSid, rhkSid, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindSid binds and validates parameter Sid from path.
func (o *UnsubscribeParams) bindSid(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.Sid = raw

	return nil
}
