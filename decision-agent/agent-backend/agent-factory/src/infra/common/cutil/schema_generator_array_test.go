package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDynamicStruct_WithArrays(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"names": {
			"type": "array",
			"items": {
				"type": "string"
			}
		},
		"scores": {
			"type": "array",
			"items": {
				"type": "number"
			}
		},
		"flags": {
			"type": "array",
			"items": {
				"type": "boolean"
			}
		},
		"counts": {
			"type": "array",
			"items": {
				"type": "integer"
			}
		},
		"dynamic": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": {"type": "string"}
				}
			}
		}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error for arrays")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
	assert.Equal(t, "struct", typ.Kind().String(), "CreateDynamicStruct should return struct type")
	assert.Equal(t, 5, typ.NumField(), "CreateDynamicStruct should return struct with 5 fields")

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		assert.Equal(t, "slice", field.Type.Kind().String(), "All fields should be slice type")
	}
}

func TestCreateDynamicStruct_ArrayWithObjectItems(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"items": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"id": {"type": "integer"},
					"name": {"type": "string"}
				}
			}
		}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error for array with object items")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
	assert.Equal(t, 1, typ.NumField(), "CreateDynamicStruct should return struct with 1 field")

	field := typ.Field(0)
	assert.Equal(t, "slice", field.Type.Kind().String(), "Field should be slice type")

	elemType := field.Type.Elem()
	assert.Equal(t, "struct", elemType.Kind().String(), "Slice element should be struct type")
	assert.Equal(t, 2, elemType.NumField(), "Struct should have 2 fields")
}

func TestCreateDynamicStruct_ArrayWithObjectNoProperties(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"items": {
			"type": "array",
			"items": {
				"type": "object"
			}
		}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error for array with object items without properties")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
	assert.Equal(t, 1, typ.NumField(), "CreateDynamicStruct should return struct with 1 field")

	field := typ.Field(0)
	assert.Equal(t, "slice", field.Type.Kind().String(), "Field should be slice type")

	// When object has no properties, it creates an empty struct
	elemType := field.Type.Elem()
	assert.Equal(t, "struct", elemType.Kind().String(), "Slice element should be struct type for object without properties")
	assert.Equal(t, 0, elemType.NumField(), "Struct should have no fields for object without properties")
}

func TestCreateDynamicStruct_ArrayWithUnknownType(t *testing.T) {
	t.Parallel()

	schemaStr := `{
		"unknown": {
			"type": "array",
			"items": {
				"type": "unknown_type"
			}
		}
	}`

	typ, err := CreateDynamicStruct(schemaStr)

	assert.NoError(t, err, "CreateDynamicStruct should not return error for array with unknown type")
	assert.NotNil(t, typ, "CreateDynamicStruct should return non-nil type")
	assert.Equal(t, 1, typ.NumField(), "CreateDynamicStruct should return struct with 1 field")

	field := typ.Field(0)
	assert.Equal(t, "slice", field.Type.Kind().String(), "Field should be slice type")

	// Unknown type defaults to interface{}
	elemType := field.Type.Elem()
	assert.NotNil(t, elemType, "Slice element should have a type even for unknown type")
}

func TestGetArrayElementType_ObjectWithNilProperties(t *testing.T) {
	t.Parallel()

	// Test the case where prop.Type is "object" and prop.Properties is nil
	// This should return map[string]interface{}
	prop := &JSONSchemaProperty{
		Type:       "object",
		Properties: nil, // Explicitly nil to trigger the map return path
	}

	elemType, err := getArrayElementType(prop)

	assert.NoError(t, err, "getArrayElementType should not return error for object with nil properties")
	assert.NotNil(t, elemType, "getArrayElementType should return non-nil type")
	assert.Equal(t, "map", elemType.Kind().String(), "Element type should be map for object with nil properties")
}
