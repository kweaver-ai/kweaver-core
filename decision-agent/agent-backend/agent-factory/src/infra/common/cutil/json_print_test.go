package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintFormatJSONString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		jsonStr string
		prefix  string
		wantErr bool
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"name":"John","age":30}`,
			prefix:  "Test",
			wantErr: false,
		},
		{
			name:    "nested JSON",
			jsonStr: `{"person":{"name":"John"}}`,
			prefix:  "Data",
			wantErr: false,
		},
		{
			name:    "empty object",
			jsonStr: `{}`,
			prefix:  "Empty",
			wantErr: false,
		},
		{
			name:    "empty string",
			jsonStr: "",
			prefix:  "EmptyStr",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			jsonStr: `{"name":"John","age":30`,
			prefix:  "Invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := PrintFormatJSONString(tt.jsonStr, tt.prefix)

			if tt.wantErr {
				assert.Error(t, err, "PrintFormatJSONString should return error for invalid JSON")
			} else {
				assert.NoError(t, err, "PrintFormatJSONString should not return error")
			}
		})
	}
}

func TestPrintFormatJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   interface{}
		prefix  string
		wantErr bool
	}{
		{
			name:    "simple map",
			input:   map[string]interface{}{"name": "John", "age": 30},
			prefix:  "Data",
			wantErr: false,
		},
		{
			name:    "nested map",
			input:   map[string]interface{}{"person": map[string]interface{}{"name": "John"}},
			prefix:  "Nested",
			wantErr: false,
		},
		{
			name:    "slice",
			input:   []interface{}{"a", "b", "c"},
			prefix:  "Slice",
			wantErr: false,
		},
		{
			name:    "string",
			input:   "test",
			prefix:  "String",
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			prefix:  "Nil",
			wantErr: false,
		},
		{
			name:    "unmarshalable value (function)",
			input:   func() {},
			prefix:  "Func",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := PrintFormatJSON(tt.input, tt.prefix)
			if tt.wantErr {
				assert.Error(t, err, "PrintFormatJSON should return error for unmarshalable value")
			} else {
				assert.NoError(t, err, "PrintFormatJSON should not return error")
			}
		})
	}
}
