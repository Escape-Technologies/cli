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

// checks if the FrontendConfigurationAuthenticationPresetsInnerOneOf10 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationAuthenticationPresetsInnerOneOf10{}

// FrontendConfigurationAuthenticationPresetsInnerOneOf10 struct for FrontendConfigurationAuthenticationPresetsInnerOneOf10
type FrontendConfigurationAuthenticationPresetsInnerOneOf10 struct {
	Type EnumBROWSERAGENT `json:"type"`
	Users []FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner `json:"users"`
	LoginUrl string `json:"login_url"`
	LoggedInDetectorText *string `json:"logged_in_detector_text,omitempty"`
	LoggedInDetectorTimeout *float32 `json:"logged_in_detector_timeout,omitempty"`
	Extractions []FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner `json:"extractions,omitempty"`
	Injections NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10Injections `json:"injections,omitempty"`
	AutoExtractionUrls []string `json:"auto_extraction_urls,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationAuthenticationPresetsInnerOneOf10 FrontendConfigurationAuthenticationPresetsInnerOneOf10

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf10 instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf10 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf10(type_ EnumBROWSERAGENT, users []FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner, loginUrl string) *FrontendConfigurationAuthenticationPresetsInnerOneOf10 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf10{}
	this.Type = type_
	this.Users = users
	this.LoginUrl = loginUrl
	return &this
}

// NewFrontendConfigurationAuthenticationPresetsInnerOneOf10WithDefaults instantiates a new FrontendConfigurationAuthenticationPresetsInnerOneOf10 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationAuthenticationPresetsInnerOneOf10WithDefaults() *FrontendConfigurationAuthenticationPresetsInnerOneOf10 {
	this := FrontendConfigurationAuthenticationPresetsInnerOneOf10{}
	return &this
}

// GetType returns the Type field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetType() EnumBROWSERAGENT {
	if o == nil {
		var ret EnumBROWSERAGENT
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetTypeOk() (*EnumBROWSERAGENT, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetType(v EnumBROWSERAGENT) {
	o.Type = v
}

// GetUsers returns the Users field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetUsers() []FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner {
	if o == nil {
		var ret []FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner
		return ret
	}

	return o.Users
}

// GetUsersOk returns a tuple with the Users field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetUsersOk() ([]FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner, bool) {
	if o == nil {
		return nil, false
	}
	return o.Users, true
}

// SetUsers sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetUsers(v []FrontendConfigurationAuthenticationPresetsInnerOneOf10UsersInner) {
	o.Users = v
}

// GetLoginUrl returns the LoginUrl field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoginUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LoginUrl
}

// GetLoginUrlOk returns a tuple with the LoginUrl field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoginUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LoginUrl, true
}

// SetLoginUrl sets field value
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetLoginUrl(v string) {
	o.LoginUrl = v
}

// GetLoggedInDetectorText returns the LoggedInDetectorText field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoggedInDetectorText() string {
	if o == nil || IsNil(o.LoggedInDetectorText) {
		var ret string
		return ret
	}
	return *o.LoggedInDetectorText
}

// GetLoggedInDetectorTextOk returns a tuple with the LoggedInDetectorText field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoggedInDetectorTextOk() (*string, bool) {
	if o == nil || IsNil(o.LoggedInDetectorText) {
		return nil, false
	}
	return o.LoggedInDetectorText, true
}

// HasLoggedInDetectorText returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) HasLoggedInDetectorText() bool {
	if o != nil && !IsNil(o.LoggedInDetectorText) {
		return true
	}

	return false
}

// SetLoggedInDetectorText gets a reference to the given string and assigns it to the LoggedInDetectorText field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetLoggedInDetectorText(v string) {
	o.LoggedInDetectorText = &v
}

// GetLoggedInDetectorTimeout returns the LoggedInDetectorTimeout field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoggedInDetectorTimeout() float32 {
	if o == nil || IsNil(o.LoggedInDetectorTimeout) {
		var ret float32
		return ret
	}
	return *o.LoggedInDetectorTimeout
}

// GetLoggedInDetectorTimeoutOk returns a tuple with the LoggedInDetectorTimeout field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetLoggedInDetectorTimeoutOk() (*float32, bool) {
	if o == nil || IsNil(o.LoggedInDetectorTimeout) {
		return nil, false
	}
	return o.LoggedInDetectorTimeout, true
}

// HasLoggedInDetectorTimeout returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) HasLoggedInDetectorTimeout() bool {
	if o != nil && !IsNil(o.LoggedInDetectorTimeout) {
		return true
	}

	return false
}

// SetLoggedInDetectorTimeout gets a reference to the given float32 and assigns it to the LoggedInDetectorTimeout field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetLoggedInDetectorTimeout(v float32) {
	o.LoggedInDetectorTimeout = &v
}

// GetExtractions returns the Extractions field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetExtractions() []FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner {
	if o == nil || IsNil(o.Extractions) {
		var ret []FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner
		return ret
	}
	return o.Extractions
}

// GetExtractionsOk returns a tuple with the Extractions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetExtractionsOk() ([]FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner, bool) {
	if o == nil || IsNil(o.Extractions) {
		return nil, false
	}
	return o.Extractions, true
}

// HasExtractions returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) HasExtractions() bool {
	if o != nil && !IsNil(o.Extractions) {
		return true
	}

	return false
}

// SetExtractions gets a reference to the given []FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner and assigns it to the Extractions field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetExtractions(v []FrontendConfigurationAuthenticationProceduresInnerOperationsInnerOneOf1ExtractionsAnyOfInner) {
	o.Extractions = v
}

// GetInjections returns the Injections field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetInjections() FrontendConfigurationAuthenticationPresetsInnerOneOf10Injections {
	if o == nil || IsNil(o.Injections.Get()) {
		var ret FrontendConfigurationAuthenticationPresetsInnerOneOf10Injections
		return ret
	}
	return *o.Injections.Get()
}

// GetInjectionsOk returns a tuple with the Injections field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetInjectionsOk() (*FrontendConfigurationAuthenticationPresetsInnerOneOf10Injections, bool) {
	if o == nil {
		return nil, false
	}
	return o.Injections.Get(), o.Injections.IsSet()
}

// HasInjections returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) HasInjections() bool {
	if o != nil && o.Injections.IsSet() {
		return true
	}

	return false
}

// SetInjections gets a reference to the given NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10Injections and assigns it to the Injections field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetInjections(v FrontendConfigurationAuthenticationPresetsInnerOneOf10Injections) {
	o.Injections.Set(&v)
}
// SetInjectionsNil sets the value for Injections to be an explicit nil
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetInjectionsNil() {
	o.Injections.Set(nil)
}

// UnsetInjections ensures that no value is present for Injections, not even an explicit nil
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) UnsetInjections() {
	o.Injections.Unset()
}

// GetAutoExtractionUrls returns the AutoExtractionUrls field value if set, zero value otherwise.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetAutoExtractionUrls() []string {
	if o == nil || IsNil(o.AutoExtractionUrls) {
		var ret []string
		return ret
	}
	return o.AutoExtractionUrls
}

// GetAutoExtractionUrlsOk returns a tuple with the AutoExtractionUrls field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) GetAutoExtractionUrlsOk() ([]string, bool) {
	if o == nil || IsNil(o.AutoExtractionUrls) {
		return nil, false
	}
	return o.AutoExtractionUrls, true
}

// HasAutoExtractionUrls returns a boolean if a field has been set.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) HasAutoExtractionUrls() bool {
	if o != nil && !IsNil(o.AutoExtractionUrls) {
		return true
	}

	return false
}

// SetAutoExtractionUrls gets a reference to the given []string and assigns it to the AutoExtractionUrls field.
func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) SetAutoExtractionUrls(v []string) {
	o.AutoExtractionUrls = v
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf10) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationAuthenticationPresetsInnerOneOf10) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["type"] = o.Type
	toSerialize["users"] = o.Users
	toSerialize["login_url"] = o.LoginUrl
	if !IsNil(o.LoggedInDetectorText) {
		toSerialize["logged_in_detector_text"] = o.LoggedInDetectorText
	}
	if !IsNil(o.LoggedInDetectorTimeout) {
		toSerialize["logged_in_detector_timeout"] = o.LoggedInDetectorTimeout
	}
	if !IsNil(o.Extractions) {
		toSerialize["extractions"] = o.Extractions
	}
	if o.Injections.IsSet() {
		toSerialize["injections"] = o.Injections.Get()
	}
	if !IsNil(o.AutoExtractionUrls) {
		toSerialize["auto_extraction_urls"] = o.AutoExtractionUrls
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationAuthenticationPresetsInnerOneOf10) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"type",
		"users",
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

	varFrontendConfigurationAuthenticationPresetsInnerOneOf10 := _FrontendConfigurationAuthenticationPresetsInnerOneOf10{}

	err = json.Unmarshal(data, &varFrontendConfigurationAuthenticationPresetsInnerOneOf10)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationAuthenticationPresetsInnerOneOf10(varFrontendConfigurationAuthenticationPresetsInnerOneOf10)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "type")
		delete(additionalProperties, "users")
		delete(additionalProperties, "login_url")
		delete(additionalProperties, "logged_in_detector_text")
		delete(additionalProperties, "logged_in_detector_timeout")
		delete(additionalProperties, "extractions")
		delete(additionalProperties, "injections")
		delete(additionalProperties, "auto_extraction_urls")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10 struct {
	value *FrontendConfigurationAuthenticationPresetsInnerOneOf10
	isSet bool
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) Get() *FrontendConfigurationAuthenticationPresetsInnerOneOf10 {
	return v.value
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) Set(val *FrontendConfigurationAuthenticationPresetsInnerOneOf10) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationAuthenticationPresetsInnerOneOf10(val *FrontendConfigurationAuthenticationPresetsInnerOneOf10) *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10 {
	return &NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10{value: val, isSet: true}
}

func (v NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationAuthenticationPresetsInnerOneOf10) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


