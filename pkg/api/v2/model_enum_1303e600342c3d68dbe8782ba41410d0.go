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

// Enum1303e600342c3d68dbe8782ba41410d0 the model 'Enum1303e600342c3d68dbe8782ba41410d0'
type Enum1303e600342c3d68dbe8782ba41410d0 string

// List of Enum_1303e600342c3d68dbe8782ba41410d0
const (
	ENUM1303E600342C3D68DBE8782BA41410D0__16 Enum1303e600342c3d68dbe8782ba41410d0 = "16"
	ENUM1303E600342C3D68DBE8782BA41410D0__20 Enum1303e600342c3d68dbe8782ba41410d0 = "20"
	ENUM1303E600342C3D68DBE8782BA41410D0__22 Enum1303e600342c3d68dbe8782ba41410d0 = "22"
	ENUM1303E600342C3D68DBE8782BA41410D0__78 Enum1303e600342c3d68dbe8782ba41410d0 = "78"
	ENUM1303E600342C3D68DBE8782BA41410D0__79 Enum1303e600342c3d68dbe8782ba41410d0 = "79"
	ENUM1303E600342C3D68DBE8782BA41410D0__89 Enum1303e600342c3d68dbe8782ba41410d0 = "89"
	ENUM1303E600342C3D68DBE8782BA41410D0__93 Enum1303e600342c3d68dbe8782ba41410d0 = "93"
	ENUM1303E600342C3D68DBE8782BA41410D0__94 Enum1303e600342c3d68dbe8782ba41410d0 = "94"
	ENUM1303E600342C3D68DBE8782BA41410D0__116 Enum1303e600342c3d68dbe8782ba41410d0 = "116"
	ENUM1303E600342C3D68DBE8782BA41410D0__119 Enum1303e600342c3d68dbe8782ba41410d0 = "119"
	ENUM1303E600342C3D68DBE8782BA41410D0__200 Enum1303e600342c3d68dbe8782ba41410d0 = "200"
	ENUM1303E600342C3D68DBE8782BA41410D0__209 Enum1303e600342c3d68dbe8782ba41410d0 = "209"
	ENUM1303E600342C3D68DBE8782BA41410D0__215 Enum1303e600342c3d68dbe8782ba41410d0 = "215"
	ENUM1303E600342C3D68DBE8782BA41410D0__264 Enum1303e600342c3d68dbe8782ba41410d0 = "264"
	ENUM1303E600342C3D68DBE8782BA41410D0__284 Enum1303e600342c3d68dbe8782ba41410d0 = "284"
	ENUM1303E600342C3D68DBE8782BA41410D0__285 Enum1303e600342c3d68dbe8782ba41410d0 = "285"
	ENUM1303E600342C3D68DBE8782BA41410D0__287 Enum1303e600342c3d68dbe8782ba41410d0 = "287"
	ENUM1303E600342C3D68DBE8782BA41410D0__295 Enum1303e600342c3d68dbe8782ba41410d0 = "295"
	ENUM1303E600342C3D68DBE8782BA41410D0__306 Enum1303e600342c3d68dbe8782ba41410d0 = "306"
	ENUM1303E600342C3D68DBE8782BA41410D0__307 Enum1303e600342c3d68dbe8782ba41410d0 = "307"
	ENUM1303E600342C3D68DBE8782BA41410D0__319 Enum1303e600342c3d68dbe8782ba41410d0 = "319"
	ENUM1303E600342C3D68DBE8782BA41410D0__326 Enum1303e600342c3d68dbe8782ba41410d0 = "326"
	ENUM1303E600342C3D68DBE8782BA41410D0__346 Enum1303e600342c3d68dbe8782ba41410d0 = "346"
	ENUM1303E600342C3D68DBE8782BA41410D0__347 Enum1303e600342c3d68dbe8782ba41410d0 = "347"
	ENUM1303E600342C3D68DBE8782BA41410D0__352 Enum1303e600342c3d68dbe8782ba41410d0 = "352"
	ENUM1303E600342C3D68DBE8782BA41410D0__354 Enum1303e600342c3d68dbe8782ba41410d0 = "354"
	ENUM1303E600342C3D68DBE8782BA41410D0__400 Enum1303e600342c3d68dbe8782ba41410d0 = "400"
	ENUM1303E600342C3D68DBE8782BA41410D0__444 Enum1303e600342c3d68dbe8782ba41410d0 = "444"
	ENUM1303E600342C3D68DBE8782BA41410D0__453 Enum1303e600342c3d68dbe8782ba41410d0 = "453"
	ENUM1303E600342C3D68DBE8782BA41410D0__489 Enum1303e600342c3d68dbe8782ba41410d0 = "489"
	ENUM1303E600342C3D68DBE8782BA41410D0__502 Enum1303e600342c3d68dbe8782ba41410d0 = "502"
	ENUM1303E600342C3D68DBE8782BA41410D0__523 Enum1303e600342c3d68dbe8782ba41410d0 = "523"
	ENUM1303E600342C3D68DBE8782BA41410D0__524 Enum1303e600342c3d68dbe8782ba41410d0 = "524"
	ENUM1303E600342C3D68DBE8782BA41410D0__548 Enum1303e600342c3d68dbe8782ba41410d0 = "548"
	ENUM1303E600342C3D68DBE8782BA41410D0__551 Enum1303e600342c3d68dbe8782ba41410d0 = "551"
	ENUM1303E600342C3D68DBE8782BA41410D0__573 Enum1303e600342c3d68dbe8782ba41410d0 = "573"
	ENUM1303E600342C3D68DBE8782BA41410D0__601 Enum1303e600342c3d68dbe8782ba41410d0 = "601"
	ENUM1303E600342C3D68DBE8782BA41410D0__611 Enum1303e600342c3d68dbe8782ba41410d0 = "611"
	ENUM1303E600342C3D68DBE8782BA41410D0__614 Enum1303e600342c3d68dbe8782ba41410d0 = "614"
	ENUM1303E600342C3D68DBE8782BA41410D0__676 Enum1303e600342c3d68dbe8782ba41410d0 = "676"
	ENUM1303E600342C3D68DBE8782BA41410D0__704 Enum1303e600342c3d68dbe8782ba41410d0 = "704"
	ENUM1303E600342C3D68DBE8782BA41410D0__710 Enum1303e600342c3d68dbe8782ba41410d0 = "710"
	ENUM1303E600342C3D68DBE8782BA41410D0__730 Enum1303e600342c3d68dbe8782ba41410d0 = "730"
	ENUM1303E600342C3D68DBE8782BA41410D0__732 Enum1303e600342c3d68dbe8782ba41410d0 = "732"
	ENUM1303E600342C3D68DBE8782BA41410D0__758 Enum1303e600342c3d68dbe8782ba41410d0 = "758"
	ENUM1303E600342C3D68DBE8782BA41410D0__770 Enum1303e600342c3d68dbe8782ba41410d0 = "770"
	ENUM1303E600342C3D68DBE8782BA41410D0__829 Enum1303e600342c3d68dbe8782ba41410d0 = "829"
	ENUM1303E600342C3D68DBE8782BA41410D0__862 Enum1303e600342c3d68dbe8782ba41410d0 = "862"
	ENUM1303E600342C3D68DBE8782BA41410D0__863 Enum1303e600342c3d68dbe8782ba41410d0 = "863"
	ENUM1303E600342C3D68DBE8782BA41410D0__915 Enum1303e600342c3d68dbe8782ba41410d0 = "915"
	ENUM1303E600342C3D68DBE8782BA41410D0__918 Enum1303e600342c3d68dbe8782ba41410d0 = "918"
	ENUM1303E600342C3D68DBE8782BA41410D0__942 Enum1303e600342c3d68dbe8782ba41410d0 = "942"
	ENUM1303E600342C3D68DBE8782BA41410D0__943 Enum1303e600342c3d68dbe8782ba41410d0 = "943"
	ENUM1303E600342C3D68DBE8782BA41410D0__1029 Enum1303e600342c3d68dbe8782ba41410d0 = "1029"
	ENUM1303E600342C3D68DBE8782BA41410D0__1195 Enum1303e600342c3d68dbe8782ba41410d0 = "1195"
)

