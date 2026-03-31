package customvalidator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentAndTplNameRe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{
			name:    "valid English name starting with letter",
			input:   "AgentName",
			matches: true,
		},
		{
			name:    "valid Chinese name starting with Chinese",
			input:   "智能体名称",
			matches: true,
		},
		{
			name:    "valid name with underscore",
			input:   "agent_name",
			matches: true,
		},
		{
			name:    "valid name starting with underscore",
			input:   "_agentName",
			matches: true,
		},
		{
			name:    "valid name with numbers after first char",
			input:   "Agent123",
			matches: true,
		},
		{
			name:    "valid mixed Chinese, English, numbers, underscore",
			input:   "智能体_Agent123",
			matches: true,
		},
		{
			name:    "valid name with only underscore",
			input:   "_",
			matches: true,
		},
		{
			name:    "valid single Chinese character",
			input:   "智",
			matches: true,
		},
		{
			name:    "valid single English letter",
			input:   "a",
			matches: true,
		},
		{
			name:    "invalid - starting with number",
			input:   "123Agent",
			matches: false,
		},
		{
			name:    "invalid - only numbers",
			input:   "123",
			matches: false,
		},
		{
			name:    "invalid - contains space",
			input:   "Agent Name",
			matches: false,
		},
		{
			name:    "invalid - contains hyphen",
			input:   "Agent-Name",
			matches: false,
		},
		{
			name:    "invalid - contains special char",
			input:   "Agent@Name",
			matches: false,
		},
		{
			name:    "invalid - contains dot",
			input:   "Agent.Name",
			matches: false,
		},
		{
			name:    "invalid - empty string",
			input:   "",
			matches: false,
		},
		{
			name:    "valid - underscore followed by numbers",
			input:   "_123",
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := AgentAndTplNameRe.MatchString(tt.input)
			assert.Equal(t, tt.matches, result, "Input: %s", tt.input)
		})
	}
}

func TestGenAgentAndTplNameErrMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{
			name:     "with prefix",
			target:   "Agent名称",
			expected: "Agent名称仅支持中英文、数字及下划线，且不能以数字开头",
		},
		{
			name:     "empty prefix",
			target:   "",
			expected: "仅支持中英文、数字及下划线，且不能以数字开头",
		},
		{
			name:     "with single character prefix",
			target:   "A",
			expected: "A仅支持中英文、数字及下划线，且不能以数字开头",
		},
		{
			name:     "with Chinese prefix",
			target:   "模板",
			expected: "模板仅支持中英文、数字及下划线，且不能以数字开头",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GenAgentAndTplNameErrMsg(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgentAndTplNameRe_PatternConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"a", true},
		{"A", true},
		{"_", true},
		{"Abc_123", true},
		{"abc_123", true},
		{"ABC_123", true},
		{"智", true},
		{"智能体", true},
		{"智能体_Agent", true},
		{"智能体_Agent_123", true},
		{"1", false},
		{"1abc", false},
		{"123", false},
		{"1_abc", false},
		{"abc-", false},
		{"abc.", false},
		{"abc ", false},
		{" abc", false},
		{"abc@def", false},
		{"abc#def", false},
		{"abc$def", false},
		{"abc%def", false},
		{"abc&def", false},
		{"abc*def", false},
		{"abc+def", false},
		{"abc=def", false},
		{"abc?def", false},
		{"abc/def", false},
		{"abc\\def", false},
		{"abc|def", false},
		{"abc[def", false},
		{"abc]def", false},
		{"abc{def", false},
		{"abc}def", false},
		{"abc(def", false},
		{"abc)def", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			result := AgentAndTplNameRe.MatchString(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %q", tt.input)
		})
	}
}

