// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// GreetingSummary greeting summary
// swagger:model greetingSummary
type GreetingSummary struct {

	// greeted persons
	Greeted []string `json:"greeted"`

	// total greeting count
	Total int32 `json:"total,omitempty"`
}

// Validate validates this greeting summary
func (m *GreetingSummary) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GreetingSummary) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GreetingSummary) UnmarshalBinary(b []byte) error {
	var res GreetingSummary
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
