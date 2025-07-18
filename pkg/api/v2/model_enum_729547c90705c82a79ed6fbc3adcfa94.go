/*
Escape Public API

This API enables you to operate [Escape](https://escape.tech/) programmatically.  All requests must be authenticated with a valid API key, provided in the `Authorization` header. For example: `Authorization: Key YOUR_API_KEY`.  You can find your API key in the [Escape dashboard](http://app.escape.tech/user/).

API version: 2.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package v2

import (
	"encoding/json"
	"fmt"
)

// Enum729547c90705c82a79ed6fbc3adcfa94 the model 'Enum729547c90705c82a79ed6fbc3adcfa94'
type Enum729547c90705c82a79ed6fbc3adcfa94 string

// List of Enum_729547c90705c82a79ed6fbc3adcfa94
const (
	ENUM729547C90705C82A79ED6FBC3ADCFA94_ACCESS_CONTROL Enum729547c90705c82a79ed6fbc3adcfa94 = "ACCESS_CONTROL"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_CONFIGURATION Enum729547c90705c82a79ed6fbc3adcfa94 = "CONFIGURATION"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_INFORMATION_DISCLOSURE Enum729547c90705c82a79ed6fbc3adcfa94 = "INFORMATION_DISCLOSURE"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_INJECTION Enum729547c90705c82a79ed6fbc3adcfa94 = "INJECTION"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_PROTOCOL Enum729547c90705c82a79ed6fbc3adcfa94 = "PROTOCOL"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_REQUEST_FORGERY Enum729547c90705c82a79ed6fbc3adcfa94 = "REQUEST_FORGERY"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_RESOURCE_LIMITATION Enum729547c90705c82a79ed6fbc3adcfa94 = "RESOURCE_LIMITATION"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_SENSITIVE_DATA Enum729547c90705c82a79ed6fbc3adcfa94 = "SENSITIVE_DATA"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_SCHEMA Enum729547c90705c82a79ed6fbc3adcfa94 = "SCHEMA"
	ENUM729547C90705C82A79ED6FBC3ADCFA94_CUSTOM Enum729547c90705c82a79ed6fbc3adcfa94 = "CUSTOM"
)

// All allowed values of Enum729547c90705c82a79ed6fbc3adcfa94 enum
var AllowedEnum729547c90705c82a79ed6fbc3adcfa94EnumValues = []Enum729547c90705c82a79ed6fbc3adcfa94{
	"ACCESS_CONTROL",
	"CONFIGURATION",
	"INFORMATION_DISCLOSURE",
	"INJECTION",
	"PROTOCOL",
	"REQUEST_FORGERY",
	"RESOURCE_LIMITATION",
	"SENSITIVE_DATA",
	"SCHEMA",
	"CUSTOM",
}

func (v *Enum729547c90705c82a79ed6fbc3adcfa94) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := Enum729547c90705c82a79ed6fbc3adcfa94(value)
	for _, existing := range AllowedEnum729547c90705c82a79ed6fbc3adcfa94EnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid Enum729547c90705c82a79ed6fbc3adcfa94", value)
}

// NewEnum729547c90705c82a79ed6fbc3adcfa94FromValue returns a pointer to a valid Enum729547c90705c82a79ed6fbc3adcfa94
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnum729547c90705c82a79ed6fbc3adcfa94FromValue(v string) (*Enum729547c90705c82a79ed6fbc3adcfa94, error) {
	ev := Enum729547c90705c82a79ed6fbc3adcfa94(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for Enum729547c90705c82a79ed6fbc3adcfa94: valid values are %v", v, AllowedEnum729547c90705c82a79ed6fbc3adcfa94EnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v Enum729547c90705c82a79ed6fbc3adcfa94) IsValid() bool {
	for _, existing := range AllowedEnum729547c90705c82a79ed6fbc3adcfa94EnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_729547c90705c82a79ed6fbc3adcfa94 value
func (v Enum729547c90705c82a79ed6fbc3adcfa94) Ptr() *Enum729547c90705c82a79ed6fbc3adcfa94 {
	return &v
}

type NullableEnum729547c90705c82a79ed6fbc3adcfa94 struct {
	value *Enum729547c90705c82a79ed6fbc3adcfa94
	isSet bool
}

func (v NullableEnum729547c90705c82a79ed6fbc3adcfa94) Get() *Enum729547c90705c82a79ed6fbc3adcfa94 {
	return v.value
}

func (v *NullableEnum729547c90705c82a79ed6fbc3adcfa94) Set(val *Enum729547c90705c82a79ed6fbc3adcfa94) {
	v.value = val
	v.isSet = true
}

func (v NullableEnum729547c90705c82a79ed6fbc3adcfa94) IsSet() bool {
	return v.isSet
}

func (v *NullableEnum729547c90705c82a79ed6fbc3adcfa94) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnum729547c90705c82a79ed6fbc3adcfa94(val *Enum729547c90705c82a79ed6fbc3adcfa94) *NullableEnum729547c90705c82a79ed6fbc3adcfa94 {
	return &NullableEnum729547c90705c82a79ed6fbc3adcfa94{value: val, isSet: true}
}

func (v NullableEnum729547c90705c82a79ed6fbc3adcfa94) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnum729547c90705c82a79ed6fbc3adcfa94) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

