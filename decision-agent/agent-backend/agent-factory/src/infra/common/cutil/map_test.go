package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepCopyMap(t *testing.T) {
	t.Parallel()

	// 1. Test case with simple key-value pairs
	original := map[string]interface{}{
		"key1": "value1",
		"key2": 10,
		"key3": []string{"value3a", "value3b"},
	}
	copied, err := DeepCopyMap(original)

	assert.Nil(t, err)
	assert.NotNil(t, copied)

	// Convert []interface{} to []string for comparison
	if copiedSlice, ok := copied["key3"].([]interface{}); ok {
		var stringSlice []string

		for _, v := range copiedSlice {
			if str, ok := v.(string); ok {
				stringSlice = append(stringSlice, str)
			}
		}

		copied["key3"] = stringSlice
	}

	// Compare key2 separately considering float conversion
	if v, ok := copied["key2"].(float64); ok {
		copied["key2"] = int(v)
	}

	assert.Equal(t, original, copied)

	// Modify the original map and check the copied map remains unchanged
	original["key1"] = "new_value1"
	assert.NotEqual(t, original, copied)

	// 2. Test case with nested maps
	nestedOriginal := map[string]interface{}{
		"nestedKey": map[string]interface{}{
			"subKey": "subValue",
		},
	}
	nestedCopied, err := DeepCopyMap(nestedOriginal)

	assert.Nil(t, err)
	assert.NotNil(t, nestedCopied)
	assert.Equal(t, nestedOriginal, nestedCopied)

	// Modify nested map in original and check the copied map remains unchanged
	originalNestedMap := nestedOriginal["nestedKey"].(map[string]interface{})
	originalNestedMap["subKey"] = "newSubValue"

	assert.NotEqual(t, nestedOriginal, nestedCopied)

	// 3. Test with an empty map
	emptyOriginal := make(map[string]interface{})
	emptyCopied, err := DeepCopyMap(emptyOriginal)

	assert.Nil(t, err)
	assert.NotNil(t, emptyCopied)
	assert.Equal(t, emptyOriginal, emptyCopied)
}
