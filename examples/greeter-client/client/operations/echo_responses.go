// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/elakito/swagsock/examples/greeter-client/models"
)

// EchoReader is a Reader for the Echo structure.
type EchoReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *EchoReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewEchoOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewEchoOK creates a EchoOK with default headers values
func NewEchoOK() *EchoOK {
	return &EchoOK{}
}

/*EchoOK handles this case with default header values.

Returns an echo message
*/
type EchoOK struct {
	Payload *models.EchoReply
}

func (o *EchoOK) Error() string {
	return fmt.Sprintf("[POST /v1/echo][%d] echoOK  %+v", 200, o.Payload)
}

func (o *EchoOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.EchoReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}