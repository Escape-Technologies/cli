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

// checks if the FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3{}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 struct for FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
type FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 struct {
	Tech EnumBROWSERACTIONS `json:"tech"`
	Parameters FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters `json:"parameters"`
	Extractions FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions `json:"extractions"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3(tech EnumBROWSERACTIONS, parameters FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters, extractions FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions) *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3{}
	this.Tech = tech
	this.Parameters = parameters
	this.Extractions = extractions
	return &this
}

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3WithDefaults instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3WithDefaults() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3{}
	return &this
}

// GetTech returns the Tech field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetTech() EnumBROWSERACTIONS {
	if o == nil {
		var ret EnumBROWSERACTIONS
		return ret
	}

	return o.Tech
}

// GetTechOk returns a tuple with the Tech field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetTechOk() (*EnumBROWSERACTIONS, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Tech, true
}

// SetTech sets field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) SetTech(v EnumBROWSERACTIONS) {
	o.Tech = v
}

// GetParameters returns the Parameters field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetParameters() FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters {
	if o == nil {
		var ret FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters
		return ret
	}

	return o.Parameters
}

// GetParametersOk returns a tuple with the Parameters field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetParametersOk() (*FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Parameters, true
}

// SetParameters sets field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) SetParameters(v FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) {
	o.Parameters = v
}

// GetExtractions returns the Extractions field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetExtractions() FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions {
	if o == nil {
		var ret FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions
		return ret
	}

	return o.Extractions
}

// GetExtractionsOk returns a tuple with the Extractions field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) GetExtractionsOk() (*FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Extractions, true
}

// SetExtractions sets field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) SetExtractions(v FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1Extractions) {
	o.Extractions = v
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["tech"] = o.Tech
	toSerialize["parameters"] = o.Parameters
	toSerialize["extractions"] = o.Extractions

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"tech",
		"parameters",
		"extractions",
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

	varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 := _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3(varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "tech")
		delete(additionalProperties, "parameters")
		delete(additionalProperties, "extractions")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 struct {
	value *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) Get() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) Set(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3 {
	return &NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf3) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


