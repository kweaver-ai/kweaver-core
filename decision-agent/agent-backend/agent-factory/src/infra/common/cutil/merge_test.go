package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeJSONStrings_Basic(t *testing.T) {
	t.Parallel()

	json1 := `{"name":"John","age":30}`
	json2 := `{"age":35,"city":"NY"}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "35")
	assert.Contains(t, result, "NY")
}

func TestMergeJSONStrings_InvalidJSON1(t *testing.T) {
	t.Parallel()

	json1 := `{invalid json}`
	json2 := `{"age":35}`

	result, err := MergeJSONStrings(json1, json2)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to parse jsonStr1")
}

func TestMergeJSONStrings_InvalidJSON2(t *testing.T) {
	t.Parallel()

	json1 := `{"name":"John"}`
	json2 := `{invalid json}`

	result, err := MergeJSONStrings(json1, json2)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to parse jsonStr2")
}

func TestMergeJSONStrings_EmptyStrings(t *testing.T) {
	t.Parallel()

	json1 := `{}`
	json2 := `{}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.Equal(t, `{}`, result)
}

func TestMergeJSONStrings_Nested(t *testing.T) {
	t.Parallel()

	json1 := `{"user":{"name":"John"},"age":30}`
	json2 := `{"user":{"age":35},"city":"NY"}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.Contains(t, result, "John")
	assert.Contains(t, result, "35")
}

func TestMergeMapInterface_Basic(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	i := map[string]interface{}{
		"age":  35,
		"city": "NY",
	}

	err := MergeMapInterface(m, i)

	assert.NoError(t, err)
	assert.Equal(t, "John", m["name"])
	assert.Equal(t, float64(35), m["age"])
	assert.Equal(t, "NY", m["city"])
}

func TestMergeMapInterface_NestedMap(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
		},
	}

	i := map[string]interface{}{
		"user": map[string]interface{}{
			"age": 35,
		},
	}

	err := MergeMapInterface(m, i)

	assert.NoError(t, err)
	assert.NotNil(t, m["user"])
}

func TestMergeMapInterface_InvalidInterface(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{}

	i := make(chan int) // channels cannot be marshaled to JSON

	err := MergeMapInterface(m, i)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal")
}

func TestMergeJSONStrings_Overwrite(t *testing.T) {
	t.Parallel()

	json1 := `{"name":"John","age":30}`
	json2 := `{"name":"Jane","age":35}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.NotContains(t, result, "John")
	assert.Contains(t, result, "Jane")
	assert.Contains(t, result, "35")
}

func TestMergeJSONStrings_Arrays(t *testing.T) {
	t.Parallel()

	json1 := `{"items":[1,2,3]}`
	json2 := `{"items":[4,5]}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.Contains(t, result, "items")
}

func TestMergeMapInterface_EmptyMap(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{}

	i := map[string]interface{}{
		"name": "John",
	}

	err := MergeMapInterface(m, i)

	assert.NoError(t, err)
	assert.Equal(t, "John", m["name"])
}

func TestMergeJSONStrings_WithArray(t *testing.T) {
	t.Parallel()

	json1 := `{"data":["a","b"],"name":"test"}`
	json2 := `{"data":["c"],"age":30}`

	result, err := MergeJSONStrings(json1, json2)

	assert.NoError(t, err)
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "30")
}
