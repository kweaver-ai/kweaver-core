package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustParseInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		want     int
		panicMsg string
	}{
		{
			name:  "valid positive integer",
			input: "123",
			want:  123,
		},
		{
			name:  "valid zero",
			input: "0",
			want:  0,
		},
		{
			name:  "valid negative integer",
			input: "-456",
			want:  -456,
		},
		{
			name:     "invalid string",
			input:    "abc",
			panicMsg: "invalid",
		},
		{
			name:     "empty string",
			input:    "",
			panicMsg: "invalid",
		},
		{
			name:     "float string",
			input:    "123.45",
			panicMsg: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panicMsg != "" {
				assert.Panics(t, func() {
					MustParseInt(tt.input)
				}, "Should panic for invalid input")
			} else {
				result := MustParseInt(tt.input)
				assert.Equal(t, tt.want, result, "Result should match expected value")
			}
		})
	}
}

func TestMustParseInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		want     int64
		panicMsg string
	}{
		{
			name:  "valid positive integer",
			input: "123",
			want:  123,
		},
		{
			name:  "valid zero",
			input: "0",
			want:  0,
		},
		{
			name:  "valid negative integer",
			input: "-456",
			want:  -456,
		},
		{
			name:  "large number",
			input: "9223372036854775807",
			want:  9223372036854775807,
		},
		{
			name:     "invalid string",
			input:    "abc",
			panicMsg: "invalid",
		},
		{
			name:     "empty string",
			input:    "",
			panicMsg: "invalid",
		},
		{
			name:     "float string",
			input:    "123.45",
			panicMsg: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.panicMsg != "" {
				assert.Panics(t, func() {
					MustParseInt64(tt.input)
				}, "Should panic for invalid input")
			} else {
				result := MustParseInt64(tt.input)
				assert.Equal(t, tt.want, result, "Result should match expected value")
			}
		})
	}
}

func TestStringToBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "true lowercase",
			input: "true",
			want:  true,
		},
		{
			name:  "true uppercase",
			input: "TRUE",
			want:  true,
		},
		{
			name:  "true mixed case",
			input: "True",
			want:  true,
		},
		{
			name:  "false lowercase",
			input: "false",
			want:  false,
		},
		{
			name:  "false uppercase",
			input: "FALSE",
			want:  false,
		},
		{
			name:  "false mixed case",
			input: "False",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "random string",
			input: "random",
			want:  true,
		},
		{
			name:  "number 1",
			input: "1",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := StringToBool(tt.input)
			assert.Equal(t, tt.want, result, "Result should match expected value")
		})
	}
}

func TestRuneLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "ASCII string",
			input: "hello",
			want:  5,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "single character",
			input: "a",
			want:  1,
		},
		{
			name:  "Chinese characters",
			input: "你好",
			want:  2,
		},
		{
			name:  "mixed ASCII and Chinese",
			input: "hello你好",
			want:  7,
		},
		{
			name:  "emoji",
			input: "😀😀",
			want:  2,
		},
		{
			name:  "with spaces",
			input: "hello world",
			want:  11,
		},
		{
			name:  "special characters",
			input: "!@#$%",
			want:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := RuneLength(tt.input)
			assert.Equal(t, tt.want, result, "Result should match expected length")
		})
	}
}
