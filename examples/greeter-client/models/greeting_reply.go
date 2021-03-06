// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// GreetingReply greeting reply
// swagger:model greetingReply
type GreetingReply struct {

	// greeting person
	From string `json:"from,omitempty"`

	// greeted person
	Name string `json:"name,omitempty"`

	// greeting reply text
	Text string `json:"text,omitempty"`
}

// Validate validates this greeting reply
func (m *GreetingReply) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GreetingReply) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GreetingReply) UnmarshalBinary(b []byte) error {
	var res GreetingReply
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
