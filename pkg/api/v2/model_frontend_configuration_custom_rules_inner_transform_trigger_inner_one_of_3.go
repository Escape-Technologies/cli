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

// checks if the FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3{}

// FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 struct for FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3
type FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 struct {
	Is *float32 `json:"is,omitempty"`
	IsNot *float32 `json:"is_not,omitempty"`
	In []float32 `json:"in,omitempty"`
	Gt *float32 `json:"gt,omitempty"`
	Lt *float32 `json:"lt,omitempty"`
	If EnumRESPONSEDURATIONMS `json:"if"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3(if_ EnumRESPONSEDURATIONMS) *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3{}
	this.If = if_
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3WithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3WithDefaults() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 {
	this := FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3{}
	return &this
}

// GetIs returns the Is field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIs() float32 {
	if o == nil || IsNil(o.Is) {
		var ret float32
		return ret
	}
	return *o.Is
}

// GetIsOk returns a tuple with the Is field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIsOk() (*float32, bool) {
	if o == nil || IsNil(o.Is) {
		return nil, false
	}
	return o.Is, true
}

// HasIs returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) HasIs() bool {
	if o != nil && !IsNil(o.Is) {
		return true
	}

	return false
}

// SetIs gets a reference to the given float32 and assigns it to the Is field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetIs(v float32) {
	o.Is = &v
}

// GetIsNot returns the IsNot field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIsNot() float32 {
	if o == nil || IsNil(o.IsNot) {
		var ret float32
		return ret
	}
	return *o.IsNot
}

// GetIsNotOk returns a tuple with the IsNot field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIsNotOk() (*float32, bool) {
	if o == nil || IsNil(o.IsNot) {
		return nil, false
	}
	return o.IsNot, true
}

// HasIsNot returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) HasIsNot() bool {
	if o != nil && !IsNil(o.IsNot) {
		return true
	}

	return false
}

// SetIsNot gets a reference to the given float32 and assigns it to the IsNot field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetIsNot(v float32) {
	o.IsNot = &v
}

// GetIn returns the In field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIn() []float32 {
	if o == nil || IsNil(o.In) {
		var ret []float32
		return ret
	}
	return o.In
}

// GetInOk returns a tuple with the In field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetInOk() ([]float32, bool) {
	if o == nil || IsNil(o.In) {
		return nil, false
	}
	return o.In, true
}

// HasIn returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) HasIn() bool {
	if o != nil && !IsNil(o.In) {
		return true
	}

	return false
}

// SetIn gets a reference to the given []float32 and assigns it to the In field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetIn(v []float32) {
	o.In = v
}

// GetGt returns the Gt field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetGt() float32 {
	if o == nil || IsNil(o.Gt) {
		var ret float32
		return ret
	}
	return *o.Gt
}

// GetGtOk returns a tuple with the Gt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetGtOk() (*float32, bool) {
	if o == nil || IsNil(o.Gt) {
		return nil, false
	}
	return o.Gt, true
}

// HasGt returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) HasGt() bool {
	if o != nil && !IsNil(o.Gt) {
		return true
	}

	return false
}

// SetGt gets a reference to the given float32 and assigns it to the Gt field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetGt(v float32) {
	o.Gt = &v
}

// GetLt returns the Lt field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetLt() float32 {
	if o == nil || IsNil(o.Lt) {
		var ret float32
		return ret
	}
	return *o.Lt
}

// GetLtOk returns a tuple with the Lt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetLtOk() (*float32, bool) {
	if o == nil || IsNil(o.Lt) {
		return nil, false
	}
	return o.Lt, true
}

// HasLt returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) HasLt() bool {
	if o != nil && !IsNil(o.Lt) {
		return true
	}

	return false
}

// SetLt gets a reference to the given float32 and assigns it to the Lt field.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetLt(v float32) {
	o.Lt = &v
}

// GetIf returns the If field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIf() EnumRESPONSEDURATIONMS {
	if o == nil {
		var ret EnumRESPONSEDURATIONMS
		return ret
	}

	return o.If
}

// GetIfOk returns a tuple with the If field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) GetIfOk() (*EnumRESPONSEDURATIONMS, bool) {
	if o == nil {
		return nil, false
	}
	return &o.If, true
}

// SetIf sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) SetIf(v EnumRESPONSEDURATIONMS) {
	o.If = v
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) ToMap() (map[string]interface{}, error) {
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
	if !IsNil(o.Gt) {
		toSerialize["gt"] = o.Gt
	}
	if !IsNil(o.Lt) {
		toSerialize["lt"] = o.Lt
	}
	toSerialize["if"] = o.If

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) UnmarshalJSON(data []byte) (err error) {
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

	varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 := _FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3(varFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "is")
		delete(additionalProperties, "is_not")
		delete(additionalProperties, "in")
		delete(additionalProperties, "gt")
		delete(additionalProperties, "lt")
		delete(additionalProperties, "if")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 struct {
	value *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) Get() *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) Set(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3(val *FrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3 {
	return &NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformTriggerInnerOneOf3) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


