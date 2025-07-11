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
	"gopkg.in/validator.v2"
)

// FrontendConfigurationAuthenticationProceduresInnerOperationsInner - struct for FrontendConfigurationAuthenticationProceduresInnerOperationsInner
type FrontendConfigurationAuthenticationProceduresInnerOperationsInner struct {
	FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf
	FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1
	FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2
	FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOfAsFrontendConfigurationAuthenticationProceduresInnerOperationsInner is a convenience function that returns FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf wrapped in FrontendConfigurationAuthenticationProceduresInnerOperationsInner
func FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOfAsFrontendConfigurationAuthenticationProceduresInnerOperationsInner(v *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf) FrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return FrontendConfigurationAuthenticationProceduresInnerOperationsInner{
		FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf: v,
	}
}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner is a convenience function that returns FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 wrapped in FrontendConfigurationAuthenticationProceduresInnerOperationsInner
func FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner(v *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1) FrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return FrontendConfigurationAuthenticationProceduresInnerOperationsInner{
		FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1: v,
	}
}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner is a convenience function that returns FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 wrapped in FrontendConfigurationAuthenticationProceduresInnerOperationsInner
func FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner(v *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2) FrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return FrontendConfigurationAuthenticationProceduresInnerOperationsInner{
		FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2: v,
	}
}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner is a convenience function that returns FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 wrapped in FrontendConfigurationAuthenticationProceduresInnerOperationsInner
func FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3AsFrontendConfigurationAuthenticationProceduresInnerOperationsInner(v *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) FrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return FrontendConfigurationAuthenticationProceduresInnerOperationsInner{
		FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3: v,
	}
}


// Unmarshal JSON data into one of the pointers in the struct
func (dst *FrontendConfigurationAuthenticationProceduresInnerOperationsInner) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf)
	if err == nil {
		jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf, _ := json.Marshal(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf)
		if string(jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf) == "{}" { // empty struct
			dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf); err != nil {
				dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf = nil
	}

	// try to unmarshal data into FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1)
	if err == nil {
		jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1, _ := json.Marshal(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1)
		if string(jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1) == "{}" { // empty struct
			dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1); err != nil {
				dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 = nil
	}

	// try to unmarshal data into FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2)
	if err == nil {
		jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2, _ := json.Marshal(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2)
		if string(jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2) == "{}" { // empty struct
			dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2); err != nil {
				dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 = nil
	}

	// try to unmarshal data into FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3)
	if err == nil {
		jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3, _ := json.Marshal(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3)
		if string(jsonFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) == "{}" { // empty struct
			dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3); err != nil {
				dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf = nil
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 = nil
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 = nil
		dst.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 = nil

		return fmt.Errorf("data matches more than one schema in oneOf(FrontendConfigurationAuthenticationProceduresInnerOperationsInner)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(FrontendConfigurationAuthenticationProceduresInnerOperationsInner)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src FrontendConfigurationAuthenticationProceduresInnerOperationsInner) MarshalJSON() ([]byte, error) {
	if src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf != nil {
		return json.Marshal(&src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf)
	}

	if src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 != nil {
		return json.Marshal(&src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1)
	}

	if src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 != nil {
		return json.Marshal(&src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2)
	}

	if src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 != nil {
		return json.Marshal(&src.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *FrontendConfigurationAuthenticationProceduresInnerOperationsInner) GetActualInstance() (interface{}) {
	if obj == nil {
		return nil
	}
	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf != nil {
		return obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 != nil {
		return obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 != nil {
		return obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 != nil {
		return obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
	}

	// all schemas are nil
	return nil
}

// Get the actual instance value
func (obj FrontendConfigurationAuthenticationProceduresInnerOperationsInner) GetActualInstanceValue() (interface{}) {
	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf != nil {
		return *obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1 != nil {
		return *obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2 != nil {
		return *obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2
	}

	if obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 != nil {
		return *obj.FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
	}

	// all schemas are nil
	return nil
}

type NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner struct {
	value *FrontendConfigurationAuthenticationProceduresInnerOperationsInner
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) Get() *FrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) Set(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInner) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInner) *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner {
	return &NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


