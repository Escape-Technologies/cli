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

// EnumREQUESTISAUTHENTICATED the model 'EnumREQUESTISAUTHENTICATED'
type EnumREQUESTISAUTHENTICATED string

// List of Enum_REQUEST_IS_AUTHENTICATED
const (
	ENUMREQUESTISAUTHENTICATED_REQUEST_IS_AUTHENTICATED EnumREQUESTISAUTHENTICATED = "request.is_authenticated"
)

// All allowed values of EnumREQUESTISAUTHENTICATED enum
var AllowedEnumREQUESTISAUTHENTICATEDEnumValues = []EnumREQUESTISAUTHENTICATED{
	"request.is_authenticated",
}

func (v *EnumREQUESTISAUTHENTICATED) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumREQUESTISAUTHENTICATED(value)
	for _, existing := range AllowedEnumREQUESTISAUTHENTICATEDEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumREQUESTISAUTHENTICATED", value)
}

// NewEnumREQUESTISAUTHENTICATEDFromValue returns a pointer to a valid EnumREQUESTISAUTHENTICATED
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumREQUESTISAUTHENTICATEDFromValue(v string) (*EnumREQUESTISAUTHENTICATED, error) {
	ev := EnumREQUESTISAUTHENTICATED(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumREQUESTISAUTHENTICATED: valid values are %v", v, AllowedEnumREQUESTISAUTHENTICATEDEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumREQUESTISAUTHENTICATED) IsValid() bool {
	for _, existing := range AllowedEnumREQUESTISAUTHENTICATEDEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_REQUEST_IS_AUTHENTICATED value
func (v EnumREQUESTISAUTHENTICATED) Ptr() *EnumREQUESTISAUTHENTICATED {
	return &v
}

type NullableEnumREQUESTISAUTHENTICATED struct {
	value *EnumREQUESTISAUTHENTICATED
	isSet bool
}

func (v NullableEnumREQUESTISAUTHENTICATED) Get() *EnumREQUESTISAUTHENTICATED {
	return v.value
}

func (v *NullableEnumREQUESTISAUTHENTICATED) Set(val *EnumREQUESTISAUTHENTICATED) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumREQUESTISAUTHENTICATED) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumREQUESTISAUTHENTICATED) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumREQUESTISAUTHENTICATED(val *EnumREQUESTISAUTHENTICATED) *NullableEnumREQUESTISAUTHENTICATED {
	return &NullableEnumREQUESTISAUTHENTICATED{value: val, isSet: true}
}

func (v NullableEnumREQUESTISAUTHENTICATED) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumREQUESTISAUTHENTICATED) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

