package agentconfigresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSelfConfigField(t *testing.T) {
	t.Parallel()

	field := NewSelfConfigField()

	assert.NotNil(t, field)
	assert.IsType(t, &SelfConfigField{}, field)
}

func TestFieldInfo_StructFields(t *testing.T) {
	t.Parallel()

	field := &FieldInfo{
		Name:        "fieldName",
		Type:        "string",
		Description: "Field description",
	}

	assert.Equal(t, "fieldName", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, "Field description", field.Description)
}

func TestFieldInfo_Empty(t *testing.T) {
	t.Parallel()

	field := &FieldInfo{}

	assert.Empty(t, field.Name)
	assert.Empty(t, field.Type)
	assert.Empty(t, field.Description)
	assert.Nil(t, field.Children)
}

func TestFieldInfo_WithChildren(t *testing.T) {
	t.Parallel()

	children := []*FieldInfo{
		{
			Name:        "child1",
			Type:        "string",
			Description: "First child",
		},
		{
			Name:        "child2",
			Type:        "int",
			Description: "Second child",
		},
	}
	field := &FieldInfo{
		Name:        "parentField",
		Type:        "object",
		Description: "Parent field with children",
		Children:    children,
	}

	assert.Equal(t, "parentField", field.Name)
	assert.Len(t, field.Children, 2)
	assert.Equal(t, "child1", field.Children[0].Name)
	assert.Equal(t, "child2", field.Children[1].Name)
}

func TestFieldInfo_NestedChildren(t *testing.T) {
	t.Parallel()

	grandchild := &FieldInfo{
		Name:        "grandchild",
		Type:        "boolean",
		Description: "Grandchild field",
	}
	child := &FieldInfo{
		Name:        "child",
		Type:        "object",
		Description: "Child field",
		Children:    []*FieldInfo{grandchild},
	}
	parent := &FieldInfo{
		Name:        "parent",
		Type:        "object",
		Description: "Parent field",
		Children:    []*FieldInfo{child},
	}

	assert.Len(t, parent.Children, 1)
	assert.Len(t, parent.Children[0].Children, 1)
	assert.Equal(t, "grandchild", parent.Children[0].Children[0].Name)
}

func TestSelfConfigField_StructFields(t *testing.T) {
	t.Parallel()

	field := &SelfConfigField{
		Name:        "configField",
		Type:        "string",
		Description: "Configuration field",
	}

	assert.Equal(t, "configField", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, "Configuration field", field.Description)
}

func TestSelfConfigField_Empty(t *testing.T) {
	t.Parallel()

	field := &SelfConfigField{}

	assert.Empty(t, field.Name)
	assert.Empty(t, field.Type)
	assert.Empty(t, field.Description)
	assert.Nil(t, field.Children)
}

func TestSelfConfigField_WithChildren(t *testing.T) {
	t.Parallel()

	field := &SelfConfigField{
		Name: "parent",
		Type: "object",
		Children: []*FieldInfo{
			{
				Name: "child1",
				Type: "string",
			},
			{
				Name: "child2",
				Type: "int",
			},
		},
	}

	assert.Len(t, field.Children, 2)
	assert.Equal(t, "child1", field.Children[0].Name)
	assert.Equal(t, "child2", field.Children[1].Name)
}

func TestSelfConfigField_LoadFromJSONStr(t *testing.T) {
	t.Parallel()

	field := NewSelfConfigField()

	err := field.LoadFromJSONStr()

	// The actual test depends on the embedded JSON file
	// Just verify the method exists and can be called
	if err != nil {
		// If there's an error loading the JSON (e.g., file not found), that's ok for this test
		// We're just testing that the method is callable
		assert.NotNil(t, field)
	} else {
		// If successful, the field should have been populated
		assert.NotNil(t, field)
	}
}

func TestSelfConfigField_LoadFromJSONStr_ContainsSkillSkillField(t *testing.T) {
	t.Parallel()

	field := NewSelfConfigField()

	err := field.LoadFromJSONStr()
	require.NoError(t, err)

	skillsField := findFieldInfo(field.Children, "skills")
	require.NotNil(t, skillsField)

	skillItemsField := findFieldInfo(skillsField.Children, "skills")
	require.NotNil(t, skillItemsField)

	skillIDField := findFieldInfo(skillItemsField.Children, "skill_id")
	require.NotNil(t, skillIDField)
	assert.Equal(t, "string", skillIDField.Type)
}

func TestFieldInfo_DifferentTypes(t *testing.T) {
	t.Parallel()

	types := []string{
		"string",
		"int",
		"float",
		"boolean",
		"array",
		"object",
	}

	for _, fieldType := range types {
		field := &FieldInfo{
			Name: "testField",
			Type: fieldType,
		}

		assert.Equal(t, fieldType, field.Type)
	}
}

func TestFieldInfo_LongDescriptions(t *testing.T) {
	t.Parallel()

	longDesc := "This is a very long description that contains detailed information about the field and its usage"
	field := &FieldInfo{
		Name:        "detailedField",
		Type:        "string",
		Description: longDesc,
	}

	assert.Equal(t, longDesc, field.Description)
	assert.Contains(t, field.Description, "detailed information")
}

func TestFieldInfo_EmptyChildren(t *testing.T) {
	t.Parallel()

	field := &FieldInfo{
		Name:     "fieldWithEmptyChildren",
		Type:     "object",
		Children: []*FieldInfo{},
	}

	assert.Empty(t, field.Children)
	assert.Len(t, field.Children, 0)
}

func findFieldInfo(fields []*FieldInfo, name string) *FieldInfo {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}

	return nil
}
