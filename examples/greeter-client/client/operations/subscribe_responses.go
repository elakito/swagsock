// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/elakito/swagsock/examples/greeter-client/models"
)

// SubscribeReader is a Reader for the Subscribe structure.
type SubscribeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SubscribeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewSubscribeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewSubscribeOK creates a SubscribeOK with default headers values
func NewSubscribeOK() *SubscribeOK {
	return &SubscribeOK{}
}

/*SubscribeOK handles this case with default header values.

Returns an empty or a sequence of greeting events
*/
type SubscribeOK struct {
	Payload *models.GreetingReply
}

func (o *SubscribeOK) Error() string {
	return fmt.Sprintf("[GET /v1/subscribe/{name}][%d] subscribeOK  %+v", 200, o.Payload)
}

func (o *SubscribeOK) GetPayload() *models.GreetingReply {
	return o.Payload
}

func (o *SubscribeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GreetingReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
