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

// GetGreetStatusReader is a Reader for the GetGreetStatus structure.
type GetGreetStatusReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetGreetStatusReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetGreetStatusOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetGreetStatusOK creates a GetGreetStatusOK with default headers values
func NewGetGreetStatusOK() *GetGreetStatusOK {
	return &GetGreetStatusOK{}
}

/*GetGreetStatusOK handles this case with default header values.

Greeting status
*/
type GetGreetStatusOK struct {
	Payload *models.GreetingStatus
}

func (o *GetGreetStatusOK) Error() string {
	return fmt.Sprintf("[GET /v1/greet/{name}][%d] getGreetStatusOK  %+v", 200, o.Payload)
}

func (o *GetGreetStatusOK) GetPayload() *models.GreetingStatus {
	return o.Payload
}

func (o *GetGreetStatusOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GreetingStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
