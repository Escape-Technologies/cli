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

// EnumEMAIL the model 'EnumEMAIL'
type EnumEMAIL string

// List of Enum_EMAIL
const (
	ENUMEMAIL_EMAIL EnumEMAIL = "EMAIL"
)

// All allowed values of EnumEMAIL enum
var AllowedEnumEMAILEnumValues = []EnumEMAIL{
	"EMAIL",
}

func (v *EnumEMAIL) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumEMAIL(value)
	for _, existing := range AllowedEnumEMAILEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumEMAIL", value)
}

// NewEnumEMAILFromValue returns a pointer to a valid EnumEMAIL
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumEMAILFromValue(v string) (*EnumEMAIL, error) {
	ev := EnumEMAIL(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumEMAIL: valid values are %v", v, AllowedEnumEMAILEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumEMAIL) IsValid() bool {
	for _, existing := range AllowedEnumEMAILEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_EMAIL value
func (v EnumEMAIL) Ptr() *EnumEMAIL {
	return &v
}

type NullableEnumEMAIL struct {
	value *EnumEMAIL
	isSet bool
}

func (v NullableEnumEMAIL) Get() *EnumEMAIL {
	return v.value
}

func (v *NullableEnumEMAIL) Set(val *EnumEMAIL) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumEMAIL) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumEMAIL) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumEMAIL(val *EnumEMAIL) *NullableEnumEMAIL {
	return &NullableEnumEMAIL{value: val, isSet: true}
}

func (v NullableEnumEMAIL) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumEMAIL) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

