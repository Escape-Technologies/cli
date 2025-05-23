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

// EnumWIZ the model 'EnumWIZ'
type EnumWIZ string

// List of Enum_WIZ
const (
	ENUMWIZ_WIZ EnumWIZ = "WIZ"
)

// All allowed values of EnumWIZ enum
var AllowedEnumWIZEnumValues = []EnumWIZ{
	"WIZ",
}

func (v *EnumWIZ) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumWIZ(value)
	for _, existing := range AllowedEnumWIZEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumWIZ", value)
}

// NewEnumWIZFromValue returns a pointer to a valid EnumWIZ
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumWIZFromValue(v string) (*EnumWIZ, error) {
	ev := EnumWIZ(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumWIZ: valid values are %v", v, AllowedEnumWIZEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumWIZ) IsValid() bool {
	for _, existing := range AllowedEnumWIZEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_WIZ value
func (v EnumWIZ) Ptr() *EnumWIZ {
	return &v
}

type NullableEnumWIZ struct {
	value *EnumWIZ
	isSet bool
}

func (v NullableEnumWIZ) Get() *EnumWIZ {
	return v.value
}

func (v *NullableEnumWIZ) Set(val *EnumWIZ) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumWIZ) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumWIZ) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumWIZ(val *EnumWIZ) *NullableEnumWIZ {
	return &NullableEnumWIZ{value: val, isSet: true}
}

func (v NullableEnumWIZ) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumWIZ) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

