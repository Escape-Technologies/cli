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

// EnumAZUREDEVOPS the model 'EnumAZUREDEVOPS'
type EnumAZUREDEVOPS string

// List of Enum_AZURE_DEVOPS
const (
	ENUMAZUREDEVOPS_AZURE_DEVOPS EnumAZUREDEVOPS = "AZURE_DEVOPS"
)

// All allowed values of EnumAZUREDEVOPS enum
var AllowedEnumAZUREDEVOPSEnumValues = []EnumAZUREDEVOPS{
	"AZURE_DEVOPS",
}

func (v *EnumAZUREDEVOPS) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumAZUREDEVOPS(value)
	for _, existing := range AllowedEnumAZUREDEVOPSEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumAZUREDEVOPS", value)
}

// NewEnumAZUREDEVOPSFromValue returns a pointer to a valid EnumAZUREDEVOPS
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumAZUREDEVOPSFromValue(v string) (*EnumAZUREDEVOPS, error) {
	ev := EnumAZUREDEVOPS(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumAZUREDEVOPS: valid values are %v", v, AllowedEnumAZUREDEVOPSEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumAZUREDEVOPS) IsValid() bool {
	for _, existing := range AllowedEnumAZUREDEVOPSEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_AZURE_DEVOPS value
func (v EnumAZUREDEVOPS) Ptr() *EnumAZUREDEVOPS {
	return &v
}

type NullableEnumAZUREDEVOPS struct {
	value *EnumAZUREDEVOPS
	isSet bool
}

func (v NullableEnumAZUREDEVOPS) Get() *EnumAZUREDEVOPS {
	return v.value
}

func (v *NullableEnumAZUREDEVOPS) Set(val *EnumAZUREDEVOPS) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumAZUREDEVOPS) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumAZUREDEVOPS) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumAZUREDEVOPS(val *EnumAZUREDEVOPS) *NullableEnumAZUREDEVOPS {
	return &NullableEnumAZUREDEVOPS{value: val, isSet: true}
}

func (v NullableEnumAZUREDEVOPS) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumAZUREDEVOPS) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

