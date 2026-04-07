package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "snake_case to camelCase",
			input: "snake_case",
			want:  "SnakeCase",
		},
		{
			name:  "single word",
			input: "hello",
			want:  "Hello",
		},
		{
			name:  "multiple underscores",
			input: "multiple_underscores_case",
			want:  "MultipleUnderscoresCase",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "leading underscore",
			input: "_leading",
			want:  "Leading",
		},
		{
			name:  "trailing underscore",
			input: "trailing_",
			want:  "Trailing",
		},
		{
			name:  "consecutive underscores",
			input: "double__underscore",
			want:  "DoubleUnderscore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := toCamelCase(tt.input)
			assert.Equal(t, tt.want, result, "toCamelCase should return expected result")
		})
	}
}

func TestGenerateTypeCacheKey(t *testing.T) {
	t.Parallel()

	prop1 := &JSONSchemaProperty{
		Type: "string",
	}

	prop2 := &JSONSchemaProperty{
		Type: "object",
	}

	tests := []struct {
		name string
		prop *JSONSchemaProperty
	}{
		{
			name: "string property",
			prop: prop1,
		},
		{
			name: "object property",
			prop: prop2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := generateTypeCacheKey(tt.prop)
			assert.NotEmpty(t, result, "generateTypeCacheKey should return non-empty result")
		})
	}
}
