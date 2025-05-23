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

// EnumA8e620669cc60d45b9f04eb14bdfef5d the model 'EnumA8e620669cc60d45b9f04eb14bdfef5d'
type EnumA8e620669cc60d45b9f04eb14bdfef5d string

// List of Enum_a8e620669cc60d45b9f04eb14bdfef5d
const (
	ENUMA8E620669CC60D45B9F04EB14BDFEF5D_STRING EnumA8e620669cc60d45b9f04eb14bdfef5d = "String"
	ENUMA8E620669CC60D45B9F04EB14BDFEF5D_INT EnumA8e620669cc60d45b9f04eb14bdfef5d = "Int"
	ENUMA8E620669CC60D45B9F04EB14BDFEF5D_FLOAT EnumA8e620669cc60d45b9f04eb14bdfef5d = "Float"
	ENUMA8E620669CC60D45B9F04EB14BDFEF5D_BOOLEAN EnumA8e620669cc60d45b9f04eb14bdfef5d = "Boolean"
)

// All allowed values of EnumA8e620669cc60d45b9f04eb14bdfef5d enum
var AllowedEnumA8e620669cc60d45b9f04eb14bdfef5dEnumValues = []EnumA8e620669cc60d45b9f04eb14bdfef5d{
	"String",
	"Int",
	"Float",
	"Boolean",
}

func (v *EnumA8e620669cc60d45b9f04eb14bdfef5d) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnumA8e620669cc60d45b9f04eb14bdfef5d(value)
	for _, existing := range AllowedEnumA8e620669cc60d45b9f04eb14bdfef5dEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnumA8e620669cc60d45b9f04eb14bdfef5d", value)
}

// NewEnumA8e620669cc60d45b9f04eb14bdfef5dFromValue returns a pointer to a valid EnumA8e620669cc60d45b9f04eb14bdfef5d
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnumA8e620669cc60d45b9f04eb14bdfef5dFromValue(v string) (*EnumA8e620669cc60d45b9f04eb14bdfef5d, error) {
	ev := EnumA8e620669cc60d45b9f04eb14bdfef5d(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnumA8e620669cc60d45b9f04eb14bdfef5d: valid values are %v", v, AllowedEnumA8e620669cc60d45b9f04eb14bdfef5dEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnumA8e620669cc60d45b9f04eb14bdfef5d) IsValid() bool {
	for _, existing := range AllowedEnumA8e620669cc60d45b9f04eb14bdfef5dEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_a8e620669cc60d45b9f04eb14bdfef5d value
func (v EnumA8e620669cc60d45b9f04eb14bdfef5d) Ptr() *EnumA8e620669cc60d45b9f04eb14bdfef5d {
	return &v
}

type NullableEnumA8e620669cc60d45b9f04eb14bdfef5d struct {
	value *EnumA8e620669cc60d45b9f04eb14bdfef5d
	isSet bool
}

func (v NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) Get() *EnumA8e620669cc60d45b9f04eb14bdfef5d {
	return v.value
}

func (v *NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) Set(val *EnumA8e620669cc60d45b9f04eb14bdfef5d) {
	v.value = val
	v.isSet = true
}

func (v NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) IsSet() bool {
	return v.isSet
}

func (v *NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnumA8e620669cc60d45b9f04eb14bdfef5d(val *EnumA8e620669cc60d45b9f04eb14bdfef5d) *NullableEnumA8e620669cc60d45b9f04eb14bdfef5d {
	return &NullableEnumA8e620669cc60d45b9f04eb14bdfef5d{value: val, isSet: true}
}

func (v NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnumA8e620669cc60d45b9f04eb14bdfef5d) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

