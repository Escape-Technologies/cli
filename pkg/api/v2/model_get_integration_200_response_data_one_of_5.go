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

// checks if the GetIntegration200ResponseDataOneOf5 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetIntegration200ResponseDataOneOf5{}

// GetIntegration200ResponseDataOneOf5 struct for GetIntegration200ResponseDataOneOf5
type GetIntegration200ResponseDataOneOf5 struct {
	Kind EnumBITBUCKETREPO `json:"kind"`
	Parameters GetIntegration200ResponseDataOneOf5Parameters `json:"parameters"`
	AdditionalProperties map[string]interface{}
}

type _GetIntegration200ResponseDataOneOf5 GetIntegration200ResponseDataOneOf5

// NewGetIntegration200ResponseDataOneOf5 instantiates a new GetIntegration200ResponseDataOneOf5 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetIntegration200ResponseDataOneOf5(kind EnumBITBUCKETREPO, parameters GetIntegration200ResponseDataOneOf5Parameters) *GetIntegration200ResponseDataOneOf5 {
	this := GetIntegration200ResponseDataOneOf5{}
	this.Kind = kind
	this.Parameters = parameters
	return &this
}

// NewGetIntegration200ResponseDataOneOf5WithDefaults instantiates a new GetIntegration200ResponseDataOneOf5 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetIntegration200ResponseDataOneOf5WithDefaults() *GetIntegration200ResponseDataOneOf5 {
	this := GetIntegration200ResponseDataOneOf5{}
	return &this
}

// GetKind returns the Kind field value
func (o *GetIntegration200ResponseDataOneOf5) GetKind() EnumBITBUCKETREPO {
	if o == nil {
		var ret EnumBITBUCKETREPO
		return ret
	}

	return o.Kind
}

// GetKindOk returns a tuple with the Kind field value
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf5) GetKindOk() (*EnumBITBUCKETREPO, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Kind, true
}

// SetKind sets field value
func (o *GetIntegration200ResponseDataOneOf5) SetKind(v EnumBITBUCKETREPO) {
	o.Kind = v
}

// GetParameters returns the Parameters field value
func (o *GetIntegration200ResponseDataOneOf5) GetParameters() GetIntegration200ResponseDataOneOf5Parameters {
	if o == nil {
		var ret GetIntegration200ResponseDataOneOf5Parameters
		return ret
	}

	return o.Parameters
}

// GetParametersOk returns a tuple with the Parameters field value
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf5) GetParametersOk() (*GetIntegration200ResponseDataOneOf5Parameters, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Parameters, true
}

// SetParameters sets field value
func (o *GetIntegration200ResponseDataOneOf5) SetParameters(v GetIntegration200ResponseDataOneOf5Parameters) {
	o.Parameters = v
}

func (o GetIntegration200ResponseDataOneOf5) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetIntegration200ResponseDataOneOf5) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["kind"] = o.Kind
	toSerialize["parameters"] = o.Parameters

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GetIntegration200ResponseDataOneOf5) UnmarshalJSON(data []byte) (err error) {
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

	varGetIntegration200ResponseDataOneOf5 := _GetIntegration200ResponseDataOneOf5{}

	err = json.Unmarshal(data, &varGetIntegration200ResponseDataOneOf5)

	if err != nil {
		return err
	}

	*o = GetIntegration200ResponseDataOneOf5(varGetIntegration200ResponseDataOneOf5)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "kind")
		delete(additionalProperties, "parameters")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGetIntegration200ResponseDataOneOf5 struct {
	value *GetIntegration200ResponseDataOneOf5
	isSet bool
}

func (v NullableGetIntegration200ResponseDataOneOf5) Get() *GetIntegration200ResponseDataOneOf5 {
	return v.value
}

func (v *NullableGetIntegration200ResponseDataOneOf5) Set(val *GetIntegration200ResponseDataOneOf5) {
	v.value = val
	v.isSet = true
}

func (v NullableGetIntegration200ResponseDataOneOf5) IsSet() bool {
	return v.isSet
}

func (v *NullableGetIntegration200ResponseDataOneOf5) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetIntegration200ResponseDataOneOf5(val *GetIntegration200ResponseDataOneOf5) *NullableGetIntegration200ResponseDataOneOf5 {
	return &NullableGetIntegration200ResponseDataOneOf5{value: val, isSet: true}
}

func (v NullableGetIntegration200ResponseDataOneOf5) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetIntegration200ResponseDataOneOf5) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


