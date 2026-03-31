package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidJsonSchema(t *testing.T) {
	t.Parallel()

	// 测试有效的 JSON Schema
	schemaStr := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	doc := `{"name": "John"}`
	invalidFields, err := ValidJsonSchema(schemaStr, doc)
	assert.Nil(t, err)
	assert.Empty(t, invalidFields)

	// 测试无效的 JSON 文档
	invalidDoc := `{"name": 123}`
	invalidFields, err = ValidJsonSchema(schemaStr, invalidDoc)
	assert.Nil(t, err)
	assert.Contains(t, invalidFields[0], "name")

	// 测试无效的 JSON Schema 本身
	invalidSchemaStr := `{"type": "object", "properties": {"name": {"type": "unknown"}}}`
	invalidFields, err = ValidJsonSchema(invalidSchemaStr, doc) //nolint:staticcheck,ineffassign
	assert.NotNil(t, err)
}
