package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceToStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		slice []string
		sep   string
		want  string
	}{
		{
			name:  "string slice with empty sep",
			slice: []string{"a", "b", "c"},
			sep:   "",
			want:  "abc",
		},
		{
			name:  "string slice with comma sep",
			slice: []string{"a", "b", "c"},
			sep:   ",",
			want:  "a,b,c",
		},
		{
			name:  "empty string slice",
			slice: []string{},
			sep:   "",
			want:  "",
		},
		{
			name:  "single element",
			slice: []string{"x"},
			sep:   "",
			want:  "x",
		},
		{
			name:  "with empty strings",
			slice: []string{"a", "", "c"},
			sep:   "",
			want:  "ac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := SliceToStr(tt.slice, tt.sep)
			assert.Equal(t, tt.want, result, "SliceToStr should return expected result")
		})
	}
}
