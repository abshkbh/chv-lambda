/*
Cloud Hypervisor API

Local HTTP based API for managing and inspecting a cloud-hypervisor virtual machine.

API version: 0.3.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the SgxEpcConfig type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SgxEpcConfig{}

// SgxEpcConfig struct for SgxEpcConfig
type SgxEpcConfig struct {
	Id string `json:"id"`
	Size int64 `json:"size"`
	Prefault *bool `json:"prefault,omitempty"`
}

type _SgxEpcConfig SgxEpcConfig

// NewSgxEpcConfig instantiates a new SgxEpcConfig object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSgxEpcConfig(id string, size int64) *SgxEpcConfig {
	this := SgxEpcConfig{}
	this.Id = id
	this.Size = size
	var prefault bool = false
	this.Prefault = &prefault
	return &this
}

// NewSgxEpcConfigWithDefaults instantiates a new SgxEpcConfig object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSgxEpcConfigWithDefaults() *SgxEpcConfig {
	this := SgxEpcConfig{}
	var prefault bool = false
	this.Prefault = &prefault
	return &this
}

// GetId returns the Id field value
func (o *SgxEpcConfig) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *SgxEpcConfig) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *SgxEpcConfig) SetId(v string) {
	o.Id = v
}

// GetSize returns the Size field value
func (o *SgxEpcConfig) GetSize() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.Size
}

// GetSizeOk returns a tuple with the Size field value
// and a boolean to check if the value has been set.
func (o *SgxEpcConfig) GetSizeOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Size, true
}

// SetSize sets field value
func (o *SgxEpcConfig) SetSize(v int64) {
	o.Size = v
}

// GetPrefault returns the Prefault field value if set, zero value otherwise.
func (o *SgxEpcConfig) GetPrefault() bool {
	if o == nil || IsNil(o.Prefault) {
		var ret bool
		return ret
	}
	return *o.Prefault
}

// GetPrefaultOk returns a tuple with the Prefault field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SgxEpcConfig) GetPrefaultOk() (*bool, bool) {
	if o == nil || IsNil(o.Prefault) {
		return nil, false
	}
	return o.Prefault, true
}

// HasPrefault returns a boolean if a field has been set.
func (o *SgxEpcConfig) HasPrefault() bool {
	if o != nil && !IsNil(o.Prefault) {
		return true
	}

	return false
}

// SetPrefault gets a reference to the given bool and assigns it to the Prefault field.
func (o *SgxEpcConfig) SetPrefault(v bool) {
	o.Prefault = &v
}

func (o SgxEpcConfig) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SgxEpcConfig) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["id"] = o.Id
	toSerialize["size"] = o.Size
	if !IsNil(o.Prefault) {
		toSerialize["prefault"] = o.Prefault
	}
	return toSerialize, nil
}

func (o *SgxEpcConfig) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"id",
		"size",
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

	varSgxEpcConfig := _SgxEpcConfig{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varSgxEpcConfig)

	if err != nil {
		return err
	}

	*o = SgxEpcConfig(varSgxEpcConfig)

	return err
}

type NullableSgxEpcConfig struct {
	value *SgxEpcConfig
	isSet bool
}

func (v NullableSgxEpcConfig) Get() *SgxEpcConfig {
	return v.value
}

func (v *NullableSgxEpcConfig) Set(val *SgxEpcConfig) {
	v.value = val
	v.isSet = true
}

func (v NullableSgxEpcConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableSgxEpcConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSgxEpcConfig(val *SgxEpcConfig) *NullableSgxEpcConfig {
	return &NullableSgxEpcConfig{value: val, isSet: true}
}

func (v NullableSgxEpcConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSgxEpcConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

