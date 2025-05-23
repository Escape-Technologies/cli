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

// EnumOAUTHUSERPASS the model 'EnumOAUTHUSERPASS'
type EnumOAUTHUSERPASS string

// List of Enum_OAUTH_USERPASS
const (
	ENUMOAUTHUSERPASS_OAUTH_USERPASS EnumOAUTHUSERPASS = "oauth_userpass"
)

// All allowed values of EnumOAUTHUSERPASS enum
var AllowedEnumOAUTHUSERPASSEnumValues = []EnumOAUTHUSERPASS{
	"oauth_userpass",
}

func (v *EnumOAUTHUSERPASS) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumOAUTHUSERPASS(value)
	for _, existing := range AllowedEnumOAUTHUSERPASSEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumOAUTHUSERPASS", value)
}

// NewEnumOAUTHUSERPASSFromValue returns a pointer to a valid EnumOAUTHUSERPASS
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumOAUTHUSERPASSFromValue(v string) (*EnumOAUTHUSERPASS, error) {
	ev := EnumOAUTHUSERPASS(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumOAUTHUSERPASS: valid values are %v", v, AllowedEnumOAUTHUSERPASSEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumOAUTHUSERPASS) IsValid() bool {
	for _, existing := range AllowedEnumOAUTHUSERPASSEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_OAUTH_USERPASS value
func (v EnumOAUTHUSERPASS) Ptr() *EnumOAUTHUSERPASS {
	return &v
}

type NullableEnumOAUTHUSERPASS struct {
	value *EnumOAUTHUSERPASS
	isSet bool
}

func (v NullableEnumOAUTHUSERPASS) Get() *EnumOAUTHUSERPASS {
	return v.value
}

func (v *NullableEnumOAUTHUSERPASS) Set(val *EnumOAUTHUSERPASS) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumOAUTHUSERPASS) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumOAUTHUSERPASS) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumOAUTHUSERPASS(val *EnumOAUTHUSERPASS) *NullableEnumOAUTHUSERPASS {
	return &NullableEnumOAUTHUSERPASS{value: val, isSet: true}
}

func (v NullableEnumOAUTHUSERPASS) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumOAUTHUSERPASS) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

