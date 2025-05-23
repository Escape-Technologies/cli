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

// checks if the CreateApplicationRequestAnyOf type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateApplicationRequestAnyOf{}

// CreateApplicationRequestAnyOf struct for CreateApplicationRequestAnyOf
type CreateApplicationRequestAnyOf struct {
	// Application name
	Name string `json:"name"`
	// Application URL
	Url string `json:"url"`
	Type EnumFRONTEND `json:"type"`
	Configuration *CreateApplicationRequestAnyOfConfiguration `json:"configuration,omitempty"`
	LocationId *string `json:"locationId,omitempty"`
	Cron *string `json:"cron,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CreateApplicationRequestAnyOf CreateApplicationRequestAnyOf

// NewCreateApplicationRequestAnyOf instantiates a new CreateApplicationRequestAnyOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateApplicationRequestAnyOf(name string, url string, type_ EnumFRONTEND) *CreateApplicationRequestAnyOf {
	this := CreateApplicationRequestAnyOf{}
	this.Name = name
	this.Url = url
	this.Type = type_
	return &this
}

// NewCreateApplicationRequestAnyOfWithDefaults instantiates a new CreateApplicationRequestAnyOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateApplicationRequestAnyOfWithDefaults() *CreateApplicationRequestAnyOf {
	this := CreateApplicationRequestAnyOf{}
	return &this
}

// GetName returns the Name field value
func (o *CreateApplicationRequestAnyOf) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateApplicationRequestAnyOf) SetName(v string) {
	o.Name = v
}

// GetUrl returns the Url field value
func (o *CreateApplicationRequestAnyOf) GetUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Url
}

// GetUrlOk returns a tuple with the Url field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Url, true
}

// SetUrl sets field value
func (o *CreateApplicationRequestAnyOf) SetUrl(v string) {
	o.Url = v
}

// GetType returns the Type field value
func (o *CreateApplicationRequestAnyOf) GetType() EnumFRONTEND {
	if o == nil {
		var ret EnumFRONTEND
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetTypeOk() (*EnumFRONTEND, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *CreateApplicationRequestAnyOf) SetType(v EnumFRONTEND) {
	o.Type = v
}

// GetConfiguration returns the Configuration field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOf) GetConfiguration() CreateApplicationRequestAnyOfConfiguration {
	if o == nil || IsNil(o.Configuration) {
		var ret CreateApplicationRequestAnyOfConfiguration
		return ret
	}
	return *o.Configuration
}

// GetConfigurationOk returns a tuple with the Configuration field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetConfigurationOk() (*CreateApplicationRequestAnyOfConfiguration, bool) {
	if o == nil || IsNil(o.Configuration) {
		return nil, false
	}
	return o.Configuration, true
}

// HasConfiguration returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOf) HasConfiguration() bool {
	if o != nil && !IsNil(o.Configuration) {
		return true
	}

	return false
}

// SetConfiguration gets a reference to the given CreateApplicationRequestAnyOfConfiguration and assigns it to the Configuration field.
func (o *CreateApplicationRequestAnyOf) SetConfiguration(v CreateApplicationRequestAnyOfConfiguration) {
	o.Configuration = &v
}

// GetLocationId returns the LocationId field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOf) GetLocationId() string {
	if o == nil || IsNil(o.LocationId) {
		var ret string
		return ret
	}
	return *o.LocationId
}

// GetLocationIdOk returns a tuple with the LocationId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetLocationIdOk() (*string, bool) {
	if o == nil || IsNil(o.LocationId) {
		return nil, false
	}
	return o.LocationId, true
}

// HasLocationId returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOf) HasLocationId() bool {
	if o != nil && !IsNil(o.LocationId) {
		return true
	}

	return false
}

// SetLocationId gets a reference to the given string and assigns it to the LocationId field.
func (o *CreateApplicationRequestAnyOf) SetLocationId(v string) {
	o.LocationId = &v
}

// GetCron returns the Cron field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOf) GetCron() string {
	if o == nil || IsNil(o.Cron) {
		var ret string
		return ret
	}
	return *o.Cron
}

// GetCronOk returns a tuple with the Cron field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOf) GetCronOk() (*string, bool) {
	if o == nil || IsNil(o.Cron) {
		return nil, false
	}
	return o.Cron, true
}

// HasCron returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOf) HasCron() bool {
	if o != nil && !IsNil(o.Cron) {
		return true
	}

	return false
}

// SetCron gets a reference to the given string and assigns it to the Cron field.
func (o *CreateApplicationRequestAnyOf) SetCron(v string) {
	o.Cron = &v
}

func (o CreateApplicationRequestAnyOf) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateApplicationRequestAnyOf) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	toSerialize["url"] = o.Url
	toSerialize["type"] = o.Type
	if !IsNil(o.Configuration) {
		toSerialize["configuration"] = o.Configuration
	}
	if !IsNil(o.LocationId) {
		toSerialize["locationId"] = o.LocationId
	}
	if !IsNil(o.Cron) {
		toSerialize["cron"] = o.Cron
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *CreateApplicationRequestAnyOf) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"url",
		"type",
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

	varCreateApplicationRequestAnyOf := _CreateApplicationRequestAnyOf{}

	err = json.Unmarshal(data, &varCreateApplicationRequestAnyOf)

	if err != nil {
		return err
	}

	*o = CreateApplicationRequestAnyOf(varCreateApplicationRequestAnyOf)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "name")
		delete(additionalProperties, "url")
		delete(additionalProperties, "type")
		delete(additionalProperties, "configuration")
		delete(additionalProperties, "locationId")
		delete(additionalProperties, "cron")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCreateApplicationRequestAnyOf struct {
	value *CreateApplicationRequestAnyOf
	isSet bool
}

func (v NullableCreateApplicationRequestAnyOf) Get() *CreateApplicationRequestAnyOf {
	return v.value
}

func (v *NullableCreateApplicationRequestAnyOf) Set(val *CreateApplicationRequestAnyOf) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateApplicationRequestAnyOf) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateApplicationRequestAnyOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateApplicationRequestAnyOf(val *CreateApplicationRequestAnyOf) *NullableCreateApplicationRequestAnyOf {
	return &NullableCreateApplicationRequestAnyOf{value: val, isSet: true}
}

func (v NullableCreateApplicationRequestAnyOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateApplicationRequestAnyOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


