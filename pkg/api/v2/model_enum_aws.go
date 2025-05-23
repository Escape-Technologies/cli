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

// EnumAWS the model 'EnumAWS'
type EnumAWS string

// List of Enum_AWS
const (
	ENUMAWS_AWS EnumAWS = "AWS"
)

// All allowed values of EnumAWS enum
var AllowedEnumAWSEnumValues = []EnumAWS{
	"AWS",
}

func (v *EnumAWS) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumAWS(value)
	for _, existing := range AllowedEnumAWSEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumAWS", value)
}

// NewEnumAWSFromValue returns a pointer to a valid EnumAWS
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumAWSFromValue(v string) (*EnumAWS, error) {
	ev := EnumAWS(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumAWS: valid values are %v", v, AllowedEnumAWSEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumAWS) IsValid() bool {
	for _, existing := range AllowedEnumAWSEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_AWS value
func (v EnumAWS) Ptr() *EnumAWS {
	return &v
}

type NullableEnumAWS struct {
	value *EnumAWS
	isSet bool
}

func (v NullableEnumAWS) Get() *EnumAWS {
	return v.value
}

func (v *NullableEnumAWS) Set(val *EnumAWS) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumAWS) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumAWS) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumAWS(val *EnumAWS) *NullableEnumAWS {
	return &NullableEnumAWS{value: val, isSet: true}
}

func (v NullableEnumAWS) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumAWS) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

