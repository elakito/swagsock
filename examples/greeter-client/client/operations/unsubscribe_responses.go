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

// UnsubscribeReader is a Reader for the Unsubscribe structure.
type UnsubscribeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UnsubscribeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUnsubscribeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewUnsubscribeOK creates a UnsubscribeOK with default headers values
func NewUnsubscribeOK() *UnsubscribeOK {
	return &UnsubscribeOK{}
}

/*UnsubscribeOK handles this case with default header values.

Returns the subscription summary
*/
type UnsubscribeOK struct {
	Payload *models.GreetingSummary
}

func (o *UnsubscribeOK) Error() string {
	return fmt.Sprintf("[DELETE /v1/unsubscribe/{sid}][%d] unsubscribeOK  %+v", 200, o.Payload)
}

func (o *UnsubscribeOK) GetPayload() *models.GreetingSummary {
	return o.Payload
}

func (o *UnsubscribeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GreetingSummary)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
