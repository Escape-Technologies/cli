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

// FrontendConfigurationCustomRulesInnerTransformMutateInner - struct for FrontendConfigurationCustomRulesInnerTransformMutateInner
type FrontendConfigurationCustomRulesInnerTransformMutateInner struct {
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6
	FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfAsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfAsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6: v,
	}
}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7AsFrontendConfigurationCustomRulesInnerTransformMutateInner is a convenience function that returns FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 wrapped in FrontendConfigurationCustomRulesInnerTransformMutateInner
func FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7AsFrontendConfigurationCustomRulesInnerTransformMutateInner(v *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7) FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return FrontendConfigurationCustomRulesInnerTransformMutateInner{
		FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7: v,
	}
}


// Unmarshal JSON data into one of the pointers in the struct
func (dst *FrontendConfigurationCustomRulesInnerTransformMutateInner) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 = nil
	}

	// try to unmarshal data into FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7
	err = newStrictDecoder(data).Decode(&dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7)
	if err == nil {
		jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7, _ := json.Marshal(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7)
		if string(jsonFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7) == "{}" { // empty struct
			dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 = nil
		} else {
			if err = validator.Validate(dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7); err != nil {
				dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 = nil
			} else {
				match++
			}
		}
	} else {
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 = nil
		dst.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 = nil

		return fmt.Errorf("data matches more than one schema in oneOf(FrontendConfigurationCustomRulesInnerTransformMutateInner)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(FrontendConfigurationCustomRulesInnerTransformMutateInner)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src FrontendConfigurationCustomRulesInnerTransformMutateInner) MarshalJSON() ([]byte, error) {
	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6)
	}

	if src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 != nil {
		return json.Marshal(&src.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *FrontendConfigurationCustomRulesInnerTransformMutateInner) GetActualInstance() (interface{}) {
	if obj == nil {
		return nil
	}
	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 != nil {
		return obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7
	}

	// all schemas are nil
	return nil
}

// Get the actual instance value
func (obj FrontendConfigurationCustomRulesInnerTransformMutateInner) GetActualInstanceValue() (interface{}) {
	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf1
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf2
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf5
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6
	}

	if obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7 != nil {
		return *obj.FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf7
	}

	// all schemas are nil
	return nil
}

type NullableFrontendConfigurationCustomRulesInnerTransformMutateInner struct {
	value *FrontendConfigurationCustomRulesInnerTransformMutateInner
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) Get() *FrontendConfigurationCustomRulesInnerTransformMutateInner {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) Set(val *FrontendConfigurationCustomRulesInnerTransformMutateInner) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformMutateInner(val *FrontendConfigurationCustomRulesInnerTransformMutateInner) *NullableFrontendConfigurationCustomRulesInnerTransformMutateInner {
	return &NullableFrontendConfigurationCustomRulesInnerTransformMutateInner{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


