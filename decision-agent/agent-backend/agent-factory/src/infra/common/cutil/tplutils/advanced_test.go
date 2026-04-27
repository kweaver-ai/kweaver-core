package tplutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate_Advanced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tmplStr   string
		data      map[string]interface{}
		wantErr   bool
		wantEmpty bool
	}{
		{
			name:    "nil data",
			tmplStr: "{{.name}}",
			data:    nil,
			wantErr: false,
		},
		{
			name:      "empty template",
			tmplStr:   "",
			data:      map[string]interface{}{"name": "test"},
			wantErr:   false,
			wantEmpty: true,
		},
		{
			name:    "template with undefined variable",
			tmplStr: "{{.undefined}}",
			data:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
		{
			name:    "template with number value",
			tmplStr: "{{.count}}",
			data:    map[string]interface{}{"count": 42},
			wantErr: false,
		},
		{
			name:    "template with boolean value",
			tmplStr: "{{.active}}",
			data:    map[string]interface{}{"active": true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := RenderTemplate(tt.tmplStr, tt.data)

			if tt.wantErr {
				assert.Error(t, err, "RenderTemplate should return error")
			} else {
				assert.NoError(t, err, "RenderTemplate should not return error")

				if tt.wantEmpty {
					assert.Empty(t, result, "Result should be empty")
				}
			}
		})
	}
}

func TestSafeRenderTemplate_Advanced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tmplStr string
		data    map[string]interface{}
	}{
		{
			name:    "nil data",
			tmplStr: "{{.name}}",
			data:    nil,
		},
		{
			name:    "nested map with nested key",
			tmplStr: "{{.user.profile.name}}",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name": "John",
					},
				},
			},
		},
		{
			name:    "deeply nested map with missing key",
			tmplStr: "{{.a.b.c.d}}",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 123,
					},
				},
			},
		},
		{
			name:    "map with number values",
			tmplStr: "{{.count}}:{{.value}}",
			data: map[string]interface{}{
				"count": 100,
				"value": 3.14,
			},
		},
		{
			name:    "map with boolean values",
			tmplStr: "{{.enabled}}:{{.active}}",
			data: map[string]interface{}{
				"enabled": true,
				"active":  false,
			},
		},
		{
			name:    "map with nil value",
			tmplStr: "{{.nilValue}}",
			data: map[string]interface{}{
				"nilValue": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := SafeRenderTemplate(tt.tmplStr, tt.data)
			assert.NoError(t, err, "SafeRenderTemplate should not return error")
			assert.NotEmpty(t, result, "SafeRenderTemplate should return non-empty result")
		})
	}
}
