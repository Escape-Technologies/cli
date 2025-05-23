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

// checks if the CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type{}

// CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type struct for CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type
type CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type struct {
	Is *Enum88381acf2baeed408becef5653510069 `json:"is,omitempty"`
	IsNot *Enum88381acf2baeed408becef5653510069 `json:"is_not,omitempty"`
	In []Enum88381acf2baeed408becef5653510069 `json:"in,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type

// NewCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type instantiates a new CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type() *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type {
	this := CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type{}
	return &this
}

// NewCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16TypeWithDefaults instantiates a new CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16TypeWithDefaults() *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type {
	this := CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type{}
	return &this
}

// GetIs returns the Is field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetIs() Enum88381acf2baeed408becef5653510069 {
	if o == nil || IsNil(o.Is) {
		var ret Enum88381acf2baeed408becef5653510069
		return ret
	}
	return *o.Is
}

// GetIsOk returns a tuple with the Is field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetIsOk() (*Enum88381acf2baeed408becef5653510069, bool) {
	if o == nil || IsNil(o.Is) {
		return nil, false
	}
	return o.Is, true
}

// HasIs returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) HasIs() bool {
	if o != nil && !IsNil(o.Is) {
		return true
	}

	return false
}

// SetIs gets a reference to the given Enum88381acf2baeed408becef5653510069 and assigns it to the Is field.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) SetIs(v Enum88381acf2baeed408becef5653510069) {
	o.Is = &v
}

// GetIsNot returns the IsNot field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetIsNot() Enum88381acf2baeed408becef5653510069 {
	if o == nil || IsNil(o.IsNot) {
		var ret Enum88381acf2baeed408becef5653510069
		return ret
	}
	return *o.IsNot
}

// GetIsNotOk returns a tuple with the IsNot field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetIsNotOk() (*Enum88381acf2baeed408becef5653510069, bool) {
	if o == nil || IsNil(o.IsNot) {
		return nil, false
	}
	return o.IsNot, true
}

// HasIsNot returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) HasIsNot() bool {
	if o != nil && !IsNil(o.IsNot) {
		return true
	}

	return false
}

// SetIsNot gets a reference to the given Enum88381acf2baeed408becef5653510069 and assigns it to the IsNot field.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) SetIsNot(v Enum88381acf2baeed408becef5653510069) {
	o.IsNot = &v
}

// GetIn returns the In field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetIn() []Enum88381acf2baeed408becef5653510069 {
	if o == nil || IsNil(o.In) {
		var ret []Enum88381acf2baeed408becef5653510069
		return ret
	}
	return o.In
}

// GetInOk returns a tuple with the In field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) GetInOk() ([]Enum88381acf2baeed408becef5653510069, bool) {
	if o == nil || IsNil(o.In) {
		return nil, false
	}
	return o.In, true
}

// HasIn returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) HasIn() bool {
	if o != nil && !IsNil(o.In) {
		return true
	}

	return false
}

// SetIn gets a reference to the given []Enum88381acf2baeed408becef5653510069 and assigns it to the In field.
func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) SetIn(v []Enum88381acf2baeed408becef5653510069) {
	o.In = v
}

func (o CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) ToMap() (map[string]interface{}, error) {
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

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) UnmarshalJSON(data []byte) (err error) {
	varCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type := _CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type{}

	err = json.Unmarshal(data, &varCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type)

	if err != nil {
		return err
	}

	*o = CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type(varCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "is")
		delete(additionalProperties, "is_not")
		delete(additionalProperties, "in")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type struct {
	value *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type
	isSet bool
}

func (v NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) Get() *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type {
	return v.value
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) Set(val *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type(val *CreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) *NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type {
	return &NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type{value: val, isSet: true}
}

func (v NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationCustomRulesInnerTransformTriggerInnerOneOf16Type) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


