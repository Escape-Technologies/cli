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

// Enum549d6d987f9711d8c5b7a2472e0c9d65 the model 'Enum549d6d987f9711d8c5b7a2472e0c9d65'
type Enum549d6d987f9711d8c5b7a2472e0c9d65 string

// List of Enum_549d6d987f9711d8c5b7a2472e0c9d65
const (
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_GET Enum549d6d987f9711d8c5b7a2472e0c9d65 = "GET"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_POST Enum549d6d987f9711d8c5b7a2472e0c9d65 = "POST"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_PUT Enum549d6d987f9711d8c5b7a2472e0c9d65 = "PUT"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_DELETE Enum549d6d987f9711d8c5b7a2472e0c9d65 = "DELETE"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_HEAD Enum549d6d987f9711d8c5b7a2472e0c9d65 = "HEAD"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_PATCH Enum549d6d987f9711d8c5b7a2472e0c9d65 = "PATCH"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_OPTIONS Enum549d6d987f9711d8c5b7a2472e0c9d65 = "OPTIONS"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_TRACE Enum549d6d987f9711d8c5b7a2472e0c9d65 = "TRACE"
	ENUM549D6D987F9711D8C5B7A2472E0C9D65_CONNECT Enum549d6d987f9711d8c5b7a2472e0c9d65 = "CONNECT"
)

// All allowed values of Enum549d6d987f9711d8c5b7a2472e0c9d65 enum
var AllowedEnum549d6d987f9711d8c5b7a2472e0c9d65EnumValues = []Enum549d6d987f9711d8c5b7a2472e0c9d65{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"HEAD",
	"PATCH",
	"OPTIONS",
	"TRACE",
	"CONNECT",
}

func (v *Enum549d6d987f9711d8c5b7a2472e0c9d65) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := Enum549d6d987f9711d8c5b7a2472e0c9d65(value)
	for _, existing := range AllowedEnum549d6d987f9711d8c5b7a2472e0c9d65EnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid Enum549d6d987f9711d8c5b7a2472e0c9d65", value)
}

// NewEnum549d6d987f9711d8c5b7a2472e0c9d65FromValue returns a pointer to a valid Enum549d6d987f9711d8c5b7a2472e0c9d65
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnum549d6d987f9711d8c5b7a2472e0c9d65FromValue(v string) (*Enum549d6d987f9711d8c5b7a2472e0c9d65, error) {
	ev := Enum549d6d987f9711d8c5b7a2472e0c9d65(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for Enum549d6d987f9711d8c5b7a2472e0c9d65: valid values are %v", v, AllowedEnum549d6d987f9711d8c5b7a2472e0c9d65EnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v Enum549d6d987f9711d8c5b7a2472e0c9d65) IsValid() bool {
	for _, existing := range AllowedEnum549d6d987f9711d8c5b7a2472e0c9d65EnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_549d6d987f9711d8c5b7a2472e0c9d65 value
func (v Enum549d6d987f9711d8c5b7a2472e0c9d65) Ptr() *Enum549d6d987f9711d8c5b7a2472e0c9d65 {
	return &v
}

type NullableEnum549d6d987f9711d8c5b7a2472e0c9d65 struct {
	value *Enum549d6d987f9711d8c5b7a2472e0c9d65
	isSet bool
}

func (v NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) Get() *Enum549d6d987f9711d8c5b7a2472e0c9d65 {
	return v.value
}

func (v *NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) Set(val *Enum549d6d987f9711d8c5b7a2472e0c9d65) {
	v.value = val
	v.isSet = true
}

func (v NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) IsSet() bool {
	return v.isSet
}

func (v *NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnum549d6d987f9711d8c5b7a2472e0c9d65(val *Enum549d6d987f9711d8c5b7a2472e0c9d65) *NullableEnum549d6d987f9711d8c5b7a2472e0c9d65 {
	return &NullableEnum549d6d987f9711d8c5b7a2472e0c9d65{value: val, isSet: true}
}

func (v NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnum549d6d987f9711d8c5b7a2472e0c9d65) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

