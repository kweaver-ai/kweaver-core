package observabilitysvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

func TestSafeGetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{
			name:     "existing string value",
			m:        map[string]any{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "non-existent key",
			m:        map[string]any{"other": "value"},
			key:      "key",
			expected: "",
		},
		{
			name:     "nil value",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: "",
		},
		{
			name:     "non-string value",
			m:        map[string]any{"key": 123},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeGetString(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeGetBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected bool
	}{
		{
			name:     "true value",
			m:        map[string]any{"key": true},
			key:      "key",
			expected: true,
		},
		{
			name:     "false value",
			m:        map[string]any{"key": false},
			key:      "key",
			expected: false,
		},
		{
			name:     "non-existent key",
			m:        map[string]any{"other": true},
			key:      "key",
			expected: false,
		},
		{
			name:     "nil value",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: false,
		},
		{
			name:     "non-bool value",
			m:        map[string]any{"key": "true"},
			key:      "key",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeGetBool(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeGetFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected float64
	}{
		{
			name:     "existing float value",
			m:        map[string]any{"key": 123.45},
			key:      "key",
			expected: 123.45,
		},
		{
			name:     "zero value",
			m:        map[string]any{"key": 0.0},
			key:      "key",
			expected: 0.0,
		},
		{
			name:     "non-existent key",
			m:        map[string]any{"other": 123.45},
			key:      "key",
			expected: 0.0,
		},
		{
			name:     "nil value",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: 0.0,
		},
		{
			name:     "non-float value",
			m:        map[string]any{"key": "123.45"},
			key:      "key",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeGetFloat64(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeGetInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected int64
	}{
		{
			name:     "existing float value that fits in int64",
			m:        map[string]any{"key": 123.0},
			key:      "key",
			expected: 123,
		},
		{
			name:     "zero value",
			m:        map[string]any{"key": 0.0},
			key:      "key",
			expected: 0,
		},
		{
			name:     "non-existent key",
			m:        map[string]any{"other": 123.0},
			key:      "key",
			expected: 0,
		},
		{
			name:     "nil value",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeGetInt64(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeParseSkillInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     any
		expected *agentrespvo.SkillInfo
	}{
		{
			name: "valid skill info",
			data: map[string]any{
				"type":    "tool",
				"name":    "search",
				"checked": true,
				"args": []any{
					map[string]any{
						"name":  "query",
						"value": "test",
						"type":  "string",
					},
				},
			},
			expected: &agentrespvo.SkillInfo{
				Type:    "tool",
				Name:    "search",
				Checked: true,
				Args: []agentrespvo.Arg{
					{
						Name:  "query",
						Value: "test",
						Type:  "string",
					},
				},
			},
		},
		{
			name:     "nil data",
			data:     nil,
			expected: nil,
		},
		{
			name:     "invalid data type",
			data:     "string",
			expected: nil,
		},
		{
			name:     "empty map",
			data:     map[string]any{},
			expected: &agentrespvo.SkillInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeParseSkillInfo(tt.data)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Checked, result.Checked)
			}
		})
	}
}

func TestSafeParseTokenUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     any
		expected agentrespvo.TokenUsage
	}{
		{
			name: "valid token usage",
			data: map[string]any{
				"prompt_tokens":     100.0,
				"completion_tokens": 50.0,
				"total_tokens":      150.0,
			},
			expected: agentrespvo.TokenUsage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		},
		{
			name:     "nil data",
			data:     nil,
			expected: agentrespvo.TokenUsage{},
		},
		{
			name:     "invalid data type",
			data:     "string",
			expected: agentrespvo.TokenUsage{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := safeParseTokenUsage(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTimeToISO8601(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		timestamp int64
		wantEmpty bool
	}{
		{
			name:      "zero timestamp",
			timestamp: 0,
			wantEmpty: true,
		},
		{
			name:      "valid timestamp",
			timestamp: 1609459200000, // 2021-01-01 00:00:00 UTC in milliseconds
			wantEmpty: false,
		},
		{
			name:      "negative timestamp",
			timestamp: -1000,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatTimeToISO8601(tt.timestamp)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "T")
			}
		})
	}
}

func TestSafeParseArgs(t *testing.T) {
	t.Parallel()

	t.Run("nil data returns nil", func(t *testing.T) {
		t.Parallel()

		result := safeParseArgs(nil)
		assert.Nil(t, result)
	})

	t.Run("empty array returns nil", func(t *testing.T) {
		t.Parallel()

		result := safeParseArgs([]any{})
		assert.Nil(t, result)
	})

	t.Run("valid args array", func(t *testing.T) {
		t.Parallel()

		data := []any{
			map[string]any{
				"name":  "arg1",
				"value": "value1",
				"type":  "string",
			},
			map[string]any{
				"name":  "arg2",
				"value": 123,
				"type":  "int",
			},
		}
		result := safeParseArgs(data)

		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "arg1", result[0].Name)
		assert.Equal(t, "value1", result[0].Value)
		assert.Equal(t, "string", result[0].Type)
		assert.Equal(t, "arg2", result[1].Name)
	})

	t.Run("args with missing fields", func(t *testing.T) {
		t.Parallel()

		data := []any{
			map[string]any{
				"name": "arg1",
			},
		}
		result := safeParseArgs(data)

		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "arg1", result[0].Name)
	})

	t.Run("non-array data returns nil", func(t *testing.T) {
		t.Parallel()

		result := safeParseArgs("string")
		assert.Nil(t, result)
	})
}

func TestSafeParsePromptTokenDetails(t *testing.T) {
	t.Parallel()

	t.Run("nil data returns empty struct", func(t *testing.T) {
		t.Parallel()

		result := safeParsePromptTokenDetails(nil)
		assert.Equal(t, agentrespvo.PromptTokenDetails{}, result)
	})

	t.Run("valid data", func(t *testing.T) {
		t.Parallel()

		data := map[string]any{
			"cached_tokens":   100.0,
			"uncached_tokens": 200.0,
		}
		result := safeParsePromptTokenDetails(data)

		assert.Equal(t, int64(100), result.CachedTokens)
		assert.Equal(t, int64(200), result.UncachedTokens)
	})

	t.Run("data with missing fields", func(t *testing.T) {
		t.Parallel()

		data := map[string]any{
			"cached_tokens": 50.0,
		}
		result := safeParsePromptTokenDetails(data)

		assert.Equal(t, int64(50), result.CachedTokens)
		assert.Equal(t, int64(0), result.UncachedTokens)
	})

	t.Run("non-map data returns empty struct", func(t *testing.T) {
		t.Parallel()

		result := safeParsePromptTokenDetails("string")
		assert.Equal(t, agentrespvo.PromptTokenDetails{}, result)
	})
}

func TestSafeParseSkillInfo_WithComplexArgs(t *testing.T) {
	t.Parallel()

	t.Run("skill info with multiple args", func(t *testing.T) {
		t.Parallel()

		data := map[string]any{
			"type":    "tool",
			"name":    "multi_tool",
			"checked": false,
			"args": []any{
				map[string]any{
					"name":  "param1",
					"value": "value1",
					"type":  "string",
				},
				map[string]any{
					"name":  "param2",
					"value": "42",
					"type":  "number",
				},
				map[string]any{
					"name":  "param3",
					"value": true,
					"type":  "boolean",
				},
			},
		}
		result := safeParseSkillInfo(data)

		assert.NotNil(t, result)
		assert.Equal(t, "tool", result.Type)
		assert.Equal(t, "multi_tool", result.Name)
		assert.False(t, result.Checked)
		assert.Len(t, result.Args, 3)
		assert.Equal(t, "param1", result.Args[0].Name)
		assert.Equal(t, "value1", result.Args[0].Value)
		assert.Equal(t, "string", result.Args[0].Type)
	})
}

func TestSafeGetInt64_LargeValues(t *testing.T) {
	t.Parallel()

	t.Run("large float value", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"key": 9007199254740991.0}
		result := safeGetInt64(m, "key")
		assert.Equal(t, int64(9007199254740991), result)
	})

	t.Run("negative float value", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"key": -123.45}
		result := safeGetInt64(m, "key")
		assert.Equal(t, int64(-123), result)
	})
}
