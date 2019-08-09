// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Message message
// swagger:model message
type Message struct {

	// chat author
	Name string `json:"name,omitempty"`

	// message text
	Text string `json:"text,omitempty"`

	// message type
	// Enum: [message joined left]
	Type string `json:"type,omitempty"`
}

// Validate validates this message
func (m *Message) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var messageTypeTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["message","joined","left"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		messageTypeTypePropEnum = append(messageTypeTypePropEnum, v)
	}
}

const (

	// MessageTypeMessage captures enum value "message"
	MessageTypeMessage string = "message"

	// MessageTypeJoined captures enum value "joined"
	MessageTypeJoined string = "joined"

	// MessageTypeLeft captures enum value "left"
	MessageTypeLeft string = "left"
)

// prop value enum
func (m *Message) validateTypeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, messageTypeTypePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Message) validateType(formats strfmt.Registry) error {

	if swag.IsZero(m.Type) { // not required
		return nil
	}

	// value enum
	if err := m.validateTypeEnum("type", "body", m.Type); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Message) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Message) UnmarshalBinary(b []byte) error {
	var res Message
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