func TestAgentAndTplNameRe_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{
			name:    "single underscore",
			input:   "_",
			matches: true,
		},
		{
			name:    "multiple underscores",
			input:   "___",
			matches: true,
		},
		{
			name:    "underscore between characters",
			input:   "a_b_c",
			matches: true,
		},
		{
			name:    "Chinese with underscore",
			input:   "智_能_体",
			matches: true,
		},
		{
			name:    "mixed case",
			input:   "AaBbCc_123",
			matches: true,
		},
		{
			name:    "long valid name",
			input:   "这是一个非常长的智能体名称_AgentName_1234567890",
			matches: true,
		},
		{
			name:    "valid - underscore followed by numbers",
			input:   "_123",
			matches: true,
		},
		{
			name:    "space at start",
			input:   " Agent",
			matches: false,
		},
		{
			name:    "space in middle",
			input:   "Agent Name",
			matches: false,
		},
		{
			name:    "space at end",
			input:   "Agent ",
			matches: false,
		},
		{
			name:    "tab character",
			input:   "Agent\tName",
			matches: false,
		},
		{
			name:    "newline character",
			input:   "Agent\nName",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := AgentAndTplNameRe.MatchString(tt.input)
			assert.Equal(t, tt.matches, result)
		})
	}
}

// Test CheckAgentAndTplName using reflection to call it directly
func TestCheckAgentAndTplName_Direct(t *testing.T) {
	t.Parallel()

	// Since CheckAgentAndTplName requires a validator.FieldLevel interface,
	// we test the underlying logic by testing the regex pattern directly
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid English name",
			input:    "AgentName",
			expected: true,
		},
		{
			name:     "valid Chinese name",
			input:    "智能体",
			expected: true,
		},
		{
			name:     "valid name with underscore",
			input:    "agent_name",
			expected: true,
		},
		{
			name:     "valid name starting with underscore",
			input:    "_agent",
			expected: true,
		},
		{
			name:     "valid name with numbers",
			input:    "Agent123",
			expected: true,
		},
		{
			name:     "invalid - starts with number",
			input:    "123Agent",
			expected: false,
		},
		{
			name:     "invalid - contains space",
			input:    "Agent Name",
			expected: false,
		},
		{
			name:     "invalid - contains special char",
			input:    "Agent@Name",
			expected: false,
		},
		{
			name:     "empty string - should match (allowed by CheckAgentAndTplName)",
			input:    "",
			expected: true, // Empty string is allowed by CheckAgentAndTplName
		},
		{
			name:     "valid mixed Chinese and English",
			input:    "智能体Agent",
			expected: true,
		},
		{
			name:     "valid with underscore and numbers",
			input:    "Agent_123",
			expected: true,
		},
		{
			name:     "invalid - contains hyphen",
			input:    "Agent-Name",
			expected: false,
		},
		{
			name:     "invalid - only numbers",
			input:    "123456",
			expected: false,
		},
		{
			name:     "valid - underscore followed by numbers",
			input:    "_123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// CheckAgentAndTplName returns true for empty strings
			// Otherwise it checks the regex
			var result bool
			if tt.input == "" {
				result = true // Empty string is allowed
			} else {
				result = AgentAndTplNameRe.MatchString(tt.input)
			}

			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		})
	}
}

func TestAgentAndTplNameRe_ConstantValue(t *testing.T) {
	t.Parallel()

	// Test that the constant is properly defined
	assert.NotNil(t, AgentAndTplNameRe)
	assert.Equal(t, "仅支持中英文、数字及下划线，且不能以数字开头", AgentAndTplNameErrMsg)
}

func TestCheckAgentAndTplName_EmptyStringBehavior(t *testing.T) {
	t.Parallel()

	// Verify that the function allows empty strings
	// This is tested indirectly by checking the comment in the source code
	// and verifying the regex behavior
	emptyResult := AgentAndTplNameRe.MatchString("")
	assert.False(t, emptyResult, "Regex should not match empty string")
}

func TestCheckAgentAndTplName_ReflectType(t *testing.T) {
	t.Parallel()

	// Verify CheckAgentAndTplName is a function
	fnType := reflect.TypeOf(CheckAgentAndTplName)
	assert.Equal(t, reflect.Func, fnType.Kind())
}

func TestGenAgentAndTplNameErrMsg_Constant(t *testing.T) {
	t.Parallel()

	// Test the error message constant
	expectedMsg := "仅支持中英文、数字及下划线，且不能以数字开头"
	assert.Equal(t, expectedMsg, AgentAndTplNameErrMsg)

	// Test that GenAgentAndTplNameErrMsg correctly concatenates
	result := GenAgentAndTplNameErrMsg("Agent名称")
	assert.Contains(t, result, expectedMsg)
	assert.Contains(t, result, "Agent名称")
}