// All allowed values of Enum1303e600342c3d68dbe8782ba41410d0 enum
var AllowedEnum1303e600342c3d68dbe8782ba41410d0EnumValues = []Enum1303e600342c3d68dbe8782ba41410d0{
	"16",
	"20",
	"22",
	"78",
	"79",
	"89",
	"93",
	"94",
	"116",
	"119",
	"200",
	"209",
	"215",
	"264",
	"284",
	"285",
	"287",
	"295",
	"306",
	"307",
	"319",
	"326",
	"346",
	"347",
	"352",
	"354",
	"400",
	"444",
	"453",
	"489",
	"502",
	"523",
	"524",
	"548",
	"551",
	"573",
	"601",
	"611",
	"614",
	"676",
	"704",
	"710",
	"730",
	"732",
	"758",
	"770",
	"829",
	"862",
	"863",
	"915",
	"918",
	"942",
	"943",
	"1029",
	"1195",
}

func (v *Enum1303e600342c3d68dbe8782ba41410d0) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := Enum1303e600342c3d68dbe8782ba41410d0(value)
	for _, existing := range AllowedEnum1303e600342c3d68dbe8782ba41410d0EnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid Enum1303e600342c3d68dbe8782ba41410d0", value)
}

// NewEnum1303e600342c3d68dbe8782ba41410d0FromValue returns a pointer to a valid Enum1303e600342c3d68dbe8782ba41410d0
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnum1303e600342c3d68dbe8782ba41410d0FromValue(v string) (*Enum1303e600342c3d68dbe8782ba41410d0, error) {
	ev := Enum1303e600342c3d68dbe8782ba41410d0(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for Enum1303e600342c3d68dbe8782ba41410d0: valid values are %v", v, AllowedEnum1303e600342c3d68dbe8782ba41410d0EnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v Enum1303e600342c3d68dbe8782ba41410d0) IsValid() bool {
	for _, existing := range AllowedEnum1303e600342c3d68dbe8782ba41410d0EnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to Enum_1303e600342c3d68dbe8782ba41410d0 value
func (v Enum1303e600342c3d68dbe8782ba41410d0) Ptr() *Enum1303e600342c3d68dbe8782ba41410d0 {
	return &v
}

type NullableEnum1303e600342c3d68dbe8782ba41410d0 struct {
	value *Enum1303e600342c3d68dbe8782ba41410d0
	isSet bool
}

func (v NullableEnum1303e600342c3d68dbe8782ba41410d0) Get() *Enum1303e600342c3d68dbe8782ba41410d0 {
	return v.value
}

func (v *NullableEnum1303e600342c3d68dbe8782ba41410d0) Set(val *Enum1303e600342c3d68dbe8782ba41410d0) {
	v.value = val
	v.isSet = true
}

func (v NullableEnum1303e600342c3d68dbe8782ba41410d0) IsSet() bool {
	return v.isSet
}

func (v *NullableEnum1303e600342c3d68dbe8782ba41410d0) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnum1303e600342c3d68dbe8782ba41410d0(val *Enum1303e600342c3d68dbe8782ba41410d0) *NullableEnum1303e600342c3d68dbe8782ba41410d0 {
	return &NullableEnum1303e600342c3d68dbe8782ba41410d0{value: val, isSet: true}
}

func (v NullableEnum1303e600342c3d68dbe8782ba41410d0) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnum1303e600342c3d68dbe8782ba41410d0) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

