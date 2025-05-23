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

// EnumTEAMSWEBHOOK the model 'EnumTEAMSWEBHOOK'
type EnumTEAMSWEBHOOK string

// List of Enum_TEAMS_WEBHOOK
const (
	ENUMTEAMSWEBHOOK_TEAMS_WEBHOOK EnumTEAMSWEBHOOK = "TEAMS_WEBHOOK"
)

// All allowed values of EnumTEAMSWEBHOOK enum
var AllowedEnumTEAMSWEBHOOKEnumValues = []EnumTEAMSWEBHOOK{
	"TEAMS_WEBHOOK",
}

func (v *EnumTEAMSWEBHOOK) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumTEAMSWEBHOOK(value)
	for _, existing := range AllowedEnumTEAMSWEBHOOKEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumTEAMSWEBHOOK", value)
}

// NewEnumTEAMSWEBHOOKFromValue returns a pointer to a valid EnumTEAMSWEBHOOK
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumTEAMSWEBHOOKFromValue(v string) (*EnumTEAMSWEBHOOK, error) {
	ev := EnumTEAMSWEBHOOK(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumTEAMSWEBHOOK: valid values are %v", v, AllowedEnumTEAMSWEBHOOKEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumTEAMSWEBHOOK) IsValid() bool {
	for _, existing := range AllowedEnumTEAMSWEBHOOKEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_TEAMS_WEBHOOK value
func (v EnumTEAMSWEBHOOK) Ptr() *EnumTEAMSWEBHOOK {
	return &v
}

type NullableEnumTEAMSWEBHOOK struct {
	value *EnumTEAMSWEBHOOK
	isSet bool
}

func (v NullableEnumTEAMSWEBHOOK) Get() *EnumTEAMSWEBHOOK {
	return v.value
}

func (v *NullableEnumTEAMSWEBHOOK) Set(val *EnumTEAMSWEBHOOK) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumTEAMSWEBHOOK) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumTEAMSWEBHOOK) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumTEAMSWEBHOOK(val *EnumTEAMSWEBHOOK) *NullableEnumTEAMSWEBHOOK {
	return &NullableEnumTEAMSWEBHOOK{value: val, isSet: true}
}

func (v NullableEnumTEAMSWEBHOOK) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumTEAMSWEBHOOK) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

