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

// checks if the GetIntegration200ResponseDataOneOf8 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetIntegration200ResponseDataOneOf8{}

// GetIntegration200ResponseDataOneOf8 struct for GetIntegration200ResponseDataOneOf8
type GetIntegration200ResponseDataOneOf8 struct {
	Kind EnumEMAIL `json:"kind"`
	Parameters map[string]interface{} `json:"parameters"`
	AdditionalProperties map[string]interface{}
}

type _GetIntegration200ResponseDataOneOf8 GetIntegration200ResponseDataOneOf8

// NewGetIntegration200ResponseDataOneOf8 instantiates a new GetIntegration200ResponseDataOneOf8 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetIntegration200ResponseDataOneOf8(kind EnumEMAIL, parameters map[string]interface{}) *GetIntegration200ResponseDataOneOf8 {
	this := GetIntegration200ResponseDataOneOf8{}
	this.Kind = kind
	this.Parameters = parameters
	return &this
}

// NewGetIntegration200ResponseDataOneOf8WithDefaults instantiates a new GetIntegration200ResponseDataOneOf8 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetIntegration200ResponseDataOneOf8WithDefaults() *GetIntegration200ResponseDataOneOf8 {
	this := GetIntegration200ResponseDataOneOf8{}
	return &this
}

// GetKind returns the Kind field value
func (o *GetIntegration200ResponseDataOneOf8) GetKind() EnumEMAIL {
	if o == nil {
		var ret EnumEMAIL
		return ret
	}

	return o.Kind
}

// GetKindOk returns a tuple with the Kind field value
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf8) GetKindOk() (*EnumEMAIL, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Kind, true
}

// SetKind sets field value
func (o *GetIntegration200ResponseDataOneOf8) SetKind(v EnumEMAIL) {
	o.Kind = v
}

// GetParameters returns the Parameters field value
func (o *GetIntegration200ResponseDataOneOf8) GetParameters() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Parameters
}

// GetParametersOk returns a tuple with the Parameters field value
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf8) GetParametersOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Parameters, true
}

// SetParameters sets field value
func (o *GetIntegration200ResponseDataOneOf8) SetParameters(v map[string]interface{}) {
	o.Parameters = v
}

func (o GetIntegration200ResponseDataOneOf8) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetIntegration200ResponseDataOneOf8) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["kind"] = o.Kind
	toSerialize["parameters"] = o.Parameters

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GetIntegration200ResponseDataOneOf8) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"kind",
		"parameters",
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

	varGetIntegration200ResponseDataOneOf8 := _GetIntegration200ResponseDataOneOf8{}

	err = json.Unmarshal(data, &varGetIntegration200ResponseDataOneOf8)

	if err != nil {
		return err
	}

	*o = GetIntegration200ResponseDataOneOf8(varGetIntegration200ResponseDataOneOf8)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "kind")
		delete(additionalProperties, "parameters")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGetIntegration200ResponseDataOneOf8 struct {
	value *GetIntegration200ResponseDataOneOf8
	isSet bool
}

func (v NullableGetIntegration200ResponseDataOneOf8) Get() *GetIntegration200ResponseDataOneOf8 {
	return v.value
}

func (v *NullableGetIntegration200ResponseDataOneOf8) Set(val *GetIntegration200ResponseDataOneOf8) {
	v.value = val
	v.isSet = true
}

func (v NullableGetIntegration200ResponseDataOneOf8) IsSet() bool {
	return v.isSet
}

func (v *NullableGetIntegration200ResponseDataOneOf8) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetIntegration200ResponseDataOneOf8(val *GetIntegration200ResponseDataOneOf8) *NullableGetIntegration200ResponseDataOneOf8 {
	return &NullableGetIntegration200ResponseDataOneOf8{value: val, isSet: true}
}

func (v NullableGetIntegration200ResponseDataOneOf8) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetIntegration200ResponseDataOneOf8) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


