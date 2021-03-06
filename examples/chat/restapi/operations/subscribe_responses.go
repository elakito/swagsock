// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/elakito/swagsock/examples/chat/models"
)

// SubscribeOKCode is the HTTP code returned for type SubscribeOK
const SubscribeOKCode int = 200

/*SubscribeOK Subscription information

swagger:response subscribeOK
*/
type SubscribeOK struct {

	/*
	  In: Body
	*/
	Payload *models.Message `json:"body,omitempty"`
}

// NewSubscribeOK creates SubscribeOK with default headers values
func NewSubscribeOK() *SubscribeOK {

	return &SubscribeOK{}
}

// WithPayload adds the payload to the subscribe o k response
func (o *SubscribeOK) WithPayload(payload *models.Message) *SubscribeOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the subscribe o k response
func (o *SubscribeOK) SetPayload(payload *models.Message) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SubscribeOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
