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

// checks if the CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner{}

// CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner struct for CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner
type CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner struct {
	Method *string `json:"method,omitempty"`
	Path *string `json:"path,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner

// NewCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner instantiates a new CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner() *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner {
	this := CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner{}
	return &this
}

// NewCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInnerWithDefaults instantiates a new CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInnerWithDefaults() *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner {
	this := CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner{}
	return &this
}

// GetMethod returns the Method field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) GetMethod() string {
	if o == nil || IsNil(o.Method) {
		var ret string
		return ret
	}
	return *o.Method
}

// GetMethodOk returns a tuple with the Method field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) GetMethodOk() (*string, bool) {
	if o == nil || IsNil(o.Method) {
		return nil, false
	}
	return o.Method, true
}

// HasMethod returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) HasMethod() bool {
	if o != nil && !IsNil(o.Method) {
		return true
	}

	return false
}

// SetMethod gets a reference to the given string and assigns it to the Method field.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) SetMethod(v string) {
	o.Method = &v
}

// GetPath returns the Path field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) GetPath() string {
	if o == nil || IsNil(o.Path) {
		var ret string
		return ret
	}
	return *o.Path
}

// GetPathOk returns a tuple with the Path field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) GetPathOk() (*string, bool) {
	if o == nil || IsNil(o.Path) {
		return nil, false
	}
	return o.Path, true
}

// HasPath returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) HasPath() bool {
	if o != nil && !IsNil(o.Path) {
		return true
	}

	return false
}

// SetPath gets a reference to the given string and assigns it to the Path field.
func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) SetPath(v string) {
	o.Path = &v
}

func (o CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Method) {
		toSerialize["method"] = o.Method
	}
	if !IsNil(o.Path) {
		toSerialize["path"] = o.Path
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) UnmarshalJSON(data []byte) (err error) {
	varCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner := _CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner{}

	err = json.Unmarshal(data, &varCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner)

	if err != nil {
		return err
	}

	*o = CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner(varCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "method")
		delete(additionalProperties, "path")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner struct {
	value *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner
	isSet bool
}

func (v NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) Get() *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner {
	return v.value
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) Set(val *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner(val *CreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) *NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner {
	return &NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner{value: val, isSet: true}
}

func (v NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationScanBlocklistRoutesInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


