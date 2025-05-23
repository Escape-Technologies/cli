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

// Enum83878678114937d099c9b534cb78beaf the model 'Enum83878678114937d099c9b534cb78beaf'
type Enum83878678114937d099c9b534cb78beaf string

// List of Enum_83878678114937d099c9b534cb78beaf
const (
	ENUM83878678114937D099C9B534CB78BEAF_HEADER Enum83878678114937d099c9b534cb78beaf = "header"
	ENUM83878678114937D099C9B534CB78BEAF_COOKIE Enum83878678114937d099c9b534cb78beaf = "cookie"
	ENUM83878678114937D099C9B534CB78BEAF_BODY Enum83878678114937d099c9b534cb78beaf = "body"
	ENUM83878678114937D099C9B534CB78BEAF_QUERY Enum83878678114937d099c9b534cb78beaf = "query"
)

// All allowed values of Enum83878678114937d099c9b534cb78beaf enum
var AllowedEnum83878678114937d099c9b534cb78beafEnumValues = []Enum83878678114937d099c9b534cb78beaf{
	"header",
	"cookie",
	"body",
	"query",
}

func (v *Enum83878678114937d099c9b534cb78beaf) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := Enum83878678114937d099c9b534cb78beaf(value)
	for _, existing := range AllowedEnum83878678114937d099c9b534cb78beafEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid Enum83878678114937d099c9b534cb78beaf", value)
}

// NewEnum83878678114937d099c9b534cb78beafFromValue returns a pointer to a valid Enum83878678114937d099c9b534cb78beaf
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnum83878678114937d099c9b534cb78beafFromValue(v string) (*Enum83878678114937d099c9b534cb78beaf, error) {
	ev := Enum83878678114937d099c9b534cb78beaf(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for Enum83878678114937d099c9b534cb78beaf: valid values are %v", v, AllowedEnum83878678114937d099c9b534cb78beafEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v Enum83878678114937d099c9b534cb78beaf) IsValid() bool {
	for _, existing := range AllowedEnum83878678114937d099c9b534cb78beafEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_83878678114937d099c9b534cb78beaf value
func (v Enum83878678114937d099c9b534cb78beaf) Ptr() *Enum83878678114937d099c9b534cb78beaf {
	return &v
}

type NullableEnum83878678114937d099c9b534cb78beaf struct {
	value *Enum83878678114937d099c9b534cb78beaf
	isSet bool
}

func (v NullableEnum83878678114937d099c9b534cb78beaf) Get() *Enum83878678114937d099c9b534cb78beaf {
	return v.value
}

func (v *NullableEnum83878678114937d099c9b534cb78beaf) Set(val *Enum83878678114937d099c9b534cb78beaf) {
	v.value = val
	v.isSet = true
}

func (v NullableEnum83878678114937d099c9b534cb78beaf) IsSet() bool {
	return v.isSet
}

func (v *NullableEnum83878678114937d099c9b534cb78beaf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnum83878678114937d099c9b534cb78beaf(val *Enum83878678114937d099c9b534cb78beaf) *NullableEnum83878678114937d099c9b534cb78beaf {
	return &NullableEnum83878678114937d099c9b534cb78beaf{value: val, isSet: true}
}

func (v NullableEnum83878678114937d099c9b534cb78beaf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnum83878678114937d099c9b534cb78beaf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

