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

// checks if the FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest{}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest struct for FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest
type FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest struct {
	Url *string `json:"url,omitempty"`
	Method *Enum4e0943c4ae7a2a2d426c0a6c0b839e82 `json:"method,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest{}
	return &this
}

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequestWithDefaults instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequestWithDefaults() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest{}
	return &this
}

// GetUrl returns the Url field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) GetUrl() string {
	if o == nil || IsNil(o.Url) {
		var ret string
		return ret
	}
	return *o.Url
}

// GetUrlOk returns a tuple with the Url field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) GetUrlOk() (*string, bool) {
	if o == nil || IsNil(o.Url) {
		return nil, false
	}
	return o.Url, true
}

// HasUrl returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) HasUrl() bool {
	if o != nil && !IsNil(o.Url) {
		return true
	}

	return false
}

// SetUrl gets a reference to the given string and assigns it to the Url field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) SetUrl(v string) {
	o.Url = &v
}

// GetMethod returns the Method field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) GetMethod() Enum4e0943c4ae7a2a2d426c0a6c0b839e82 {
	if o == nil || IsNil(o.Method) {
		var ret Enum4e0943c4ae7a2a2d426c0a6c0b839e82
		return ret
	}
	return *o.Method
}

// GetMethodOk returns a tuple with the Method field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) GetMethodOk() (*Enum4e0943c4ae7a2a2d426c0a6c0b839e82, bool) {
	if o == nil || IsNil(o.Method) {
		return nil, false
	}
	return o.Method, true
}

// HasMethod returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) HasMethod() bool {
	if o != nil && !IsNil(o.Method) {
		return true
	}

	return false
}

// SetMethod gets a reference to the given Enum4e0943c4ae7a2a2d426c0a6c0b839e82 and assigns it to the Method field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) SetMethod(v Enum4e0943c4ae7a2a2d426c0a6c0b839e82) {
	o.Method = &v
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Url) {
		toSerialize["url"] = o.Url
	}
	if !IsNil(o.Method) {
		toSerialize["method"] = o.Method
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) UnmarshalJSON(data []byte) (err error) {
	varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest := _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest(varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "url")
		delete(additionalProperties, "method")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest struct {
	value *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) Get() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) Set(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest {
	return &NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


