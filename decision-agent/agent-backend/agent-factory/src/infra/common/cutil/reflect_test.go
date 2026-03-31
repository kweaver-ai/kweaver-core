package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStringOrNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{
			name:  "string",
			value: "hello",
			want:  true,
		},
		{
			name:  "int",
			value: 42,
			want:  true,
		},
		{
			name:  "int8",
			value: int8(42),
			want:  true,
		},
		{
			name:  "int16",
			value: int16(42),
			want:  true,
		},
		{
			name:  "int32",
			value: int32(42),
			want:  true,
		},
		{
			name:  "int64",
			value: int64(42),
			want:  true,
		},
		{
			name:  "uint",
			value: uint(42),
			want:  true,
		},
		{
			name:  "uint8",
			value: uint8(42),
			want:  true,
		},
		{
			name:  "uint16",
			value: uint16(42),
			want:  true,
		},
		{
			name:  "uint32",
			value: uint32(42),
			want:  true,
		},
		{
			name:  "uint64",
			value: uint64(42),
			want:  true,
		},
		{
			name:  "float32",
			value: float32(42.5),
			want:  true,
		},
		{
			name:  "float64",
			value: 42.5,
			want:  true,
		},
		{
			name:  "bool",
			value: true,
			want:  false,
		},
		{
			name:  "slice",
			value: []int{1, 2, 3},
			want:  false,
		},
		{
			name:  "map",
			value: map[string]int{"a": 1},
			want:  false,
		},
		{
			name:  "struct",
			value: struct{ Name string }{"test"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsStringOrNumber(tt.value)
			assert.Equal(t, tt.want, result, "IsStringOrNumber should return expected result")
		})
	}

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			IsStringOrNumber(nil)
		}, "IsStringOrNumber should panic for nil value")
	})
}

func TestIsZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{
			name:  "zero string",
			value: "",
			want:  true,
		},
		{
			name:  "zero int",
			value: 0,
			want:  true,
		},
		{
			name:  "zero float",
			value: 0.0,
			want:  true,
		},
		{
			name:  "zero bool",
			value: false,
			want:  true,
		},
		{
			name:  "nil pointer",
			value: (*int)(nil),
			want:  true,
		},
		{
			name:  "empty slice",
			value: []int{},
			want:  false,
		},
		{
			name:  "empty map",
			value: map[string]int{},
			want:  false,
		},
		{
			name:  "non-zero string",
			value: "hello",
			want:  false,
		},
		{
			name:  "non-zero int",
			value: 42,
			want:  false,
		},
		{
			name:  "non-zero float",
			value: 42.5,
			want:  false,
		},
		{
			name:  "true bool",
			value: true,
			want:  false,
		},
		{
			name:  "non-empty slice",
			value: []int{1, 2, 3},
			want:  false,
		},
		{
			name:  "non-empty map",
			value: map[string]int{"a": 1},
			want:  false,
		},
		{
			name:  "non-nil pointer",
			value: func() *int { i := 42; return &i }(),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsZeroValue(tt.value)
			assert.Equal(t, tt.want, result, "IsZeroValue should return expected result")
		})
	}
}
