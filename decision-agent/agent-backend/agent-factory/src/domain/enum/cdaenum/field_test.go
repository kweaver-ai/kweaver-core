package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputFieldType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, InputFieldType("string"), InputFieldTypeString)
	assert.Equal(t, InputFieldType("file"), InputFieldTypeFile)
	assert.Equal(t, InputFieldType("object"), InputFieldTypeJSONObject)
}

func TestInputFieldType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []InputFieldType{
		InputFieldTypeString,
		InputFieldTypeFile,
		InputFieldTypeJSONObject,
	}

	for _, fieldType := range validTypes {
		t.Run(string(fieldType), func(t *testing.T) {
			t.Parallel()

			err := fieldType.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestInputFieldType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := InputFieldType("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input field type")
}

func TestInputFieldType_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := InputFieldType("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input field type")
}

func TestInputFieldType_AllUnique(t *testing.T) {
	t.Parallel()

	fieldTypes := []InputFieldType{
		InputFieldTypeString,
		InputFieldTypeFile,
		InputFieldTypeJSONObject,
	}

	uniqueTypes := make(map[InputFieldType]bool)
	for _, ft := range fieldTypes {
		assert.False(t, uniqueTypes[ft], "Duplicate field type found: %s", ft)
		uniqueTypes[ft] = true
	}
}

func TestInputFieldType_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldType InputFieldType
		expected  string
	}{
		{
			name:      "string field type",
			fieldType: InputFieldTypeString,
			expected:  "string",
		},
		{
			name:      "file field type",
			fieldType: InputFieldTypeFile,
			expected:  "file",
		},
		{
			name:      "json object field type",
			fieldType: InputFieldTypeJSONObject,
			expected:  "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
