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

// checks if the FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21{}

// FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 struct for FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21
type FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 struct {
	Not interface{} `json:"not,omitempty"`
	If EnumNOT `json:"if"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21(if_ EnumNOT) *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21{}
	this.If = if_
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21WithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21WithDefaults() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21{}
	return &this
}

// GetNot returns the Not field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) GetNot() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Not
}

// GetNotOk returns a tuple with the Not field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) GetNotOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Not) {
		return nil, false
	}
	return &o.Not, true
}

// HasNot returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) HasNot() bool {
	if o != nil && !IsNil(o.Not) {
		return true
	}

	return false
}

// SetNot gets a reference to the given interface{} and assigns it to the Not field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) SetNot(v interface{}) {
	o.Not = v
}

// GetIf returns the If field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) GetIf() EnumNOT {
	if o == nil {
		var ret EnumNOT
		return ret
	}

	return o.If
}

// GetIfOk returns a tuple with the If field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) GetIfOk() (*EnumNOT, bool) {
	if o == nil {
		return nil, false
	}
	return &o.If, true
}

// SetIf sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) SetIf(v EnumNOT) {
	o.If = v
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if o.Not != nil {
		toSerialize["not"] = o.Not
	}
	toSerialize["if"] = o.If

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"if",
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

	varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 := _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21(varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "not")
		delete(additionalProperties, "if")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 struct {
	value *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) Get() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) Set(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21 {
	return &NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf21) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


