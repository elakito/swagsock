// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/elakito/swagsock/examples/greeter/models"
)

// UploadOKCode is the HTTP code returned for type UploadOK
const UploadOKCode int = 200

/*UploadOK Upload status

swagger:response uploadOK
*/
type UploadOK struct {

	/*
	  In: Body
	*/
	Payload *models.UploadStatus `json:"body,omitempty"`
}

// NewUploadOK creates UploadOK with default headers values
func NewUploadOK() *UploadOK {

	return &UploadOK{}
}

// WithPayload adds the payload to the upload o k response
func (o *UploadOK) WithPayload(payload *models.UploadStatus) *UploadOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the upload o k response
func (o *UploadOK) SetPayload(payload *models.UploadStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UploadOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UploadInternalServerErrorCode is the HTTP code returned for type UploadInternalServerError
const UploadInternalServerErrorCode int = 500

/*UploadInternalServerError Storage Full

swagger:response uploadInternalServerError
*/
type UploadInternalServerError struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewUploadInternalServerError creates UploadInternalServerError with default headers values
func NewUploadInternalServerError() *UploadInternalServerError {

	return &UploadInternalServerError{}
}

// WithPayload adds the payload to the upload internal server error response
func (o *UploadInternalServerError) WithPayload(payload string) *UploadInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the upload internal server error response
func (o *UploadInternalServerError) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UploadInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}
