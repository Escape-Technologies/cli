/*
Escape Public API

This API enables you to operate [Escape](https://escape.tech/) programmatically.  All requests must be authenticated with a valid API key, provided in the `Authorization` header. For example: `Authorization: Key YOUR_API_KEY`.  You can find your API key in the [Escape dashboard](http://app.escape.tech/user/).

API version: 2.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package v2

import (
	"encoding/json"
)

// checks if the FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select{}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select struct for FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select
type FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select struct {
	Type *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type `json:"type,omitempty"`
	Name *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key `json:"name,omitempty"`
	Value *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key `json:"value,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select{}
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6SelectWithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6SelectWithDefaults() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select{}
	return &this
}

// GetType returns the Type field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetType() FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type {
	if o == nil || IsNil(o.Type) {
		var ret FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type
		return ret
	}
	return *o.Type
}

// GetTypeOk returns a tuple with the Type field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetTypeOk() (*FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type, bool) {
	if o == nil || IsNil(o.Type) {
		return nil, false
	}
	return o.Type, true
}

// HasType returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) HasType() bool {
	if o != nil && !IsNil(o.Type) {
		return true
	}

	return false
}

// SetType gets a reference to the given FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type and assigns it to the Type field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) SetType(v FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) {
	o.Type = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetName() FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key {
	if o == nil || IsNil(o.Name) {
		var ret FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetNameOk() (*FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key and assigns it to the Name field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) SetName(v FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key) {
	o.Name = &v
}

// GetValue returns the Value field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetValue() FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key {
	if o == nil || IsNil(o.Value) {
		var ret FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key
		return ret
	}
	return *o.Value
}

// GetValueOk returns a tuple with the Value field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) GetValueOk() (*FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key, bool) {
	if o == nil || IsNil(o.Value) {
		return nil, false
	}
	return o.Value, true
}

// HasValue returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) HasValue() bool {
	if o != nil && !IsNil(o.Value) {
		return true
	}

	return false
}

// SetValue gets a reference to the given FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key and assigns it to the Value field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) SetValue(v FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf10Key) {
	o.Value = &v
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Type) {
		toSerialize["type"] = o.Type
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !IsNil(o.Value) {
		toSerialize["value"] = o.Value
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) UnmarshalJSON(data []byte) (err error) {
	varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select := _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select(varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "type")
		delete(additionalProperties, "name")
		delete(additionalProperties, "value")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select struct {
	value *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) Get() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) Set(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select {
	return &NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf6Select) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


