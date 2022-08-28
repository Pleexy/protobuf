package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

var unmarshalingTests = []struct {
	desc string
	json string
	pb   proto.Message
}{
	{"empty Struct", `{}`, &types.Struct{}},
	{"basic Struct", `{"a":"x","b":null,"c":3,"d":true}`, &types.Struct{Fields: map[string]*types.Value{
		"a": {Kind: &types.Value_StringValue{StringValue: "x"}},
		"b": {Kind: &types.Value_NullValue{}},
		"c": {Kind: &types.Value_NumberValue{NumberValue: 3}},
		"d": {Kind: &types.Value_BoolValue{BoolValue: true}},
	}}},
	{"nested Struct", `{"a":{"b":1,"c":[{"d":true},"f"]}}`, &types.Struct{Fields: map[string]*types.Value{
		"a": {Kind: &types.Value_StructValue{StructValue: &types.Struct{Fields: map[string]*types.Value{
			"b": {Kind: &types.Value_NumberValue{NumberValue: 1}},
			"c": {Kind: &types.Value_ListValue{ListValue: &types.ListValue{Values: []*types.Value{
				{Kind: &types.Value_StructValue{StructValue: &types.Struct{Fields: map[string]*types.Value{"d": {Kind: &types.Value_BoolValue{BoolValue: true}}}}}},
				{Kind: &types.Value_StringValue{StringValue: "f"}},
			}}}},
		}}}},
	}}},
	{"empty ListValue", `[]`, &types.ListValue{}},
	{"basic ListValue", `["x",null,3,true]`, &types.ListValue{Values: []*types.Value{
		{Kind: &types.Value_StringValue{StringValue: "x"}},
		{Kind: &types.Value_NullValue{}},
		{Kind: &types.Value_NumberValue{NumberValue: 3}},
		{Kind: &types.Value_BoolValue{BoolValue: true}},
	}}},
	{"number Value", `1`, &types.Value{Kind: &types.Value_NumberValue{NumberValue: 1}}},
	{"null Value", `null`, &types.Value{Kind: &types.Value_NullValue{NullValue: types.NullValue_NULL_VALUE}}},
	{"bool Value", `true`, &types.Value{Kind: &types.Value_BoolValue{BoolValue: true}}},
	{"string Value", `"x"`, &types.Value{Kind: &types.Value_StringValue{StringValue: "x"}}},
	{"string number value", `"9223372036854775807"`, &types.Value{Kind: &types.Value_StringValue{StringValue: "9223372036854775807"}}},
	{"list of lists Value", `["x",[["y"],"z"]]`, &types.Value{
		Kind: &types.Value_ListValue{ListValue: &types.ListValue{
			Values: []*types.Value{
				{Kind: &types.Value_StringValue{StringValue: "x"}},
				{Kind: &types.Value_ListValue{ListValue: &types.ListValue{
					Values: []*types.Value{
						{Kind: &types.Value_ListValue{ListValue: &types.ListValue{
							Values: []*types.Value{{Kind: &types.Value_StringValue{StringValue: "y"}}},
						}}},
						{Kind: &types.Value_StringValue{StringValue: "z"}},
					},
				}}},
			},
		}}}},

	{"StructValue containing StringValue's", `{"escaped":"a/b","unicode":"\u00004E16\u0000754C"}`,
		&types.Struct{
			Fields: map[string]*types.Value{
				"escaped": {Kind: &types.Value_StringValue{StringValue: "a/b"}},
				"unicode": {Kind: &types.Value_StringValue{StringValue: "\u00004E16\u0000754C"}},
			},
		}},
}

func TestUnmarshaling(t *testing.T) {
	for _, tt := range unmarshalingTests {
		// Make a new instance of the type of our expected object.
		p := reflect.New(reflect.TypeOf(tt.pb).Elem()).Interface().(proto.Message)
		err := json.Unmarshal([]byte(tt.json), p)

		if err != nil {
			t.Errorf("unmarshalling %s: %v", tt.desc, err)
			continue
		}

		// For easier diffs, compare text strings of the protos.
		exp := proto.MarshalTextString(tt.pb)
		act := proto.MarshalTextString(p)
		if string(exp) != string(act) {
			t.Errorf("%s: got [%s] want [%s]", tt.desc, act, exp)
		}
	}
}

func TestMarshaling(t *testing.T) {
	for _, tt := range unmarshalingTests {
		p := reflect.Indirect(reflect.ValueOf(tt.pb)).Interface()
		data, err := json.Marshal(p)
		if err != nil {
			t.Errorf("marshalling %s: %v", tt.desc, err)
			continue
		}

		// For easier diffs, compare text strings of the protos.
		if string(data) != string(tt.json) {
			t.Errorf("%s: got %s want %s", tt.desc, string(data), tt.json)
		}
	}
}
