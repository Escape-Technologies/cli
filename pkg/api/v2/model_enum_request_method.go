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

// EnumREQUESTMETHOD the model 'EnumREQUESTMETHOD'
type EnumREQUESTMETHOD string

// List of Enum_REQUEST_METHOD
const (
	ENUMREQUESTMETHOD_REQUEST_METHOD EnumREQUESTMETHOD = "request.method"
)

// All allowed values of EnumREQUESTMETHOD enum
var AllowedEnumREQUESTMETHODEnumValues = []EnumREQUESTMETHOD{
	"request.method",
}

func (v *EnumREQUESTMETHOD) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumREQUESTMETHOD(value)
	for _, existing := range AllowedEnumREQUESTMETHODEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumREQUESTMETHOD", value)
}

// NewEnumREQUESTMETHODFromValue returns a pointer to a valid EnumREQUESTMETHOD
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumREQUESTMETHODFromValue(v string) (*EnumREQUESTMETHOD, error) {
	ev := EnumREQUESTMETHOD(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumREQUESTMETHOD: valid values are %v", v, AllowedEnumREQUESTMETHODEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumREQUESTMETHOD) IsValid() bool {
	for _, existing := range AllowedEnumREQUESTMETHODEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_REQUEST_METHOD value
func (v EnumREQUESTMETHOD) Ptr() *EnumREQUESTMETHOD {
	return &v
}

type NullableEnumREQUESTMETHOD struct {
	value *EnumREQUESTMETHOD
	isSet bool
}

func (v NullableEnumREQUESTMETHOD) Get() *EnumREQUESTMETHOD {
	return v.value
}

func (v *NullableEnumREQUESTMETHOD) Set(val *EnumREQUESTMETHOD) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumREQUESTMETHOD) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumREQUESTMETHOD) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumREQUESTMETHOD(val *EnumREQUESTMETHOD) *NullableEnumREQUESTMETHOD {
	return &NullableEnumREQUESTMETHOD{value: val, isSet: true}
}

func (v NullableEnumREQUESTMETHOD) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumREQUESTMETHOD) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

