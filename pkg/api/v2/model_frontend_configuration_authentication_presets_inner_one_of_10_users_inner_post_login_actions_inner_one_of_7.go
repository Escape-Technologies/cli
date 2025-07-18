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

// checks if the FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7{}

// FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 struct for FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7
type FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 struct {
	Action EnumFOCUSPAGE `json:"action"`
	UrlPattern string `json:"url_pattern"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7(action EnumFOCUSPAGE, urlPattern string) *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7{}
	this.Action = action
	this.UrlPattern = urlPattern
	return &this
}

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7WithDefaults instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7WithDefaults() *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7{}
	return &this
}

// GetAction returns the Action field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) GetAction() EnumFOCUSPAGE {
	if o == nil {
		var ret EnumFOCUSPAGE
		return ret
	}

	return o.Action
}

// GetActionOk returns a tuple with the Action field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) GetActionOk() (*EnumFOCUSPAGE, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Action, true
}

// SetAction sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) SetAction(v EnumFOCUSPAGE) {
	o.Action = v
}

// GetUrlPattern returns the UrlPattern field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) GetUrlPattern() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UrlPattern
}

// GetUrlPatternOk returns a tuple with the UrlPattern field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) GetUrlPatternOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UrlPattern, true
}

// SetUrlPattern sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) SetUrlPattern(v string) {
	o.UrlPattern = v
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["action"] = o.Action
	toSerialize["url_pattern"] = o.UrlPattern

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"action",
		"url_pattern",
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

	varFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 := _FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7(varFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "action")
		delete(additionalProperties, "url_pattern")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 struct {
	value *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) Get() *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) Set(val *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7(val *FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7 {
	return &NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInnerPostLoginActionsInnerOneOf7) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


