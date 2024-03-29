/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package queue

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterSpec_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		want FilterSpec
		args args
	}{
		{
			name: "single filter",
			want: FilterSpec{
				Filters: []FilterFieldsConfig{
					{
						FieldName:     "session_alias",
						Operation:     Equal,
						ExpectedValue: "42",
					},
				},
			},
			args: args{
				data: []byte(`
				{
					"session_alias": {
						"operation": "EQUAL",
						"value": "42"
					}
				}`),
			},
		},
		{
			name: "multiple filters",
			want: FilterSpec{
				Filters: []FilterFieldsConfig{
					{
						FieldName:     "session_alias",
						Operation:     NotEqual,
						ExpectedValue: "42",
					},
					{
						FieldName:     "session_alias",
						Operation:     NotEqual,
						ExpectedValue: "43",
					},
				},
			},
			args: args{
				data: []byte(`
				[
					{
						"fieldName": "session_alias",
						"operation": "NOT_EQUAL",
						"expectedValue": "42"
					},
					{
						"fieldName": "session_alias",
						"operation": "NOT_EQUAL",
						"expectedValue": "43"
					}
				]`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := FilterSpec{}
			if err := json.Unmarshal(tt.args.data, &fc); err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
			}
			if !reflect.DeepEqual(fc, tt.want) {
				t.Errorf("deserialized object %#v does not match expected", fc)
			}
		})
	}
}

func TestFilterOperation_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		args   args
		wantOp FilterOperation
	}{
		{
			wantOp: Wildcard,
			args: args{
				data: []byte(`"WILDCARD"`),
			},
		},
		{
			wantOp: Empty,
			args: args{
				data: []byte(`"EMPTY"`),
			},
		},
		{
			wantOp: NotEmpty,
			args: args{
				data: []byte(`"NOT_EMPTY"`),
			},
		},
		{
			wantOp: Equal,
			args: args{
				data: []byte(`"EQUAL"`),
			},
		},
		{
			wantOp: NotEqual,
			args: args{
				data: []byte(`"NOT_EQUAL"`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.wantOp), func(t *testing.T) {
			var op FilterOperation
			if err := json.Unmarshal(tt.args.data, &op); err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
			}
			if tt.wantOp != op {
				t.Errorf("Unxpected operation (expected: %v, actual: %v)",
					tt.wantOp, op)
			}
		})
	}
}
