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

// checks if the GetScanExchangesArchive200Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetScanExchangesArchive200Response{}

// GetScanExchangesArchive200Response struct for GetScanExchangesArchive200Response
type GetScanExchangesArchive200Response struct {
	Archive string `json:"archive"`
	AdditionalProperties map[string]interface{}
}

type _GetScanExchangesArchive200Response GetScanExchangesArchive200Response

// NewGetScanExchangesArchive200Response instantiates a new GetScanExchangesArchive200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetScanExchangesArchive200Response(archive string) *GetScanExchangesArchive200Response {
	this := GetScanExchangesArchive200Response{}
	this.Archive = archive
	return &this
}

// NewGetScanExchangesArchive200ResponseWithDefaults instantiates a new GetScanExchangesArchive200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetScanExchangesArchive200ResponseWithDefaults() *GetScanExchangesArchive200Response {
	this := GetScanExchangesArchive200Response{}
	return &this
}

// GetArchive returns the Archive field value
func (o *GetScanExchangesArchive200Response) GetArchive() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Archive
}

// GetArchiveOk returns a tuple with the Archive field value
// and a boolean to check if the value has been set.
func (o *GetScanExchangesArchive200Response) GetArchiveOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Archive, true
}

// SetArchive sets field value
func (o *GetScanExchangesArchive200Response) SetArchive(v string) {
	o.Archive = v
}

func (o GetScanExchangesArchive200Response) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetScanExchangesArchive200Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["archive"] = o.Archive

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GetScanExchangesArchive200Response) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"archive",
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

	varGetScanExchangesArchive200Response := _GetScanExchangesArchive200Response{}

	err = json.Unmarshal(data, &varGetScanExchangesArchive200Response)

	if err != nil {
		return err
	}

	*o = GetScanExchangesArchive200Response(varGetScanExchangesArchive200Response)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "archive")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGetScanExchangesArchive200Response struct {
	value *GetScanExchangesArchive200Response
	isSet bool
}

func (v NullableGetScanExchangesArchive200Response) Get() *GetScanExchangesArchive200Response {
	return v.value
}

func (v *NullableGetScanExchangesArchive200Response) Set(val *GetScanExchangesArchive200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableGetScanExchangesArchive200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableGetScanExchangesArchive200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetScanExchangesArchive200Response(val *GetScanExchangesArchive200Response) *NullableGetScanExchangesArchive200Response {
	return &NullableGetScanExchangesArchive200Response{value: val, isSet: true}
}

func (v NullableGetScanExchangesArchive200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetScanExchangesArchive200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


