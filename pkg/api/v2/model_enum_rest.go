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

// EnumREST the model 'EnumREST'
type EnumREST string

// List of Enum_REST
const (
	ENUMREST_REST EnumREST = "rest"
)

// All allowed values of EnumREST enum
var AllowedEnumRESTEnumValues = []EnumREST{
	"rest",
}

func (v *EnumREST) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumREST(value)
	for _, existing := range AllowedEnumRESTEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumREST", value)
}

// NewEnumRESTFromValue returns a pointer to a valid EnumREST
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumRESTFromValue(v string) (*EnumREST, error) {
	ev := EnumREST(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumREST: valid values are %v", v, AllowedEnumRESTEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumREST) IsValid() bool {
	for _, existing := range AllowedEnumRESTEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_REST value
func (v EnumREST) Ptr() *EnumREST {
	return &v
}

type NullableEnumREST struct {
	value *EnumREST
	isSet bool
}

func (v NullableEnumREST) Get() *EnumREST {
	return v.value
}

func (v *NullableEnumREST) Set(val *EnumREST) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumREST) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumREST) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumREST(val *EnumREST) *NullableEnumREST {
	return &NullableEnumREST{value: val, isSet: true}
}

func (v NullableEnumREST) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumREST) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

