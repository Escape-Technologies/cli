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

// EnumF33adad4c808d4d9ab51ae2bf931668b the model 'EnumF33adad4c808d4d9ab51ae2bf931668b'
type EnumF33adad4c808d4d9ab51ae2bf931668b string

// List of Enum_f33adad4c808d4d9ab51ae2bf931668b
const (
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_6 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-6"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_21 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-21"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_22 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-22"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_23 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-23"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_28 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-28"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_29 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-29"
	ENUMF33ADAD4C808D4D9AB51AE2BF931668B_ARTICLE_33 EnumF33adad4c808d4d9ab51ae2bf931668b = "Article-33"
)

// All allowed values of EnumF33adad4c808d4d9ab51ae2bf931668b enum
var AllowedEnumF33adad4c808d4d9ab51ae2bf931668bEnumValues = []EnumF33adad4c808d4d9ab51ae2bf931668b{
	"Article-6",
	"Article-21",
	"Article-22",
	"Article-23",
	"Article-28",
	"Article-29",
	"Article-33",
}

func (v *EnumF33adad4c808d4d9ab51ae2bf931668b) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumF33adad4c808d4d9ab51ae2bf931668b(value)
	for _, existing := range AllowedEnumF33adad4c808d4d9ab51ae2bf931668bEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumF33adad4c808d4d9ab51ae2bf931668b", value)
}

// NewEnumF33adad4c808d4d9ab51ae2bf931668bFromValue returns a pointer to a valid EnumF33adad4c808d4d9ab51ae2bf931668b
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumF33adad4c808d4d9ab51ae2bf931668bFromValue(v string) (*EnumF33adad4c808d4d9ab51ae2bf931668b, error) {
	ev := EnumF33adad4c808d4d9ab51ae2bf931668b(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumF33adad4c808d4d9ab51ae2bf931668b: valid values are %v", v, AllowedEnumF33adad4c808d4d9ab51ae2bf931668bEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumF33adad4c808d4d9ab51ae2bf931668b) IsValid() bool {
	for _, existing := range AllowedEnumF33adad4c808d4d9ab51ae2bf931668bEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_f33adad4c808d4d9ab51ae2bf931668b value
func (v EnumF33adad4c808d4d9ab51ae2bf931668b) Ptr() *EnumF33adad4c808d4d9ab51ae2bf931668b {
	return &v
}

type NullableEnumF33adad4c808d4d9ab51ae2bf931668b struct {
	value *EnumF33adad4c808d4d9ab51ae2bf931668b
	isSet bool
}

func (v NullableEnumF33adad4c808d4d9ab51ae2bf931668b) Get() *EnumF33adad4c808d4d9ab51ae2bf931668b {
	return v.value
}

func (v *NullableEnumF33adad4c808d4d9ab51ae2bf931668b) Set(val *EnumF33adad4c808d4d9ab51ae2bf931668b) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumF33adad4c808d4d9ab51ae2bf931668b) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumF33adad4c808d4d9ab51ae2bf931668b) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumF33adad4c808d4d9ab51ae2bf931668b(val *EnumF33adad4c808d4d9ab51ae2bf931668b) *NullableEnumF33adad4c808d4d9ab51ae2bf931668b {
	return &NullableEnumF33adad4c808d4d9ab51ae2bf931668b{value: val, isSet: true}
}

func (v NullableEnumF33adad4c808d4d9ab51ae2bf931668b) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumF33adad4c808d4d9ab51ae2bf931668b) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

