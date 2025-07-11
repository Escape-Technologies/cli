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

// checks if the ListIssues200ResponseInner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ListIssues200ResponseInner{}

// ListIssues200ResponseInner struct for ListIssues200ResponseInner
type ListIssues200ResponseInner struct {
	Id string `json:"id"`
	Severity Enum9c1e82c38fa16c4851aece69dc28da0b `json:"severity"`
	Name string `json:"name"`
	PlatformUrl string `json:"platformUrl"`
	Ignored bool `json:"ignored"`
	Type Enum15559151725d4598e75cbc5c6c9bd96e `json:"type"`
	Category Enum5cf07f4dc5d62ad66f92942c1b7ce23f `json:"category"`
	AdditionalProperties map[string]interface{}
}

type _ListIssues200ResponseInner ListIssues200ResponseInner

// NewListIssues200ResponseInner instantiates a new ListIssues200ResponseInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListIssues200ResponseInner(id string, severity Enum9c1e82c38fa16c4851aece69dc28da0b, name string, platformUrl string, ignored bool, type_ Enum15559151725d4598e75cbc5c6c9bd96e, category Enum5cf07f4dc5d62ad66f92942c1b7ce23f) *ListIssues200ResponseInner {
	this := ListIssues200ResponseInner{}
	this.Id = id
	this.Severity = severity
	this.Name = name
	this.PlatformUrl = platformUrl
	this.Ignored = ignored
	this.Type = type_
	this.Category = category
	return &this
}

// NewListIssues200ResponseInnerWithDefaults instantiates a new ListIssues200ResponseInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListIssues200ResponseInnerWithDefaults() *ListIssues200ResponseInner {
	this := ListIssues200ResponseInner{}
	return &this
}

// GetId returns the Id field value
func (o *ListIssues200ResponseInner) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *ListIssues200ResponseInner) SetId(v string) {
	o.Id = v
}

// GetSeverity returns the Severity field value
func (o *ListIssues200ResponseInner) GetSeverity() Enum9c1e82c38fa16c4851aece69dc28da0b {
	if o == nil {
		var ret Enum9c1e82c38fa16c4851aece69dc28da0b
		return ret
	}

	return o.Severity
}

// GetSeverityOk returns a tuple with the Severity field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetSeverityOk() (*Enum9c1e82c38fa16c4851aece69dc28da0b, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Severity, true
}

// SetSeverity sets field value
func (o *ListIssues200ResponseInner) SetSeverity(v Enum9c1e82c38fa16c4851aece69dc28da0b) {
	o.Severity = v
}

// GetName returns the Name field value
func (o *ListIssues200ResponseInner) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ListIssues200ResponseInner) SetName(v string) {
	o.Name = v
}

// GetPlatformUrl returns the PlatformUrl field value
func (o *ListIssues200ResponseInner) GetPlatformUrl() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PlatformUrl
}

// GetPlatformUrlOk returns a tuple with the PlatformUrl field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetPlatformUrlOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PlatformUrl, true
}

// SetPlatformUrl sets field value
func (o *ListIssues200ResponseInner) SetPlatformUrl(v string) {
	o.PlatformUrl = v
}

// GetIgnored returns the Ignored field value
func (o *ListIssues200ResponseInner) GetIgnored() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Ignored
}

// GetIgnoredOk returns a tuple with the Ignored field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetIgnoredOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Ignored, true
}

// SetIgnored sets field value
func (o *ListIssues200ResponseInner) SetIgnored(v bool) {
	o.Ignored = v
}

// GetType returns the Type field value
func (o *ListIssues200ResponseInner) GetType() Enum15559151725d4598e75cbc5c6c9bd96e {
	if o == nil {
		var ret Enum15559151725d4598e75cbc5c6c9bd96e
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetTypeOk() (*Enum15559151725d4598e75cbc5c6c9bd96e, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *ListIssues200ResponseInner) SetType(v Enum15559151725d4598e75cbc5c6c9bd96e) {
	o.Type = v
}

// GetCategory returns the Category field value
func (o *ListIssues200ResponseInner) GetCategory() Enum5cf07f4dc5d62ad66f92942c1b7ce23f {
	if o == nil {
		var ret Enum5cf07f4dc5d62ad66f92942c1b7ce23f
		return ret
	}

	return o.Category
}

// GetCategoryOk returns a tuple with the Category field value
// and a boolean to check if the value has been set.
func (o *ListIssues200ResponseInner) GetCategoryOk() (*Enum5cf07f4dc5d62ad66f92942c1b7ce23f, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Category, true
}

// SetCategory sets field value
func (o *ListIssues200ResponseInner) SetCategory(v Enum5cf07f4dc5d62ad66f92942c1b7ce23f) {
	o.Category = v
}

func (o ListIssues200ResponseInner) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ListIssues200ResponseInner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["severity"] = o.Severity
	toSerialize["name"] = o.Name
	toSerialize["platformUrl"] = o.PlatformUrl
	toSerialize["ignored"] = o.Ignored
	toSerialize["type"] = o.Type
	toSerialize["category"] = o.Category

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ListIssues200ResponseInner) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"severity",
		"name",
		"platformUrl",
		"ignored",
		"type",
		"category",
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

	varListIssues200ResponseInner := _ListIssues200ResponseInner{}

	err = json.Unmarshal(data, &varListIssues200ResponseInner)

	if err != nil {
		return err
	}

	*o = ListIssues200ResponseInner(varListIssues200ResponseInner)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "severity")
		delete(additionalProperties, "name")
		delete(additionalProperties, "platformUrl")
		delete(additionalProperties, "ignored")
		delete(additionalProperties, "type")
		delete(additionalProperties, "category")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableListIssues200ResponseInner struct {
	value *ListIssues200ResponseInner
	isSet bool
}

func (v NullableListIssues200ResponseInner) Get() *ListIssues200ResponseInner {
	return v.value
}

func (v *NullableListIssues200ResponseInner) Set(val *ListIssues200ResponseInner) {
	v.value = val
	v.isSet = true
}

func (v NullableListIssues200ResponseInner) IsSet() bool {
	return v.isSet
}

func (v *NullableListIssues200ResponseInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListIssues200ResponseInner(val *ListIssues200ResponseInner) *NullableListIssues200ResponseInner {
	return &NullableListIssues200ResponseInner{value: val, isSet: true}
}

func (v NullableListIssues200ResponseInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListIssues200ResponseInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


