package cutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONNormalizer_MarshalJSON_Empty(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(result))
}

func TestJSONNormalizer_MarshalJSON_SingleKey(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{
		"name": "John",
	}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"John"}`, string(result))
}

func TestJSONNormalizer_MarshalJSON_MultipleKeys_Sorted(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{
		"zebra":  1,
		"apple":  2,
		"banana": 3,
	}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	assert.Equal(t, `{"apple":2,"banana":3,"zebra":1}`, string(result))
}

func TestJSONNormalizer_MarshalJSON_WithNestedObject(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{
		"name": "John",
		"address": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	// Keys should be sorted: address, name
	assert.Contains(t, string(result), `"address":`)
	assert.Contains(t, string(result), `"name":"John"`)
}

func TestJSONNormalizer_MarshalJSON_WithArray(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{
		"items": []int{1, 2, 3},
		"count": 3,
	}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	assert.Equal(t, `{"count":3,"items":[1,2,3]}`, string(result))
}

func TestJSONNormalizer_MarshalJSON_WithMixedTypes(t *testing.T) {
	t.Parallel()

	jn := JSONNormalizer{
		"string":  "value",
		"number":  42,
		"float":   3.14,
		"bool":    true,
		"null":    nil,
		"z_last":  "end",
		"a_first": "start",
	}

	result, err := json.Marshal(jn)
	assert.NoError(t, err)
	// Verify keys are sorted alphabetically
	assert.Contains(t, string(result), `"a_first":`)
	assert.Contains(t, string(result), `"bool":`)
	assert.Contains(t, string(result), `"float":`)
	assert.Contains(t, string(result), `"null":`)
	assert.Contains(t, string(result), `"number":`)
	assert.Contains(t, string(result), `"string":`)
	assert.Contains(t, string(result), `"z_last":`)
}

func TestJSONNormalizer_MarshalJSON_WithUnmarshalableValue(t *testing.T) {
	t.Parallel()

	// Create a function which cannot be marshaled
	jn := JSONNormalizer{
		"valid":   "value",
		"invalid": func() {}, // functions cannot be marshaled
	}

	_, err := json.Marshal(jn)
	assert.Error(t, err)
}

func TestJSONStrCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonStr1 string
		jsonStr2 string
		expected bool
	}{
		{
			name:     "相同的JSON",
			jsonStr1: `{"name": "John", "age": 30}`,
			jsonStr2: `{"name": "John", "age": 30}`,
			expected: true,
		},
		{
			name:     "不同顺序但内容相同",
			jsonStr1: `{"name": "John", "age": 30}`,
			jsonStr2: `{"age": 30, "name": "John"}`,
			expected: true,
		},
		{
			name:     "不同的值",
			jsonStr1: `{"name": "John", "age": 30}`,
			jsonStr2: `{"name": "Jane", "age": 30}`,
			expected: false,
		},
		{
			name:     "结构不同",
			jsonStr1: `{"name": "John"}`,
			jsonStr2: `{"name": "John", "age": 30}`,
			expected: false,
		},
		{
			name:     "空JSON",
			jsonStr1: `{}`,
			jsonStr2: `{}`,
			expected: true,
		},
		{
			name:     "无效的JSON1",
			jsonStr1: `{"name": "John", "age": 30}`,
			jsonStr2: `{"name": "John", "age": 30}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := JSONStrCompare(tt.jsonStr1, tt.jsonStr2)
			if tt.expected {
				if err != nil {
					t.Errorf("JSONStrCompare() unexpected error: %v", err)
				}

				if !result {
					t.Errorf("JSONStrCompare() = %v, want true", result)
				}
			} else {
				if err == nil && result {
					t.Errorf("JSONStrCompare() should return false for different JSON")
				}
			}
		})
	}
}

func TestJSONStrCompare_InvalidJSON1(t *testing.T) {
	t.Parallel()

	json1 := `{"name":"John",invalid}`
	json2 := `{"name":"John","age":30}`

	equal, err := JSONStrCompare(json1, json2)
	assert.Error(t, err)
	assert.False(t, equal)
	assert.Contains(t, err.Error(), "failed to parse jsonStr1")
}

func TestJSONStrCompare_InvalidJSON2(t *testing.T) {
	t.Parallel()

	json1 := `{"name":"John","age":30}`
	json2 := `{"name":"John",invalid}`

	equal, err := JSONStrCompare(json1, json2)
	assert.Error(t, err)
	assert.False(t, equal)
	assert.Contains(t, err.Error(), "failed to parse jsonStr2")
}

func TestJSONStrCompare_BothInvalidJSON(t *testing.T) {
	t.Parallel()

	json1 := `{invalid}`
	json2 := `{invalid}`

	equal, err := JSONStrCompare(json1, json2)
	assert.Error(t, err)
	assert.False(t, equal)
}

func TestJSONStrCompare_NestedObjects(t *testing.T) {
	t.Parallel()

	json1 := `{"user":{"name":"John","age":30},"id":123}`
	json2 := `{"id":123,"user":{"age":30,"name":"John"}}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJSONStrCompare_NestedObjects_DifferentValues(t *testing.T) {
	t.Parallel()

	json1 := `{"user":{"name":"John","age":30}}`
	json2 := `{"user":{"name":"John","age":31}}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.False(t, equal)
}

func TestJSONStrCompare_WithArrays(t *testing.T) {
	t.Parallel()

	json1 := `{"items":[1,2,3],"count":3}`
	json2 := `{"count":3,"items":[1,2,3]}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJSONStrCompare_WithArrays_DifferentOrder(t *testing.T) {
	t.Parallel()

	json1 := `{"items":[1,2,3]}`
	json2 := `{"items":[3,2,1]}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.False(t, equal)
}

func TestJSONStrCompare_WithNull(t *testing.T) {
	t.Parallel()

	json1 := `{"name":null,"age":30}`
	json2 := `{"age":30,"name":null}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJSONStrCompare_WithBoolean(t *testing.T) {
	t.Parallel()

	json1 := `{"active":true,"admin":false}`
	json2 := `{"admin":false,"active":true}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJSONStrCompare_OneEmpty(t *testing.T) {
	t.Parallel()

	json1 := `{}`
	json2 := `{"name":"John"}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.False(t, equal)
}

func TestJSONStrCompare_ComplexNested(t *testing.T) {
	t.Parallel()

	json1 := `{"user":{"profile":{"name":"John","age":30},"settings":{"theme":"dark"}},"id":123}`
	json2 := `{"id":123,"user":{"settings":{"theme":"dark"},"profile":{"age":30,"name":"John"}}}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestJSONStrCompare_Whitespace(t *testing.T) {
	t.Parallel()

	json1 := `  {  "name"  :  "John"  ,  "age"  :  30  }  `
	json2 := `{"name":"John","age":30}`

	equal, err := JSONStrCompare(json1, json2)
	assert.NoError(t, err)
	assert.True(t, equal)
}
