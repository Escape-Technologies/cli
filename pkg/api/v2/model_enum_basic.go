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

// EnumBASIC the model 'EnumBASIC'
type EnumBASIC string

// List of Enum_BASIC
const (
	ENUMBASIC_BASIC EnumBASIC = "basic"
)

// All allowed values of EnumBASIC enum
var AllowedEnumBASICEnumValues = []EnumBASIC{
	"basic",
}

func (v *EnumBASIC) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumBASIC(value)
	for _, existing := range AllowedEnumBASICEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumBASIC", value)
}

// NewEnumBASICFromValue returns a pointer to a valid EnumBASIC
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumBASICFromValue(v string) (*EnumBASIC, error) {
	ev := EnumBASIC(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumBASIC: valid values are %v", v, AllowedEnumBASICEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumBASIC) IsValid() bool {
	for _, existing := range AllowedEnumBASICEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_BASIC value
func (v EnumBASIC) Ptr() *EnumBASIC {
	return &v
}

type NullableEnumBASIC struct {
	value *EnumBASIC
	isSet bool
}

func (v NullableEnumBASIC) Get() *EnumBASIC {
	return v.value
}

func (v *NullableEnumBASIC) Set(val *EnumBASIC) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumBASIC) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumBASIC) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumBASIC(val *EnumBASIC) *NullableEnumBASIC {
	return &NullableEnumBASIC{value: val, isSet: true}
}

func (v NullableEnumBASIC) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumBASIC) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

