package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDynamicStruct(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"name": {"type": "string"},
		"age": {"type": "integer"},
		"active": {"type": "boolean"}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
	assert.Equal(t, "struct", typ.Kind().String(), "CreateDynamicStruct should return struct type")
	assert.Equal(t, 3, typ.NumField(), "CreateDynamicStruct should return struct with 3 fields")
}

func TestCreateDynamicStruct_InvalidSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		schemaStr string
		wantErr   bool
	}{
		{
			name:      "invalid JSON",
			schemaStr: `{invalid}`,
			wantErr:   true,
		},
		{
			name:      "empty schema",
			schemaStr: `{}`,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			typ, err := CreateDynamicStruct(tt.schemaStr)

			if tt.wantErr {
				assert.Error(t, err, "CreateDynamicStruct should return error for invalid schema")
				assert.Nil(t, typ, "CreateDynamicStruct should return nil type for invalid schema")
			} else {
				assert.NoError(t, err, "CreateDynamicStruct should not return error for empty schema")
				assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type for empty schema")
			}
		})
	}
}

func TestCreateDynamicStruct_NestedObject(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"user": {
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			}
		}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error for nested object")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
}
