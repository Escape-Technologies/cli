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

// checks if the FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4{}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 struct for FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
type FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 struct {
	Key EnumREQUESTBODYJSON `json:"key"`
	Jq *string `json:"jq,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4(key EnumREQUESTBODYJSON) *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4{}
	this.Key = key
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4WithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4WithDefaults() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4{}
	return &this
}

// GetKey returns the Key field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) GetKey() EnumREQUESTBODYJSON {
	if o == nil {
		var ret EnumREQUESTBODYJSON
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) GetKeyOk() (*EnumREQUESTBODYJSON, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) SetKey(v EnumREQUESTBODYJSON) {
	o.Key = v
}

// GetJq returns the Jq field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) GetJq() string {
	if o == nil || IsNil(o.Jq) {
		var ret string
		return ret
	}
	return *o.Jq
}

// GetJqOk returns a tuple with the Jq field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) GetJqOk() (*string, bool) {
	if o == nil || IsNil(o.Jq) {
		return nil, false
	}
	return o.Jq, true
}

// HasJq returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) HasJq() bool {
	if o != nil && !IsNil(o.Jq) {
		return true
	}

	return false
}

// SetJq gets a reference to the given string and assigns it to the Jq field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) SetJq(v string) {
	o.Jq = &v
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["key"] = o.Key
	if !IsNil(o.Jq) {
		toSerialize["jq"] = o.Jq
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"key",
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

	varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 := _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4(varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "key")
		delete(additionalProperties, "jq")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 struct {
	value *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) Get() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) Set(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4 {
	return &NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf4) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


