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

// checks if the StartScanRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &StartScanRequest{}

// StartScanRequest struct for StartScanRequest
type StartScanRequest struct {
	// The ID of the application to scan
	ApplicationId string `json:"applicationId"`
	CommitHash *string `json:"commitHash,omitempty"`
	CommitLink *string `json:"commitLink,omitempty"`
	CommitBranch *string `json:"commitBranch,omitempty"`
	CommitAuthor *string `json:"commitAuthor,omitempty"`
	CommitAuthorProfilePictureLink *string `json:"commitAuthorProfilePictureLink,omitempty"`
	ConfigurationOverride *FrontendConfiguration `json:"configurationOverride,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _StartScanRequest StartScanRequest

// NewStartScanRequest instantiates a new StartScanRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewStartScanRequest(applicationId string) *StartScanRequest {
	this := StartScanRequest{}
	this.ApplicationId = applicationId
	return &this
}

// NewStartScanRequestWithDefaults instantiates a new StartScanRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewStartScanRequestWithDefaults() *StartScanRequest {
	this := StartScanRequest{}
	return &this
}

// GetApplicationId returns the ApplicationId field value
func (o *StartScanRequest) GetApplicationId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ApplicationId
}

// GetApplicationIdOk returns a tuple with the ApplicationId field value
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetApplicationIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ApplicationId, true
}

// SetApplicationId sets field value
func (o *StartScanRequest) SetApplicationId(v string) {
	o.ApplicationId = v
}

// GetCommitHash returns the CommitHash field value if set, zero value otherwise.
func (o *StartScanRequest) GetCommitHash() string {
	if o == nil || IsNil(o.CommitHash) {
		var ret string
		return ret
	}
	return *o.CommitHash
}

// GetCommitHashOk returns a tuple with the CommitHash field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetCommitHashOk() (*string, bool) {
	if o == nil || IsNil(o.CommitHash) {
		return nil, false
	}
	return o.CommitHash, true
}

// HasCommitHash returns a boolean if a field has been set.
func (o *StartScanRequest) HasCommitHash() bool {
	if o != nil && !IsNil(o.CommitHash) {
		return true
	}

	return false
}

// SetCommitHash gets a reference to the given string and assigns it to the CommitHash field.
func (o *StartScanRequest) SetCommitHash(v string) {
	o.CommitHash = &v
}

// GetCommitLink returns the CommitLink field value if set, zero value otherwise.
func (o *StartScanRequest) GetCommitLink() string {
	if o == nil || IsNil(o.CommitLink) {
		var ret string
		return ret
	}
	return *o.CommitLink
}

// GetCommitLinkOk returns a tuple with the CommitLink field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetCommitLinkOk() (*string, bool) {
	if o == nil || IsNil(o.CommitLink) {
		return nil, false
	}
	return o.CommitLink, true
}

// HasCommitLink returns a boolean if a field has been set.
func (o *StartScanRequest) HasCommitLink() bool {
	if o != nil && !IsNil(o.CommitLink) {
		return true
	}

	return false
}

// SetCommitLink gets a reference to the given string and assigns it to the CommitLink field.
func (o *StartScanRequest) SetCommitLink(v string) {
	o.CommitLink = &v
}

// GetCommitBranch returns the CommitBranch field value if set, zero value otherwise.
func (o *StartScanRequest) GetCommitBranch() string {
	if o == nil || IsNil(o.CommitBranch) {
		var ret string
		return ret
	}
	return *o.CommitBranch
}

// GetCommitBranchOk returns a tuple with the CommitBranch field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetCommitBranchOk() (*string, bool) {
	if o == nil || IsNil(o.CommitBranch) {
		return nil, false
	}
	return o.CommitBranch, true
}

// HasCommitBranch returns a boolean if a field has been set.
func (o *StartScanRequest) HasCommitBranch() bool {
	if o != nil && !IsNil(o.CommitBranch) {
		return true
	}

	return false
}

// SetCommitBranch gets a reference to the given string and assigns it to the CommitBranch field.
func (o *StartScanRequest) SetCommitBranch(v string) {
	o.CommitBranch = &v
}

// GetCommitAuthor returns the CommitAuthor field value if set, zero value otherwise.
func (o *StartScanRequest) GetCommitAuthor() string {
	if o == nil || IsNil(o.CommitAuthor) {
		var ret string
		return ret
	}
	return *o.CommitAuthor
}

// GetCommitAuthorOk returns a tuple with the CommitAuthor field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetCommitAuthorOk() (*string, bool) {
	if o == nil || IsNil(o.CommitAuthor) {
		return nil, false
	}
	return o.CommitAuthor, true
}

// HasCommitAuthor returns a boolean if a field has been set.
func (o *StartScanRequest) HasCommitAuthor() bool {
	if o != nil && !IsNil(o.CommitAuthor) {
		return true
	}

	return false
}

// SetCommitAuthor gets a reference to the given string and assigns it to the CommitAuthor field.
func (o *StartScanRequest) SetCommitAuthor(v string) {
	o.CommitAuthor = &v
}

// GetCommitAuthorProfilePictureLink returns the CommitAuthorProfilePictureLink field value if set, zero value otherwise.
func (o *StartScanRequest) GetCommitAuthorProfilePictureLink() string {
	if o == nil || IsNil(o.CommitAuthorProfilePictureLink) {
		var ret string
		return ret
	}
	return *o.CommitAuthorProfilePictureLink
}

// GetCommitAuthorProfilePictureLinkOk returns a tuple with the CommitAuthorProfilePictureLink field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetCommitAuthorProfilePictureLinkOk() (*string, bool) {
	if o == nil || IsNil(o.CommitAuthorProfilePictureLink) {
		return nil, false
	}
	return o.CommitAuthorProfilePictureLink, true
}

// HasCommitAuthorProfilePictureLink returns a boolean if a field has been set.
func (o *StartScanRequest) HasCommitAuthorProfilePictureLink() bool {
	if o != nil && !IsNil(o.CommitAuthorProfilePictureLink) {
		return true
	}

	return false
}

// SetCommitAuthorProfilePictureLink gets a reference to the given string and assigns it to the CommitAuthorProfilePictureLink field.
func (o *StartScanRequest) SetCommitAuthorProfilePictureLink(v string) {
	o.CommitAuthorProfilePictureLink = &v
}

// GetConfigurationOverride returns the ConfigurationOverride field value if set, zero value otherwise.
func (o *StartScanRequest) GetConfigurationOverride() FrontendConfiguration {
	if o == nil || IsNil(o.ConfigurationOverride) {
		var ret FrontendConfiguration
		return ret
	}
	return *o.ConfigurationOverride
}

// GetConfigurationOverrideOk returns a tuple with the ConfigurationOverride field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StartScanRequest) GetConfigurationOverrideOk() (*FrontendConfiguration, bool) {
	if o == nil || IsNil(o.ConfigurationOverride) {
		return nil, false
	}
	return o.ConfigurationOverride, true
}

// HasConfigurationOverride returns a boolean if a field has been set.
func (o *StartScanRequest) HasConfigurationOverride() bool {
	if o != nil && !IsNil(o.ConfigurationOverride) {
		return true
	}

	return false
}

// SetConfigurationOverride gets a reference to the given FrontendConfiguration and assigns it to the ConfigurationOverride field.
func (o *StartScanRequest) SetConfigurationOverride(v FrontendConfiguration) {
	o.ConfigurationOverride = &v
}

func (o StartScanRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o StartScanRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["applicationId"] = o.ApplicationId
	if !IsNil(o.CommitHash) {
		toSerialize["commitHash"] = o.CommitHash
	}
	if !IsNil(o.CommitLink) {
		toSerialize["commitLink"] = o.CommitLink
	}
	if !IsNil(o.CommitBranch) {
		toSerialize["commitBranch"] = o.CommitBranch
	}
	if !IsNil(o.CommitAuthor) {
		toSerialize["commitAuthor"] = o.CommitAuthor
	}
	if !IsNil(o.CommitAuthorProfilePictureLink) {
		toSerialize["commitAuthorProfilePictureLink"] = o.CommitAuthorProfilePictureLink
	}
	if !IsNil(o.ConfigurationOverride) {
		toSerialize["configurationOverride"] = o.ConfigurationOverride
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *StartScanRequest) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"applicationId",
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

	varStartScanRequest := _StartScanRequest{}

	err = json.Unmarshal(data, &varStartScanRequest)

	if err != nil {
		return err
	}

	*o = StartScanRequest(varStartScanRequest)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "applicationId")
		delete(additionalProperties, "commitHash")
		delete(additionalProperties, "commitLink")
		delete(additionalProperties, "commitBranch")
		delete(additionalProperties, "commitAuthor")
		delete(additionalProperties, "commitAuthorProfilePictureLink")
		delete(additionalProperties, "configurationOverride")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableStartScanRequest struct {
	value *StartScanRequest
	isSet bool
}

func (v NullableStartScanRequest) Get() *StartScanRequest {
	return v.value
}

func (v *NullableStartScanRequest) Set(val *StartScanRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableStartScanRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableStartScanRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableStartScanRequest(val *StartScanRequest) *NullableStartScanRequest {
	return &NullableStartScanRequest{value: val, isSet: true}
}

func (v NullableStartScanRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableStartScanRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


