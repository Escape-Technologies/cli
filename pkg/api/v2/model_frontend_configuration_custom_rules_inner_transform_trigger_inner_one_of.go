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

// checks if the FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf{}

// FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf struct for FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf
type FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf struct {
	Is *Enum1ab0ceef9ae9ece93c01f2d976ec3990 `json:"is,omitempty"`
	IsNot *Enum1ab0ceef9ae9ece93c01f2d976ec3990 `json:"is_not,omitempty"`
	In []Enum1ab0ceef9ae9ece93c01f2d976ec3990 `json:"in,omitempty"`
	If EnumSCANTYPE `json:"if"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf(if_ EnumSCANTYPE) *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf{}
	this.If = if_
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOfWithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOfWithDefaults() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf{}
	return &this
}

// GetIs returns the Is field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIs() Enum1ab0ceef9ae9ece93c01f2d976ec3990 {
	if o == nil || IsNil(o.Is) {
		var ret Enum1ab0ceef9ae9ece93c01f2d976ec3990
		return ret
	}
	return *o.Is
}

// GetIsOk returns a tuple with the Is field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIsOk() (*Enum1ab0ceef9ae9ece93c01f2d976ec3990, bool) {
	if o == nil || IsNil(o.Is) {
		return nil, false
	}
	return o.Is, true
}

// HasIs returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) HasIs() bool {
	if o != nil && !IsNil(o.Is) {
		return true
	}

	return false
}

// SetIs gets a reference to the given Enum1ab0ceef9ae9ece93c01f2d976ec3990 and assigns it to the Is field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) SetIs(v Enum1ab0ceef9ae9ece93c01f2d976ec3990) {
	o.Is = &v
}

// GetIsNot returns the IsNot field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIsNot() Enum1ab0ceef9ae9ece93c01f2d976ec3990 {
	if o == nil || IsNil(o.IsNot) {
		var ret Enum1ab0ceef9ae9ece93c01f2d976ec3990
		return ret
	}
	return *o.IsNot
}

// GetIsNotOk returns a tuple with the IsNot field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIsNotOk() (*Enum1ab0ceef9ae9ece93c01f2d976ec3990, bool) {
	if o == nil || IsNil(o.IsNot) {
		return nil, false
	}
	return o.IsNot, true
}

// HasIsNot returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) HasIsNot() bool {
	if o != nil && !IsNil(o.IsNot) {
		return true
	}

	return false
}

// SetIsNot gets a reference to the given Enum1ab0ceef9ae9ece93c01f2d976ec3990 and assigns it to the IsNot field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) SetIsNot(v Enum1ab0ceef9ae9ece93c01f2d976ec3990) {
	o.IsNot = &v
}

// GetIn returns the In field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIn() []Enum1ab0ceef9ae9ece93c01f2d976ec3990 {
	if o == nil || IsNil(o.In) {
		var ret []Enum1ab0ceef9ae9ece93c01f2d976ec3990
		return ret
	}
	return o.In
}

// GetInOk returns a tuple with the In field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetInOk() ([]Enum1ab0ceef9ae9ece93c01f2d976ec3990, bool) {
	if o == nil || IsNil(o.In) {
		return nil, false
	}
	return o.In, true
}

// HasIn returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) HasIn() bool {
	if o != nil && !IsNil(o.In) {
		return true
	}

	return false
}

// SetIn gets a reference to the given []Enum1ab0ceef9ae9ece93c01f2d976ec3990 and assigns it to the In field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) SetIn(v []Enum1ab0ceef9ae9ece93c01f2d976ec3990) {
	o.In = v
}

// GetIf returns the If field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIf() EnumSCANTYPE {
	if o == nil {
		var ret EnumSCANTYPE
		return ret
	}

	return o.If
}

// GetIfOk returns a tuple with the If field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) GetIfOk() (*EnumSCANTYPE, bool) {
	if o == nil {
		return nil, false
	}
	return &o.If, true
}

// SetIf sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) SetIf(v EnumSCANTYPE) {
	o.If = v
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Is) {
		toSerialize["is"] = o.Is
	}
	if !IsNil(o.IsNot) {
		toSerialize["is_not"] = o.IsNot
	}
	if !IsNil(o.In) {
		toSerialize["in"] = o.In
	}
	toSerialize["if"] = o.If

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) UnmarshalJSON(data []byte) (err error) {
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

	varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf := _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf(varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "is")
		delete(additionalProperties, "is_not")
		delete(additionalProperties, "in")
		delete(additionalProperties, "if")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf struct {
	value *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) Get() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) Set(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf {
	return &NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


