package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	json := JSON()
	assert.NotNil(t, json)
}

func TestJSONObjectToArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "简单的对象",
			json: `{"key":"value"}`,
			want: "[{\"key\":\"value\"}]",
		},
		{
			name: "嵌套对象",
			json: `{"a":{"b":"c"}}`,
			want: "[{\"a\":{\"b\":\"c\"}}]",
		},
		{
			name: "空对象",
			json: `{}`,
			want: "[{}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := JSONObjectToArray([]byte(tt.json))
			assert.Equal(t, tt.want, string(result))
		})
	}
}

func TestFormatJSONString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "有效的JSON字符串",
			input:   `{"name":"John","age":30}`,
			wantErr: false,
		},
		{
			name:    "嵌套对象",
			input:   `{"person":{"name":"John"}}`,
			wantErr: false,
		},
		{
			name:    "空对象",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "空字符串",
			input:   "",
			wantErr: false,
		},
		{
			name:    "无效的JSON",
			input:   `{"name":"John","age":30`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FormatJSONString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.name == "空字符串" {
					assert.Empty(t, result)
				} else {
					assert.NotEmpty(t, result)
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "简单对象",
			input:   map[string]interface{}{"name": "John", "age": 30},
			wantErr: false,
		},
		{
			name:    "嵌套对象",
			input:   map[string]interface{}{"person": map[string]interface{}{"name": "John"}},
			wantErr: false,
		},
		{
			name:    "切片",
			input:   []interface{}{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "unmarshalable value (function)",
			input:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FormatJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestToMapByJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   interface{}
		wantKey string
		wantErr bool
	}{
		{
			name:    "简单对象",
			input:   map[string]interface{}{"name": "John", "age": 30},
			wantKey: "name",
			wantErr: false,
		},
		{
			name:    "嵌套对象",
			input:   map[string]interface{}{"person": map[string]interface{}{"name": "John"}},
			wantKey: "person",
			wantErr: false,
		},
		{
			name:    "invalid input - channel",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToMapByJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result, tt.wantKey)
			}
		})
	}
}
