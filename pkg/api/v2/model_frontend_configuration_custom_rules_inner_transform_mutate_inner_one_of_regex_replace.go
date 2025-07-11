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

// checks if the FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace{}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace struct for FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace
type FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace struct {
	Pattern string `json:"pattern"`
	Replacement string `json:"replacement"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace(pattern string, replacement string) *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace{}
	this.Pattern = pattern
	this.Replacement = replacement
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplaceWithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplaceWithDefaults() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace{}
	return &this
}

// GetPattern returns the Pattern field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) GetPattern() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Pattern
}

// GetPatternOk returns a tuple with the Pattern field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) GetPatternOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Pattern, true
}

// SetPattern sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) SetPattern(v string) {
	o.Pattern = v
}

// GetReplacement returns the Replacement field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) GetReplacement() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Replacement
}

// GetReplacementOk returns a tuple with the Replacement field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) GetReplacementOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Replacement, true
}

// SetReplacement sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) SetReplacement(v string) {
	o.Replacement = v
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["pattern"] = o.Pattern
	toSerialize["replacement"] = o.Replacement

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"pattern",
		"replacement",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace := _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace(varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "pattern")
		delete(additionalProperties, "replacement")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace struct {
	value *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) Get() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) Set(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace {
	return &NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


