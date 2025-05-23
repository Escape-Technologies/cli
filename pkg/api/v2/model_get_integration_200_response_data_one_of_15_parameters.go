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

// checks if the GetIntegration200ResponseDataOneOf15Parameters type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetIntegration200ResponseDataOneOf15Parameters{}

// GetIntegration200ResponseDataOneOf15Parameters struct for GetIntegration200ResponseDataOneOf15Parameters
type GetIntegration200ResponseDataOneOf15Parameters struct {
	Blacklist *GetIntegration200ResponseDataOneOf15ParametersBlacklist `json:"blacklist,omitempty"`
	Tags *GetIntegration200ResponseDataOneOf15ParametersTags `json:"tags,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _GetIntegration200ResponseDataOneOf15Parameters GetIntegration200ResponseDataOneOf15Parameters

// NewGetIntegration200ResponseDataOneOf15Parameters instantiates a new GetIntegration200ResponseDataOneOf15Parameters object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetIntegration200ResponseDataOneOf15Parameters() *GetIntegration200ResponseDataOneOf15Parameters {
	this := GetIntegration200ResponseDataOneOf15Parameters{}
	return &this
}

// NewGetIntegration200ResponseDataOneOf15ParametersWithDefaults instantiates a new GetIntegration200ResponseDataOneOf15Parameters object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetIntegration200ResponseDataOneOf15ParametersWithDefaults() *GetIntegration200ResponseDataOneOf15Parameters {
	this := GetIntegration200ResponseDataOneOf15Parameters{}
	return &this
}

// GetBlacklist returns the Blacklist field value if set, zero value otherwise.
func (o *GetIntegration200ResponseDataOneOf15Parameters) GetBlacklist() GetIntegration200ResponseDataOneOf15ParametersBlacklist {
	if o == nil || IsNil(o.Blacklist) {
		var ret GetIntegration200ResponseDataOneOf15ParametersBlacklist
		return ret
	}
	return *o.Blacklist
}

// GetBlacklistOk returns a tuple with the Blacklist field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf15Parameters) GetBlacklistOk() (*GetIntegration200ResponseDataOneOf15ParametersBlacklist, bool) {
	if o == nil || IsNil(o.Blacklist) {
		return nil, false
	}
	return o.Blacklist, true
}

// HasBlacklist returns a boolean if a field has been set.
func (o *GetIntegration200ResponseDataOneOf15Parameters) HasBlacklist() bool {
	if o != nil && !IsNil(o.Blacklist) {
		return true
	}

	return false
}

// SetBlacklist gets a reference to the given GetIntegration200ResponseDataOneOf15ParametersBlacklist and assigns it to the Blacklist field.
func (o *GetIntegration200ResponseDataOneOf15Parameters) SetBlacklist(v GetIntegration200ResponseDataOneOf15ParametersBlacklist) {
	o.Blacklist = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *GetIntegration200ResponseDataOneOf15Parameters) GetTags() GetIntegration200ResponseDataOneOf15ParametersTags {
	if o == nil || IsNil(o.Tags) {
		var ret GetIntegration200ResponseDataOneOf15ParametersTags
		return ret
	}
	return *o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetIntegration200ResponseDataOneOf15Parameters) GetTagsOk() (*GetIntegration200ResponseDataOneOf15ParametersTags, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *GetIntegration200ResponseDataOneOf15Parameters) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given GetIntegration200ResponseDataOneOf15ParametersTags and assigns it to the Tags field.
func (o *GetIntegration200ResponseDataOneOf15Parameters) SetTags(v GetIntegration200ResponseDataOneOf15ParametersTags) {
	o.Tags = &v
}

func (o GetIntegration200ResponseDataOneOf15Parameters) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetIntegration200ResponseDataOneOf15Parameters) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Blacklist) {
		toSerialize["blacklist"] = o.Blacklist
	}
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GetIntegration200ResponseDataOneOf15Parameters) UnmarshalJSON(data []byte) (err error) {
	varGetIntegration200ResponseDataOneOf15Parameters := _GetIntegration200ResponseDataOneOf15Parameters{}

	err = json.Unmarshal(data, &varGetIntegration200ResponseDataOneOf15Parameters)

	if err != nil {
		return err
	}

	*o = GetIntegration200ResponseDataOneOf15Parameters(varGetIntegration200ResponseDataOneOf15Parameters)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "blacklist")
		delete(additionalProperties, "tags")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGetIntegration200ResponseDataOneOf15Parameters struct {
	value *GetIntegration200ResponseDataOneOf15Parameters
	isSet bool
}

func (v NullableGetIntegration200ResponseDataOneOf15Parameters) Get() *GetIntegration200ResponseDataOneOf15Parameters {
	return v.value
}

func (v *NullableGetIntegration200ResponseDataOneOf15Parameters) Set(val *GetIntegration200ResponseDataOneOf15Parameters) {
	v.value = val
	v.isSet = true
}

func (v NullableGetIntegration200ResponseDataOneOf15Parameters) IsSet() bool {
	return v.isSet
}

func (v *NullableGetIntegration200ResponseDataOneOf15Parameters) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetIntegration200ResponseDataOneOf15Parameters(val *GetIntegration200ResponseDataOneOf15Parameters) *NullableGetIntegration200ResponseDataOneOf15Parameters {
	return &NullableGetIntegration200ResponseDataOneOf15Parameters{value: val, isSet: true}
}

func (v NullableGetIntegration200ResponseDataOneOf15Parameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetIntegration200ResponseDataOneOf15Parameters) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


