package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		length    int
		wantLen   int
		wantPanic bool
	}{
		{
			name:      "valid length 10",
			length:    10,
			wantLen:   10,
			wantPanic: false,
		},
		{
			name:      "valid length 50",
			length:    50,
			wantLen:   50,
			wantPanic: false,
		},
		{
			name:      "valid length 100",
			length:    100,
			wantLen:   100,
			wantPanic: false,
		},
		{
			name:      "zero length",
			length:    0,
			wantPanic: true,
		},
		{
			name:      "negative length",
			length:    -1,
			wantPanic: true,
		},
		{
			name:      "too large length",
			length:    101,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				assert.Panics(t, func() {
					GenerateRandomString(tt.length)
				}, "GenerateRandomString should panic")
			} else {
				result := GenerateRandomString(tt.length)
				assert.Len(t, result, tt.wantLen, "Generated string should have correct length")
				assert.NotEmpty(t, result, "Generated string should not be empty")
			}
		})
	}
}

func TestStringSplitAndJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple colon separated",
			input: "a:b:c",
			want:  "a\nb\nc",
		},
		{
			name:  "single colon",
			input: "a:b",
			want:  "a\nb",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no colons",
			input: "abc",
			want:  "abc",
		},
		{
			name:  "multiple consecutive colons",
			input: "a::b",
			want:  "a\n\nb",
		},
		{
			name:  "with numbers",
			input: "1:2:3",
			want:  "1\n2\n3",
		},
		{
			name:  "with mixed content",
			input: "hello:world:123",
			want:  "hello\nworld\n123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := StringSplitAndJoin(tt.input)
			assert.Equal(t, tt.want, result, "StringSplitAndJoin should return expected result")
		})
	}
}
