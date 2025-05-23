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

// EnumRESPONSEOBJECT the model 'EnumRESPONSEOBJECT'
type EnumRESPONSEOBJECT string

// List of Enum_RESPONSE_OBJECT
const (
	ENUMRESPONSEOBJECT_RESPONSE_OBJECT EnumRESPONSEOBJECT = "response.object"
)

// All allowed values of EnumRESPONSEOBJECT enum
var AllowedEnumRESPONSEOBJECTEnumValues = []EnumRESPONSEOBJECT{
	"response.object",
}

func (v *EnumRESPONSEOBJECT) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumRESPONSEOBJECT(value)
	for _, existing := range AllowedEnumRESPONSEOBJECTEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumRESPONSEOBJECT", value)
}

// NewEnumRESPONSEOBJECTFromValue returns a pointer to a valid EnumRESPONSEOBJECT
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumRESPONSEOBJECTFromValue(v string) (*EnumRESPONSEOBJECT, error) {
	ev := EnumRESPONSEOBJECT(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumRESPONSEOBJECT: valid values are %v", v, AllowedEnumRESPONSEOBJECTEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumRESPONSEOBJECT) IsValid() bool {
	for _, existing := range AllowedEnumRESPONSEOBJECTEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_RESPONSE_OBJECT value
func (v EnumRESPONSEOBJECT) Ptr() *EnumRESPONSEOBJECT {
	return &v
}

type NullableEnumRESPONSEOBJECT struct {
	value *EnumRESPONSEOBJECT
	isSet bool
}

func (v NullableEnumRESPONSEOBJECT) Get() *EnumRESPONSEOBJECT {
	return v.value
}

func (v *NullableEnumRESPONSEOBJECT) Set(val *EnumRESPONSEOBJECT) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumRESPONSEOBJECT) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumRESPONSEOBJECT) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumRESPONSEOBJECT(val *EnumRESPONSEOBJECT) *NullableEnumRESPONSEOBJECT {
	return &NullableEnumRESPONSEOBJECT{value: val, isSet: true}
}

func (v NullableEnumRESPONSEOBJECT) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumRESPONSEOBJECT) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

