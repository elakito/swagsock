// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/elakito/swagsock/examples/greeter-client/models"
)

// NewGreetParams creates a new GreetParams object
// with the default values initialized.
func NewGreetParams() *GreetParams {
	var ()
	return &GreetParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGreetParamsWithTimeout creates a new GreetParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGreetParamsWithTimeout(timeout time.Duration) *GreetParams {
	var ()
	return &GreetParams{

		timeout: timeout,
	}
}

// NewGreetParamsWithContext creates a new GreetParams object
// with the default values initialized, and the ability to set a context for a request
func NewGreetParamsWithContext(ctx context.Context) *GreetParams {
	var ()
	return &GreetParams{

		Context: ctx,
	}
}

// NewGreetParamsWithHTTPClient creates a new GreetParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGreetParamsWithHTTPClient(client *http.Client) *GreetParams {
	var ()
	return &GreetParams{
		HTTPClient: client,
	}
}

/*GreetParams contains all the parameters to send to the API endpoint
for the greet operation typically these are written to a http.Request
*/
type GreetParams struct {

	/*Body
	  A greeting message

	*/
	Body *models.Body
	/*Name
	  greeting name

	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the greet params
func (o *GreetParams) WithTimeout(timeout time.Duration) *GreetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the greet params
func (o *GreetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the greet params
func (o *GreetParams) WithContext(ctx context.Context) *GreetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the greet params
func (o *GreetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the greet params
func (o *GreetParams) WithHTTPClient(client *http.Client) *GreetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the greet params
func (o *GreetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the greet params
func (o *GreetParams) WithBody(body *models.Body) *GreetParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the greet params
func (o *GreetParams) SetBody(body *models.Body) {
	o.Body = body
}

// WithName adds the name to the greet params
func (o *GreetParams) WithName(name string) *GreetParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the greet params
func (o *GreetParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *GreetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
