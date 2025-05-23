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

// checks if the UpdateIntegrationRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateIntegrationRequest{}

// UpdateIntegrationRequest struct for UpdateIntegrationRequest
type UpdateIntegrationRequest struct {
	Data GetIntegration200ResponseData `json:"data"`
	// The name of the integration.
	Name string `json:"name"`
	// A location ID to use with this integration.
	LocationId *string `json:"locationId,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _UpdateIntegrationRequest UpdateIntegrationRequest

// NewUpdateIntegrationRequest instantiates a new UpdateIntegrationRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateIntegrationRequest(data GetIntegration200ResponseData, name string) *UpdateIntegrationRequest {
	this := UpdateIntegrationRequest{}
	this.Data = data
	this.Name = name
	return &this
}

// NewUpdateIntegrationRequestWithDefaults instantiates a new UpdateIntegrationRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateIntegrationRequestWithDefaults() *UpdateIntegrationRequest {
	this := UpdateIntegrationRequest{}
	return &this
}

// GetData returns the Data field value
func (o *UpdateIntegrationRequest) GetData() GetIntegration200ResponseData {
	if o == nil {
		var ret GetIntegration200ResponseData
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *UpdateIntegrationRequest) GetDataOk() (*GetIntegration200ResponseData, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Data, true
}

// SetData sets field value
func (o *UpdateIntegrationRequest) SetData(v GetIntegration200ResponseData) {
	o.Data = v
}

// GetName returns the Name field value
func (o *UpdateIntegrationRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UpdateIntegrationRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UpdateIntegrationRequest) SetName(v string) {
	o.Name = v
}

// GetLocationId returns the LocationId field value if set, zero value otherwise.
func (o *UpdateIntegrationRequest) GetLocationId() string {
	if o == nil || IsNil(o.LocationId) {
		var ret string
		return ret
	}
	return *o.LocationId
}

// GetLocationIdOk returns a tuple with the LocationId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *UpdateIntegrationRequest) GetLocationIdOk() (*string, bool) {
	if o == nil || IsNil(o.LocationId) {
		return nil, false
	}
	return o.LocationId, true
}

// HasLocationId returns a boolean if a field has been set.
func (o *UpdateIntegrationRequest) HasLocationId() bool {
	if o != nil && !IsNil(o.LocationId) {
		return true
	}

	return false
}

// SetLocationId gets a reference to the given string and assigns it to the LocationId field.
func (o *UpdateIntegrationRequest) SetLocationId(v string) {
	o.LocationId = &v
}

func (o UpdateIntegrationRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateIntegrationRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["data"] = o.Data
	toSerialize["name"] = o.Name
	if !IsNil(o.LocationId) {
		toSerialize["locationId"] = o.LocationId
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *UpdateIntegrationRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"data",
		"name",
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

	varUpdateIntegrationRequest := _UpdateIntegrationRequest{}

	err = json.Unmarshal(data, &varUpdateIntegrationRequest)

	if err != nil {
		return err
	}

	*o = UpdateIntegrationRequest(varUpdateIntegrationRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "data")
		delete(additionalProperties, "name")
		delete(additionalProperties, "locationId")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableUpdateIntegrationRequest struct {
	value *UpdateIntegrationRequest
	isSet bool
}

func (v NullableUpdateIntegrationRequest) Get() *UpdateIntegrationRequest {
	return v.value
}

func (v *NullableUpdateIntegrationRequest) Set(val *UpdateIntegrationRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateIntegrationRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateIntegrationRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateIntegrationRequest(val *UpdateIntegrationRequest) *NullableUpdateIntegrationRequest {
	return &NullableUpdateIntegrationRequest{value: val, isSet: true}
}

func (v NullableUpdateIntegrationRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateIntegrationRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


