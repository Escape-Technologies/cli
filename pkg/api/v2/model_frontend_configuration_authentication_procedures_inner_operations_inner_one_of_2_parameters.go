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

// checks if the FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters{}

// FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters struct for FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters
type FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters struct {
	Proxy *string `json:"proxy,omitempty"`
	LoginUrl string `json:"login_url"`
	AutoExtractionUrls []string `json:"auto_extraction_urls,omitempty"`
	LoggedInDetectorText *string `json:"logged_in_detector_text,omitempty"`
	LoggedInDetectorTimeout *float32 `json:"logged_in_detector_timeout,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters(loginUrl string) *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters{}
	this.LoginUrl = loginUrl
	return &this
}

// NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2ParametersWithDefaults instantiates a new FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2ParametersWithDefaults() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters {
	this := FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters{}
	return &this
}

// GetProxy returns the Proxy field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetProxy() string {
	if o == nil || IsNil(o.Proxy) {
		var ret string
		return ret
	}
	return *o.Proxy
}

// GetProxyOk returns a tuple with the Proxy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetProxyOk() (*string, bool) {
	if o == nil || IsNil(o.Proxy) {
		return nil, false
	}
	return o.Proxy, true
}

// HasProxy returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) HasProxy() bool {
	if o != nil && !IsNil(o.Proxy) {
		return true
	}

	return false
}

// SetProxy gets a reference to the given string and assigns it to the Proxy field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) SetProxy(v string) {
	o.Proxy = &v
}

// GetLoginUrl returns the LoginUrl field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoginUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LoginUrl
}

// GetLoginUrlOk returns a tuple with the LoginUrl field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoginUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LoginUrl, true
}

// SetLoginUrl sets field value
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) SetLoginUrl(v string) {
	o.LoginUrl = v
}

// GetAutoExtractionUrls returns the AutoExtractionUrls field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetAutoExtractionUrls() []string {
	if o == nil || IsNil(o.AutoExtractionUrls) {
		var ret []string
		return ret
	}
	return o.AutoExtractionUrls
}

// GetAutoExtractionUrlsOk returns a tuple with the AutoExtractionUrls field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetAutoExtractionUrlsOk() ([]string, bool) {
	if o == nil || IsNil(o.AutoExtractionUrls) {
		return nil, false
	}
	return o.AutoExtractionUrls, true
}

// HasAutoExtractionUrls returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) HasAutoExtractionUrls() bool {
	if o != nil && !IsNil(o.AutoExtractionUrls) {
		return true
	}

	return false
}

// SetAutoExtractionUrls gets a reference to the given []string and assigns it to the AutoExtractionUrls field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) SetAutoExtractionUrls(v []string) {
	o.AutoExtractionUrls = v
}

// GetLoggedInDetectorText returns the LoggedInDetectorText field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoggedInDetectorText() string {
	if o == nil || IsNil(o.LoggedInDetectorText) {
		var ret string
		return ret
	}
	return *o.LoggedInDetectorText
}

// GetLoggedInDetectorTextOk returns a tuple with the LoggedInDetectorText field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoggedInDetectorTextOk() (*string, bool) {
	if o == nil || IsNil(o.LoggedInDetectorText) {
		return nil, false
	}
	return o.LoggedInDetectorText, true
}

// HasLoggedInDetectorText returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) HasLoggedInDetectorText() bool {
	if o != nil && !IsNil(o.LoggedInDetectorText) {
		return true
	}

	return false
}

// SetLoggedInDetectorText gets a reference to the given string and assigns it to the LoggedInDetectorText field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) SetLoggedInDetectorText(v string) {
	o.LoggedInDetectorText = &v
}

// GetLoggedInDetectorTimeout returns the LoggedInDetectorTimeout field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoggedInDetectorTimeout() float32 {
	if o == nil || IsNil(o.LoggedInDetectorTimeout) {
		var ret float32
		return ret
	}
	return *o.LoggedInDetectorTimeout
}

// GetLoggedInDetectorTimeoutOk returns a tuple with the LoggedInDetectorTimeout field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) GetLoggedInDetectorTimeoutOk() (*float32, bool) {
	if o == nil || IsNil(o.LoggedInDetectorTimeout) {
		return nil, false
	}
	return o.LoggedInDetectorTimeout, true
}

// HasLoggedInDetectorTimeout returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) HasLoggedInDetectorTimeout() bool {
	if o != nil && !IsNil(o.LoggedInDetectorTimeout) {
		return true
	}

	return false
}

// SetLoggedInDetectorTimeout gets a reference to the given float32 and assigns it to the LoggedInDetectorTimeout field.
func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) SetLoggedInDetectorTimeout(v float32) {
	o.LoggedInDetectorTimeout = &v
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Proxy) {
		toSerialize["proxy"] = o.Proxy
	}
	toSerialize["login_url"] = o.LoginUrl
	if !IsNil(o.AutoExtractionUrls) {
		toSerialize["auto_extraction_urls"] = o.AutoExtractionUrls
	}
	if !IsNil(o.LoggedInDetectorText) {
		toSerialize["logged_in_detector_text"] = o.LoggedInDetectorText
	}
	if !IsNil(o.LoggedInDetectorTimeout) {
		toSerialize["logged_in_detector_timeout"] = o.LoggedInDetectorTimeout
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"login_url",
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

	varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters := _FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters(varFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "proxy")
		delete(additionalProperties, "login_url")
		delete(additionalProperties, "auto_extraction_urls")
		delete(additionalProperties, "logged_in_detector_text")
		delete(additionalProperties, "logged_in_detector_timeout")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters struct {
	value *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) Get() *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) Set(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters(val *FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters {
	return &NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf2Parameters) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


