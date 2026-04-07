package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddKeyToJSONArrayBys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonStr  string
		key      string
		valueStr string
		want     string
		wantErr  bool
	}{
		{
			name:     "add key to array element",
			jsonStr:  `[{"id":1},{"id":2}]`,
			key:      "name",
			valueStr: "John",
			want:     `[{"id":1,"name":"John"},{"id":2,"name":"John"}]`,
			wantErr:  false,
		},
		{
			name:     "add key to single object",
			jsonStr:  `[{"id":1}]`,
			key:      "name",
			valueStr: "Jane",
			want:     `[{"id":1,"name":"Jane"}]`,
			wantErr:  false,
		},
		{
			name:     "empty array",
			jsonStr:  `[]`,
			key:      "name",
			valueStr: "Test",
			want:     `[]`,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			jsonStr:  `invalid`,
			key:      "name",
			valueStr: "Test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := AddKeyToJSONArrayBys([]byte(tt.jsonStr), tt.key, tt.valueStr)

			if tt.wantErr {
				assert.Error(t, err, "AddKeyToJSONArrayBys should return error")
			} else {
				assert.NoError(t, err, "AddKeyToJSONArrayBys should not return error")
				assert.JSONEq(t, tt.want, string(result), "AddKeyToJSONArrayBys should return expected result")
			}
		})
	}
}

func TestAddJSONArrayToJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonStr  string
		key      string
		arrayStr string
		want     string
		wantErr  bool
	}{
		{
			name:     "add array to object",
			jsonStr:  `{"id":1}`,
			key:      "items",
			arrayStr: `[1,2,3]`,
			want:     `{"id":1,"items":[1,2,3]}`,
			wantErr:  false,
		},
		{
			name:     "add empty array",
			jsonStr:  `{"name":"test"}`,
			key:      "items",
			arrayStr: `[]`,
			want:     `{"name":"test","items":[]}`,
			wantErr:  false,
		},
		{
			name:     "add to existing array key",
			jsonStr:  `{"id":1,"items":[4]}`,
			key:      "items",
			arrayStr: `[1,2,3]`,
			want:     `{"id":1,"items":[1,2,3]}`,
			wantErr:  false,
		},
		{
			name:     "invalid jsonStr",
			jsonStr:  `invalid`,
			key:      "items",
			arrayStr: `[1,2,3]`,
			wantErr:  true,
		},
		{
			name:     "invalid arrayStr",
			jsonStr:  `{"id":1}`,
			key:      "items",
			arrayStr: `invalid`,
			want:     `{"id":1,"items":0}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := AddJSONArrayToJSON(tt.jsonStr, tt.key, tt.arrayStr)

			if tt.wantErr {
				assert.Error(t, err, "AddJSONArrayToJSON should return error")
			} else {
				assert.NoError(t, err, "AddJSONArrayToJSON should not return error")
				assert.JSONEq(t, tt.want, string(result), "AddJSONArrayToJSON should return expected result")
			}
		})
	}
}
