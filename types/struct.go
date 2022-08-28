package types

// This file contains extensions to gogo Struct and Value types, addind MarshalJSON/UnmarshalJSON and some other helper functions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	math "math"
	strconv "strconv"
	"unicode/utf8"
)

// NewStruct constructs a Struct from a general-purpose Go map.
// The map keys must be valid UTF-8.
// The map values are converted using NewValue.
func NewStruct(v map[string]interface{}) (*Struct, error) {
	x := &Struct{Fields: make(map[string]*Value, len(v))}
	for k, v := range v {
		if !utf8.ValidString(k) {
			return nil, fmt.Errorf("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Fields[k], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// AsMap converts x to a general-purpose Go map.
// The map values are converted by calling Value.AsInterface.
func (x *Struct) AsMap() map[string]interface{} {
	vs := make(map[string]interface{})
	for k, v := range x.GetFields() {
		vs[k] = v.AsInterface()
	}
	return vs
}

// NewValue constructs a Value from a general-purpose Go interface.
//
//	╔════════════════════════╤════════════════════════════════════════════╗
//	║ Go type                │ Conversion                                 ║
//	╠════════════════════════╪════════════════════════════════════════════╣
//	║ nil                    │ stored as NullValue                        ║
//	║ bool                   │ stored as BoolValue                        ║
//	║ int, int32, int64      │ stored as NumberValue                      ║
//	║ uint, uint32, uint64   │ stored as NumberValue                      ║
//	║ float32, float64       │ stored as NumberValue                      ║
//	║ string                 │ stored as StringValue; must be valid UTF-8 ║
//	║ []byte                 │ stored as StringValue; base64-encoded      ║
//	║ map[string]interface{} │ stored as StructValue                      ║
//	║ []interface{}          │ stored as ListValue                        ║
//	╚════════════════════════╧════════════════════════════════════════════╝
//
// When converting an int64 or uint64 to a NumberValue, numeric precision loss
// is possible since they are stored as a float64.
func NewValue(v interface{}) (*Value, error) {
	switch v := v.(type) {
	case nil:
		return NewNullValue(), nil
	case bool:
		return NewBoolValue(v), nil
	case int:
		return NewNumberValue(float64(v)), nil
	case int32:
		return NewNumberValue(float64(v)), nil
	case int64:
		return NewNumberValue(float64(v)), nil
	case uint:
		return NewNumberValue(float64(v)), nil
	case uint32:
		return NewNumberValue(float64(v)), nil
	case uint64:
		return NewNumberValue(float64(v)), nil
	case float32:
		return NewNumberValue(float64(v)), nil
	case float64:
		return NewNumberValue(float64(v)), nil
	case string:
		if !utf8.ValidString(v) {
			return nil, fmt.Errorf("invalid UTF-8 in string: %q", v)
		}
		return NewStringValue(v), nil
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return NewStringValue(s), nil
	case map[string]interface{}:
		v2, err := NewStruct(v)
		if err != nil {
			return nil, err
		}
		return NewStructValue(v2), nil
	case []interface{}:
		v2, err := NewList(v)
		if err != nil {
			return nil, err
		}
		return NewListValue(v2), nil
	default:
		return nil, fmt.Errorf("invalid type: %T", v)
	}
}

// NewNullValue constructs a new null Value.
func NewNullValue() *Value {
	return &Value{Kind: &Value_NullValue{NullValue: NullValue_NULL_VALUE}}
}

// NewBoolValue constructs a new boolean Value.
func NewBoolValue(v bool) *Value {
	return &Value{Kind: &Value_BoolValue{BoolValue: v}}
}

// NewNumberValue constructs a new number Value.
func NewNumberValue(v float64) *Value {
	return &Value{Kind: &Value_NumberValue{NumberValue: v}}
}

// NewStringValue constructs a new string Value.
func NewStringValue(v string) *Value {
	return &Value{Kind: &Value_StringValue{StringValue: v}}
}

// NewStructValue constructs a new struct Value.
func NewStructValue(v *Struct) *Value {
	return &Value{Kind: &Value_StructValue{StructValue: v}}
}

// NewListValue constructs a new list Value.
func NewListValue(v *ListValue) *Value {
	return &Value{Kind: &Value_ListValue{ListValue: v}}
}

// AsInterface converts x to a general-purpose Go interface.
//
// Calling Value.MarshalJSON and "encoding/json".Marshal on this output produce
// semantically equivalent JSON (assuming no errors occur).
//
// Floating-point values (i.e., "NaN", "Infinity", and "-Infinity") are
// converted as strings to remain compatible with MarshalJSON.
func (x *Value) AsInterface() interface{} {
	switch v := x.GetKind().(type) {
	case *Value_NumberValue:
		if v != nil {
			switch {
			case math.IsNaN(v.NumberValue):
				return "NaN"
			case math.IsInf(v.NumberValue, +1):
				return "Infinity"
			case math.IsInf(v.NumberValue, -1):
				return "-Infinity"
			default:
				return v.NumberValue
			}
		}
	case *Value_StringValue:
		if v != nil {
			return v.StringValue
		}
	case *Value_BoolValue:
		if v != nil {
			return v.BoolValue
		}
	case *Value_StructValue:
		if v != nil {
			return v.StructValue.AsMap()
		}
	case *Value_ListValue:
		if v != nil {
			return v.ListValue.AsSlice()
		}
	}
	return nil
}

// NewList constructs a ListValue from a general-purpose Go slice.
// The slice elements are converted using NewValue.
func NewList(v []interface{}) (*ListValue, error) {
	x := &ListValue{Values: make([]*Value, len(v))}
	for i, v := range v {
		var err error
		x.Values[i], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// AsSlice converts x to a general-purpose Go slice.
// The slice elements are converted by calling Value.AsInterface.
func (x *ListValue) AsSlice() []interface{} {
	vs := make([]interface{}, len(x.GetValues()))
	for i, v := range x.GetValues() {
		vs[i] = v.AsInterface()
	}
	return vs
}

func (x Value) MarshalJSON() ([]byte, error) {
	switch v := x.GetKind().(type) {
	case *Value_NumberValue:
		if v != nil {
			return json.Marshal(x.Kind.(*Value_NumberValue).NumberValue)
		}
	case *Value_StringValue:
		if v != nil {
			return json.Marshal(x.Kind.(*Value_StringValue).StringValue)
		}
	case *Value_BoolValue:
		if v != nil {
			return json.Marshal(x.Kind.(*Value_BoolValue).BoolValue)
		}
	case *Value_StructValue:
		if v != nil {
			return x.Kind.(*Value_StructValue).StructValue.MarshalJSON()
		}
	case *Value_ListValue:
		if v != nil {
			return x.Kind.(*Value_ListValue).ListValue.MarshalJSON()
		}
	}
	return []byte("null"), nil
}

func (x *Value) UnmarshalJSON(b []byte) error {
	return x.unmarshal(b)
}

func (x Struct) MarshalJSON() ([]byte, error) {
	if x.Fields == nil {
		return json.Marshal(map[string]*Value{})
	}
	return json.Marshal(x.Fields)
}

func (x *Struct) UnmarshalJSON(b []byte) error {
	return x.unmarshal(b)
}

func (x ListValue) MarshalJSON() ([]byte, error) {
	if x.Values == nil {
		return json.Marshal([]*Value{})
	}
	return json.Marshal(x.Values)
}

func (x *ListValue) UnmarshalJSON(b []byte) error {
	return x.unmarshal(b)
}

func (x *Value) unmarshal(inputValue json.RawMessage) error {
	ivStr := string(inputValue)
	if ivStr == "null" {
		x.Kind = &Value_NullValue{}
	} else if v, err := strconv.ParseFloat(ivStr, 0); err == nil {
		x.Kind = &Value_NumberValue{NumberValue: v}
	} else if v, err := unquote(ivStr); err == nil {
		x.Kind = &Value_StringValue{StringValue: v}
	} else if v, err := strconv.ParseBool(ivStr); err == nil {
		x.Kind = &Value_BoolValue{BoolValue: v}
	} else if err := json.Unmarshal(inputValue, &[]json.RawMessage{}); err == nil {
		lv := &ListValue{}
		x.Kind = &Value_ListValue{ListValue: lv}
		return lv.unmarshal(inputValue)
	} else if err := json.Unmarshal(inputValue, &map[string]json.RawMessage{}); err == nil {
		sv := &Struct{}
		x.Kind = &Value_StructValue{StructValue: sv}
		return sv.unmarshal(inputValue)
	} else {
		return fmt.Errorf("unrecognized type for Value %q", ivStr)
	}
	return nil
}

func (x *ListValue) unmarshal(inputValue json.RawMessage) error {
	var s []json.RawMessage
	if err := json.Unmarshal(inputValue, &s); err != nil {
		return fmt.Errorf("bad ListValue: %v", err)
	}
	x.Values = make([]*Value, len(s))
	for i, sv := range s {
		x.Values[i] = &Value{}
		if err := x.Values[i].unmarshal(sv); err != nil {
			return err
		}
	}
	return nil
}

func (x *Struct) unmarshal(inputValue json.RawMessage) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(inputValue, &m); err != nil {
		return fmt.Errorf("bad StructValue: %v", err)
	}
	x.Fields = make(map[string]*Value)
	for k, jv := range m {
		pv := &Value{}
		if err := pv.unmarshal(jv); err != nil {
			return fmt.Errorf("bad value in StructValue for key %q: %v", k, err)
		}
		x.Fields[k] = pv
	}
	return nil
}

func unquote(s string) (string, error) {
	var ret string
	err := json.Unmarshal([]byte(s), &ret)
	return ret, err
}
