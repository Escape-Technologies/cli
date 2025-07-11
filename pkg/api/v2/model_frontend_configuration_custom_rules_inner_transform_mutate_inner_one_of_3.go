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

// checks if the FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3{}

// FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 struct for FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
type FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 struct {
	Value *string `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
	RegexReplace *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace `json:"regex_replace,omitempty"`
	Key EnumREQUESTHEADERS `json:"key"`
	Name string `json:"name"`
	Delete *bool `json:"delete,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3(key EnumREQUESTHEADERS, name string) *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3{}
	this.Key = key
	this.Name = name
	return &this
}

// NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3WithDefaults instantiates a new FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3WithDefaults() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 {
	this := FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3{}
	return &this
}

// GetValue returns the Value field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetValue() string {
	if o == nil || IsNil(o.Value) {
		var ret string
		return ret
	}
	return *o.Value
}

// GetValueOk returns a tuple with the Value field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetValueOk() (*string, bool) {
	if o == nil || IsNil(o.Value) {
		return nil, false
	}
	return o.Value, true
}

// HasValue returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) HasValue() bool {
	if o != nil && !IsNil(o.Value) {
		return true
	}

	return false
}

// SetValue gets a reference to the given string and assigns it to the Value field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetValue(v string) {
	o.Value = &v
}

// GetValues returns the Values field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetValues() []string {
	if o == nil || IsNil(o.Values) {
		var ret []string
		return ret
	}
	return o.Values
}

// GetValuesOk returns a tuple with the Values field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetValuesOk() ([]string, bool) {
	if o == nil || IsNil(o.Values) {
		return nil, false
	}
	return o.Values, true
}

// HasValues returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) HasValues() bool {
	if o != nil && !IsNil(o.Values) {
		return true
	}

	return false
}

// SetValues gets a reference to the given []string and assigns it to the Values field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetValues(v []string) {
	o.Values = v
}

// GetRegexReplace returns the RegexReplace field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetRegexReplace() FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace {
	if o == nil || IsNil(o.RegexReplace) {
		var ret FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace
		return ret
	}
	return *o.RegexReplace
}

// GetRegexReplaceOk returns a tuple with the RegexReplace field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetRegexReplaceOk() (*FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace, bool) {
	if o == nil || IsNil(o.RegexReplace) {
		return nil, false
	}
	return o.RegexReplace, true
}

// HasRegexReplace returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) HasRegexReplace() bool {
	if o != nil && !IsNil(o.RegexReplace) {
		return true
	}

	return false
}

// SetRegexReplace gets a reference to the given FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace and assigns it to the RegexReplace field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetRegexReplace(v FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOfRegexReplace) {
	o.RegexReplace = &v
}

// GetKey returns the Key field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetKey() EnumREQUESTHEADERS {
	if o == nil {
		var ret EnumREQUESTHEADERS
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetKeyOk() (*EnumREQUESTHEADERS, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetKey(v EnumREQUESTHEADERS) {
	o.Key = v
}

// GetName returns the Name field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetName(v string) {
	o.Name = v
}

// GetDelete returns the Delete field value if set, zero value otherwise.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetDelete() bool {
	if o == nil || IsNil(o.Delete) {
		var ret bool
		return ret
	}
	return *o.Delete
}

// GetDeleteOk returns a tuple with the Delete field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) GetDeleteOk() (*bool, bool) {
	if o == nil || IsNil(o.Delete) {
		return nil, false
	}
	return o.Delete, true
}

// HasDelete returns a boolean if a field has been set.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) HasDelete() bool {
	if o != nil && !IsNil(o.Delete) {
		return true
	}

	return false
}

// SetDelete gets a reference to the given bool and assigns it to the Delete field.
func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) SetDelete(v bool) {
	o.Delete = &v
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Value) {
		toSerialize["value"] = o.Value
	}
	if !IsNil(o.Values) {
		toSerialize["values"] = o.Values
	}
	if !IsNil(o.RegexReplace) {
		toSerialize["regex_replace"] = o.RegexReplace
	}
	toSerialize["key"] = o.Key
	toSerialize["name"] = o.Name
	if !IsNil(o.Delete) {
		toSerialize["delete"] = o.Delete
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"key",
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

	varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 := _FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3{}

	err = json.Unmarshal(data, &varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3)

	if err != nil {
		return err
	}

	*o = FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3(varFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "value")
		delete(additionalProperties, "values")
		delete(additionalProperties, "regex_replace")
		delete(additionalProperties, "key")
		delete(additionalProperties, "name")
		delete(additionalProperties, "delete")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 struct {
	value *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3
	isSet bool
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) Get() *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 {
	return v.value
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) Set(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) {
	v.value = val
	v.isSet = true
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) IsSet() bool {
	return v.isSet
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3(val *FrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3 {
	return &NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3{value: val, isSet: true}
}

func (v NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFrontendConfigurationCustomRulesInnerTransformMutateInnerOneOf3) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


