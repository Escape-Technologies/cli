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

// checks if the CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf{}

// CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf struct for CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf
type CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf struct {
	Type EnumHTTP `json:"type"`
	Users []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner `json:"users"`
	Request CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest `json:"request"`
	Extractions []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner `json:"extractions,omitempty"`
	Injections []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner `json:"injections,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf

// NewCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf instantiates a new CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf(type_ EnumHTTP, users []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner, request CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest) *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf {
	this := CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf{}
	this.Type = type_
	this.Users = users
	this.Request = request
	return &this
}

// NewCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfWithDefaults instantiates a new CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfWithDefaults() *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf {
	this := CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf{}
	return &this
}

// GetType returns the Type field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetType() EnumHTTP {
	if o == nil {
		var ret EnumHTTP
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetTypeOk() (*EnumHTTP, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) SetType(v EnumHTTP) {
	o.Type = v
}

// GetUsers returns the Users field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetUsers() []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner {
	if o == nil {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner
		return ret
	}

	return o.Users
}

// GetUsersOk returns a tuple with the Users field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetUsersOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner, bool) {
	if o == nil {
		return nil, false
	}
	return o.Users, true
}

// SetUsers sets field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) SetUsers(v []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfUsersInner) {
	o.Users = v
}

// GetRequest returns the Request field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetRequest() CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest {
	if o == nil {
		var ret CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest
		return ret
	}

	return o.Request
}

// GetRequestOk returns a tuple with the Request field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetRequestOk() (*CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Request, true
}

// SetRequest sets field value
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) SetRequest(v CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOfRequest) {
	o.Request = v
}

// GetExtractions returns the Extractions field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetExtractions() []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner {
	if o == nil || IsNil(o.Extractions) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner
		return ret
	}
	return o.Extractions
}

// GetExtractionsOk returns a tuple with the Extractions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetExtractionsOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner, bool) {
	if o == nil || IsNil(o.Extractions) {
		return nil, false
	}
	return o.Extractions, true
}

// HasExtractions returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) HasExtractions() bool {
	if o != nil && !IsNil(o.Extractions) {
		return true
	}

	return false
}

// SetExtractions gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner and assigns it to the Extractions field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) SetExtractions(v []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfExtractionsInner) {
	o.Extractions = v
}

// GetInjections returns the Injections field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetInjections() []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner {
	if o == nil || IsNil(o.Injections) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner
		return ret
	}
	return o.Injections
}

// GetInjectionsOk returns a tuple with the Injections field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) GetInjectionsOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner, bool) {
	if o == nil || IsNil(o.Injections) {
		return nil, false
	}
	return o.Injections, true
}

// HasInjections returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) HasInjections() bool {
	if o != nil && !IsNil(o.Injections) {
		return true
	}

	return false
}

// SetInjections gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner and assigns it to the Injections field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) SetInjections(v []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerInjectionsAnyOfInner) {
	o.Injections = v
}

func (o CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["type"] = o.Type
	toSerialize["users"] = o.Users
	toSerialize["request"] = o.Request
	if !IsNil(o.Extractions) {
		toSerialize["extractions"] = o.Extractions
	}
	if !IsNil(o.Injections) {
		toSerialize["injections"] = o.Injections
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"type",
		"users",
		"request",
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

	varCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf := _CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf{}

	err = json.Unmarshal(data, &varCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf)

	if err != nil {
		return err
	}

	*o = CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf(varCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "type")
		delete(additionalProperties, "users")
		delete(additionalProperties, "request")
		delete(additionalProperties, "extractions")
		delete(additionalProperties, "injections")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf struct {
	value *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf
	isSet bool
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) Get() *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf {
	return v.value
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) Set(val *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf(val *CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf {
	return &NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf{value: val, isSet: true}
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


