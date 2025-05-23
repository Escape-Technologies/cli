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

// checks if the CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials{}

// CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials struct for CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials
type CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Headers []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner `json:"headers,omitempty"`
	Cookies []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner `json:"cookies,omitempty"`
	QueryParameters []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner `json:"queryParameters,omitempty"`
	Body interface{} `json:"body,omitempty"`
	LocalStorage map[string]map[string]string `json:"local_storage,omitempty"`
	SessionStorage map[string]map[string]string `json:"session_storage,omitempty"`
	Actions []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner `json:"actions,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials

// NewCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials instantiates a new CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials() *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials {
	this := CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials{}
	return &this
}

// NewCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentialsWithDefaults instantiates a new CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentialsWithDefaults() *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials {
	this := CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials{}
	return &this
}

// GetUsername returns the Username field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetUsername() string {
	if o == nil || IsNil(o.Username) {
		var ret string
		return ret
	}
	return *o.Username
}

// GetUsernameOk returns a tuple with the Username field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetUsernameOk() (*string, bool) {
	if o == nil || IsNil(o.Username) {
		return nil, false
	}
	return o.Username, true
}

// HasUsername returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasUsername() bool {
	if o != nil && !IsNil(o.Username) {
		return true
	}

	return false
}

// SetUsername gets a reference to the given string and assigns it to the Username field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetUsername(v string) {
	o.Username = &v
}

// GetPassword returns the Password field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetPassword() string {
	if o == nil || IsNil(o.Password) {
		var ret string
		return ret
	}
	return *o.Password
}

// GetPasswordOk returns a tuple with the Password field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetPasswordOk() (*string, bool) {
	if o == nil || IsNil(o.Password) {
		return nil, false
	}
	return o.Password, true
}

// HasPassword returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasPassword() bool {
	if o != nil && !IsNil(o.Password) {
		return true
	}

	return false
}

// SetPassword gets a reference to the given string and assigns it to the Password field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetPassword(v string) {
	o.Password = &v
}

// GetHeaders returns the Headers field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetHeaders() []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner {
	if o == nil || IsNil(o.Headers) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner
		return ret
	}
	return o.Headers
}

// GetHeadersOk returns a tuple with the Headers field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetHeadersOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner, bool) {
	if o == nil || IsNil(o.Headers) {
		return nil, false
	}
	return o.Headers, true
}

// HasHeaders returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasHeaders() bool {
	if o != nil && !IsNil(o.Headers) {
		return true
	}

	return false
}

// SetHeaders gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner and assigns it to the Headers field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetHeaders(v []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner) {
	o.Headers = v
}

// GetCookies returns the Cookies field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetCookies() []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner {
	if o == nil || IsNil(o.Cookies) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner
		return ret
	}
	return o.Cookies
}

// GetCookiesOk returns a tuple with the Cookies field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetCookiesOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner, bool) {
	if o == nil || IsNil(o.Cookies) {
		return nil, false
	}
	return o.Cookies, true
}

// HasCookies returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasCookies() bool {
	if o != nil && !IsNil(o.Cookies) {
		return true
	}

	return false
}

// SetCookies gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner and assigns it to the Cookies field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetCookies(v []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersCookiesInner) {
	o.Cookies = v
}

// GetQueryParameters returns the QueryParameters field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetQueryParameters() []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner {
	if o == nil || IsNil(o.QueryParameters) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner
		return ret
	}
	return o.QueryParameters
}

// GetQueryParametersOk returns a tuple with the QueryParameters field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetQueryParametersOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner, bool) {
	if o == nil || IsNil(o.QueryParameters) {
		return nil, false
	}
	return o.QueryParameters, true
}

// HasQueryParameters returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasQueryParameters() bool {
	if o != nil && !IsNil(o.QueryParameters) {
		return true
	}

	return false
}

// SetQueryParameters gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner and assigns it to the QueryParameters field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetQueryParameters(v []CreateApplicationRequestAnyOfConfigurationAuthenticationProceduresInnerOperationsInnerOneOfParametersHeadersInner) {
	o.QueryParameters = v
}

// GetBody returns the Body field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetBody() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Body
}

// GetBodyOk returns a tuple with the Body field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetBodyOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Body) {
		return nil, false
	}
	return &o.Body, true
}

// HasBody returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasBody() bool {
	if o != nil && !IsNil(o.Body) {
		return true
	}

	return false
}

// SetBody gets a reference to the given interface{} and assigns it to the Body field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetBody(v interface{}) {
	o.Body = v
}

// GetLocalStorage returns the LocalStorage field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetLocalStorage() map[string]map[string]string {
	if o == nil || IsNil(o.LocalStorage) {
		var ret map[string]map[string]string
		return ret
	}
	return o.LocalStorage
}

// GetLocalStorageOk returns a tuple with the LocalStorage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetLocalStorageOk() (map[string]map[string]string, bool) {
	if o == nil || IsNil(o.LocalStorage) {
		return map[string]map[string]string{}, false
	}
	return o.LocalStorage, true
}

// HasLocalStorage returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasLocalStorage() bool {
	if o != nil && !IsNil(o.LocalStorage) {
		return true
	}

	return false
}

// SetLocalStorage gets a reference to the given map[string]map[string]string and assigns it to the LocalStorage field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetLocalStorage(v map[string]map[string]string) {
	o.LocalStorage = v
}

// GetSessionStorage returns the SessionStorage field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetSessionStorage() map[string]map[string]string {
	if o == nil || IsNil(o.SessionStorage) {
		var ret map[string]map[string]string
		return ret
	}
	return o.SessionStorage
}

// GetSessionStorageOk returns a tuple with the SessionStorage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetSessionStorageOk() (map[string]map[string]string, bool) {
	if o == nil || IsNil(o.SessionStorage) {
		return map[string]map[string]string{}, false
	}
	return o.SessionStorage, true
}

// HasSessionStorage returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasSessionStorage() bool {
	if o != nil && !IsNil(o.SessionStorage) {
		return true
	}

	return false
}

// SetSessionStorage gets a reference to the given map[string]map[string]string and assigns it to the SessionStorage field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetSessionStorage(v map[string]map[string]string) {
	o.SessionStorage = v
}

// GetActions returns the Actions field value if set, zero value otherwise.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetActions() []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner {
	if o == nil || IsNil(o.Actions) {
		var ret []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner
		return ret
	}
	return o.Actions
}

// GetActionsOk returns a tuple with the Actions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) GetActionsOk() ([]CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner, bool) {
	if o == nil || IsNil(o.Actions) {
		return nil, false
	}
	return o.Actions, true
}

// HasActions returns a boolean if a field has been set.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) HasActions() bool {
	if o != nil && !IsNil(o.Actions) {
		return true
	}

	return false
}

// SetActions gets a reference to the given []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner and assigns it to the Actions field.
func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) SetActions(v []CreateApplicationRequestAnyOfConfigurationAuthenticationPresetsInnerOneOf11UsersInnerActionsInner) {
	o.Actions = v
}

func (o CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Username) {
		toSerialize["username"] = o.Username
	}
	if !IsNil(o.Password) {
		toSerialize["password"] = o.Password
	}
	if !IsNil(o.Headers) {
		toSerialize["headers"] = o.Headers
	}
	if !IsNil(o.Cookies) {
		toSerialize["cookies"] = o.Cookies
	}
	if !IsNil(o.QueryParameters) {
		toSerialize["queryParameters"] = o.QueryParameters
	}
	if o.Body != nil {
		toSerialize["body"] = o.Body
	}
	if !IsNil(o.LocalStorage) {
		toSerialize["local_storage"] = o.LocalStorage
	}
	if !IsNil(o.SessionStorage) {
		toSerialize["session_storage"] = o.SessionStorage
	}
	if !IsNil(o.Actions) {
		toSerialize["actions"] = o.Actions
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) UnmarshalJSON(data []byte) (err error) {
	varCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials := _CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials{}

	err = json.Unmarshal(data, &varCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials)

	if err != nil {
		return err
	}

	*o = CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials(varCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "username")
		delete(additionalProperties, "password")
		delete(additionalProperties, "headers")
		delete(additionalProperties, "cookies")
		delete(additionalProperties, "queryParameters")
		delete(additionalProperties, "body")
		delete(additionalProperties, "local_storage")
		delete(additionalProperties, "session_storage")
		delete(additionalProperties, "actions")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials struct {
	value *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials
	isSet bool
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) Get() *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials {
	return v.value
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) Set(val *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials(val *CreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials {
	return &NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials{value: val, isSet: true}
}

func (v NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateApplicationRequestAnyOfConfigurationAuthenticationUsersInnerCredentials) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


