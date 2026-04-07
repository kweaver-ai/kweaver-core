package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyUseJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dst     interface{}
		src     interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "simple struct copy",
			dst: &struct {
				Name string
				Age  int
			}{},
			src: struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			want: &struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:    "copy with map",
			dst:     &map[string]int{},
			src:     map[string]int{"a": 1, "b": 2},
			want:    &map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
		{
			name:    "copy with slice",
			dst:     &[]string{},
			src:     []string{"a", "b", "c"},
			want:    &[]string{"a", "b", "c"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := CopyUseJSON(tt.dst, tt.src)
			if tt.wantErr {
				assert.Error(t, err, "CopyUseJSON should return error")
			} else {
				assert.NoError(t, err, "CopyUseJSON should not return error")
				assert.Equal(t, tt.want, tt.dst, "CopyUseJSON should copy correctly")
			}
		})
	}
}
