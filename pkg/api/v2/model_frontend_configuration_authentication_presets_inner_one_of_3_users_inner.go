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

// checks if the FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner{}

// FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner struct for FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner
type FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner struct {
	Username string `json:"username"`
	Headers map[string]string `json:"headers,omitempty"`
	Cookies map[string]string `json:"cookies,omitempty"`
	Password string `json:"password"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner(username string, password string) *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner{}
	this.Username = username
	this.Password = password
	return &this
}

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInnerWithDefaults instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInnerWithDefaults() *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner{}
	return &this
}

// GetUsername returns the Username field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetUsername() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Username
}

// GetUsernameOk returns a tuple with the Username field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetUsernameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Username, true
}

// SetUsername sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) SetUsername(v string) {
	o.Username = v
}

// GetHeaders returns the Headers field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetHeaders() map[string]string {
	if o == nil || IsNil(o.Headers) {
		var ret map[string]string
		return ret
	}
	return o.Headers
}

// GetHeadersOk returns a tuple with the Headers field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetHeadersOk() (map[string]string, bool) {
	if o == nil || IsNil(o.Headers) {
		return map[string]string{}, false
	}
	return o.Headers, true
}

// HasHeaders returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) HasHeaders() bool {
	if o != nil && !IsNil(o.Headers) {
		return true
	}

	return false
}

// SetHeaders gets a reference to the given map[string]string and assigns it to the Headers field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) SetHeaders(v map[string]string) {
	o.Headers = v
}

// GetCookies returns the Cookies field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetCookies() map[string]string {
	if o == nil || IsNil(o.Cookies) {
		var ret map[string]string
		return ret
	}
	return o.Cookies
}

// GetCookiesOk returns a tuple with the Cookies field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetCookiesOk() (map[string]string, bool) {
	if o == nil || IsNil(o.Cookies) {
		return map[string]string{}, false
	}
	return o.Cookies, true
}

// HasCookies returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) HasCookies() bool {
	if o != nil && !IsNil(o.Cookies) {
		return true
	}

	return false
}

// SetCookies gets a reference to the given map[string]string and assigns it to the Cookies field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) SetCookies(v map[string]string) {
	o.Cookies = v
}

// GetPassword returns the Password field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetPassword() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Password
}

// GetPasswordOk returns a tuple with the Password field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) GetPasswordOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Password, true
}

// SetPassword sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) SetPassword(v string) {
	o.Password = v
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["username"] = o.Username
	if !IsNil(o.Headers) {
		toSerialize["headers"] = o.Headers
	}
	if !IsNil(o.Cookies) {
		toSerialize["cookies"] = o.Cookies
	}
	toSerialize["password"] = o.Password

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"username",
		"password",
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

	varFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner := _FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner(varFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "username")
		delete(additionalProperties, "headers")
		delete(additionalProperties, "cookies")
		delete(additionalProperties, "password")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner struct {
	value *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) Get() *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) Set(val *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner(val *FrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner {
	return &NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf3UsersInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


