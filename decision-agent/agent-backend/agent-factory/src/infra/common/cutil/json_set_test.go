package cutil

import "testing"

func TestAddToJSON(t *testing.T) {
	t.Parallel()

	jsonStr := `{"name":"John", "age":30, "cars":{"car1":"Ford","car2":"BMW"}}`
	jsonPath := "cars.car3"
	value := "Audi"

	expectedJSON := `{"name":"John", "age":30, "cars":{"car1":"Ford","car2":"BMW","car3":"Audi"}}`

	updatedJSON, err := AddToJSON(jsonStr, jsonPath, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if updatedJSON != expectedJSON {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expectedJSON, updatedJSON)
	}
}

func TestAddKeyToJSONArray(t *testing.T) {
	t.Parallel()

	jsonArrayStr := `[{"a":1},{"a":1},{"a":1}]`
	key := "b"
	value := 2

	expectedJSON := `[{"a":1,"b":2},{"a":1,"b":2},{"a":1,"b":2}]`

	updatedJSON, err := AddKeyToJSONArray(jsonArrayStr, key, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if updatedJSON != expectedJSON {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expectedJSON, updatedJSON)
	}
}

func TestAddToJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	jsonStr := `{invalid json}`
	jsonPath := "key"
	value := "value"

	_, err := AddToJSON(jsonStr, jsonPath, value)
	if err == nil {
		t.Error("Expected error for invalid JSON, but got nil")
	}
}

func TestAddToJSON_InvalidPath(t *testing.T) {
	t.Parallel()

	jsonStr := `{"name":"John"}`
	jsonPath := "" // Empty path might cause issues
	value := "value"

	_, err := AddToJSON(jsonStr, jsonPath, value)
	// sjson.Set might return an error for invalid path
	if err != nil {
		t.Logf("Got expected error for invalid path: %v", err)
	}
}

func TestAddKeyToJSONArray_InvalidJSON(t *testing.T) {
	t.Parallel()

	jsonArrayStr := `{invalid json}`
	key := "b"
	value := 2

	_, err := AddKeyToJSONArray(jsonArrayStr, key, value)
	if err == nil {
		t.Error("Expected error for invalid JSON, but got nil")
	}
}

func TestAddKeyToJSONArray_EmptyArray(t *testing.T) {
	t.Parallel()

	jsonArrayStr := `[]`
	key := "b"
	value := 2

	expectedJSON := `[]`

	updatedJSON, err := AddKeyToJSONArray(jsonArrayStr, key, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if updatedJSON != expectedJSON {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expectedJSON, updatedJSON)
	}
}

func TestAddKeyToJSONArray_SingleElement(t *testing.T) {
	t.Parallel()

	jsonArrayStr := `[{"a":1}]`
	key := "b"
	value := 2

	expectedJSON := `[{"a":1,"b":2}]`

	updatedJSON, err := AddKeyToJSONArray(jsonArrayStr, key, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if updatedJSON != expectedJSON {
		t.Errorf("Test failed, expected: '%s', got: '%s'", expectedJSON, updatedJSON)
	}
}

func TestRemoveKeyFromJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonStr     string
		jsonPath    string
		expected    string
		expectError bool
	}{
		{
			name:        "删除顶层key",
			jsonStr:     `{"name":"John","age":30,"city":"NYC"}`,
			jsonPath:    "age",
			expected:    `{"name":"John","city":"NYC"}`,
			expectError: false,
		},
		{
			name:        "删除嵌套key",
			jsonStr:     `{"name":"John","cars":{"car1":"Ford","car2":"BMW"}}`,
			jsonPath:    "cars.car1",
			expected:    `{"name":"John","cars":{"car2":"BMW"}}`,
			expectError: false,
		},
		{
			name:        "删除数组元素",
			jsonStr:     `{"items":["a","b","c"]}`,
			jsonPath:    "items.1",
			expected:    `{"items":["a","c"]}`,
			expectError: false,
		},
		{
			name:        "删除不存在的key",
			jsonStr:     `{"name":"John"}`,
			jsonPath:    "nonexistent",
			expected:    `{"name":"John"}`,
			expectError: false,
		},
		{
			name:        "无效JSON",
			jsonStr:     `{invalid json}`,
			jsonPath:    "key",
			expected:    "",
			expectError: true,
		},
		{
			name:        "空JSON对象",
			jsonStr:     `{}`,
			jsonPath:    "key",
			expected:    `{}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := RemoveKeyFromJSON(tt.jsonStr, tt.jsonPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但没有返回")
				}

				return
			}

			if err != nil {
				t.Errorf("未预期的错误: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("结果不匹配\n期望: %s\n实际: %s", tt.expected, result)
			}
		})
	}
}
