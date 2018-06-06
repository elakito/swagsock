// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/elakito/swagsock/examples/greeter/models"
)

// GetGreetStatusOKCode is the HTTP code returned for type GetGreetStatusOK
const GetGreetStatusOKCode int = 200

/*GetGreetStatusOK Greeting status

swagger:response getGreetStatusOK
*/
type GetGreetStatusOK struct {

	/*
	  In: Body
	*/
	Payload *models.GreetingStatus `json:"body,omitempty"`
}

// NewGetGreetStatusOK creates GetGreetStatusOK with default headers values
func NewGetGreetStatusOK() *GetGreetStatusOK {
	return &GetGreetStatusOK{}
}

// WithPayload adds the payload to the get greet status o k response
func (o *GetGreetStatusOK) WithPayload(payload *models.GreetingStatus) *GetGreetStatusOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get greet status o k response
func (o *GetGreetStatusOK) SetPayload(payload *models.GreetingStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetGreetStatusOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}