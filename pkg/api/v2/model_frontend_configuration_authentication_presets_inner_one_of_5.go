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

// checks if the FrontendConfigurationAuthenticationPresetsInnerOneOf5 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationPresetsInnerOneOf5{}

// FrontendConfigurationAuthenticationPresetsInnerOneOf5 struct for FrontendConfigurationAuthenticationPresetsInnerOneOf5
type FrontendConfigurationAuthenticationPresetsInnerOneOf5 struct {
	Type EnumDIGEST `json:"type"`
	Users []FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner `json:"users"`
	FirstRequest FrontendConfigurationAuthenticationPresetsInnerOneOfRequest `json:"first_request"`
	SecondRequest *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest `json:"second_request,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationPresetsInnerOneOf5 FrontendConfigurationAuthenticationPresetsInnerOneOf5

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf5 instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf5 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf5(type_ EnumDIGEST, users []FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner, firstRequest FrontendConfigurationAuthenticationPresetsInnerOneOfRequest) *FrontendConfigurationAuthenticationPresetsInnerOneOf5 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf5{}
	this.Type = type_
	this.Users = users
	this.FirstRequest = firstRequest
	return &this
}

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf5WithDefaults instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf5 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf5WithDefaults() *FrontendConfigurationAuthenticationPresetsInnerOneOf5 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf5{}
	return &this
}

// GetType returns the Type field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetType() EnumDIGEST {
	if o == nil {
		var ret EnumDIGEST
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetTypeOk() (*EnumDIGEST, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) SetType(v EnumDIGEST) {
	o.Type = v
}

// GetUsers returns the Users field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetUsers() []FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner {
	if o == nil {
		var ret []FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner
		return ret
	}

	return o.Users
}

// GetUsersOk returns a tuple with the Users field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetUsersOk() ([]FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner, bool) {
	if o == nil {
		return nil, false
	}
	return o.Users, true
}

// SetUsers sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) SetUsers(v []FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) {
	o.Users = v
}

// GetFirstRequest returns the FirstRequest field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetFirstRequest() FrontendConfigurationAuthenticationPresetsInnerOneOfRequest {
	if o == nil {
		var ret FrontendConfigurationAuthenticationPresetsInnerOneOfRequest
		return ret
	}

	return o.FirstRequest
}

// GetFirstRequestOk returns a tuple with the FirstRequest field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetFirstRequestOk() (*FrontendConfigurationAuthenticationPresetsInnerOneOfRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.FirstRequest, true
}

// SetFirstRequest sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) SetFirstRequest(v FrontendConfigurationAuthenticationPresetsInnerOneOfRequest) {
	o.FirstRequest = v
}

// GetSecondRequest returns the SecondRequest field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetSecondRequest() FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest {
	if o == nil || IsNil(o.SecondRequest) {
		var ret FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest
		return ret
	}
	return *o.SecondRequest
}

// GetSecondRequestOk returns a tuple with the SecondRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) GetSecondRequestOk() (*FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest, bool) {
	if o == nil || IsNil(o.SecondRequest) {
		return nil, false
	}
	return o.SecondRequest, true
}

// HasSecondRequest returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) HasSecondRequest() bool {
	if o != nil && !IsNil(o.SecondRequest) {
		return true
	}

	return false
}

// SetSecondRequest gets a reference to the given FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest and assigns it to the SecondRequest field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) SetSecondRequest(v FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ParametersSecondRequest) {
	o.SecondRequest = &v
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf5) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf5) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["type"] = o.Type
	toSerialize["users"] = o.Users
	toSerialize["first_request"] = o.FirstRequest
	if !IsNil(o.SecondRequest) {
		toSerialize["second_request"] = o.SecondRequest
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf5) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"type",
		"users",
		"first_request",
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

	varFrontendConfigurationAuthenticationPresetsInnerOneOf5 := _FrontendConfigurationAuthenticationPresetsInnerOneOf5{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationPresetsInnerOneOf5)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationPresetsInnerOneOf5(varFrontendConfigurationAuthenticationPresetsInnerOneOf5)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "type")
		delete(additionalProperties, "users")
		delete(additionalProperties, "first_request")
		delete(additionalProperties, "second_request")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5 struct {
	value *FrontendConfigurationAuthenticationPresetsInnerOneOf5
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) Get() *FrontendConfigurationAuthenticationPresetsInnerOneOf5 {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) Set(val *FrontendConfigurationAuthenticationPresetsInnerOneOf5) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationPresetsInnerOneOf5(val *FrontendConfigurationAuthenticationPresetsInnerOneOf5) *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5 {
	return &NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf5) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


