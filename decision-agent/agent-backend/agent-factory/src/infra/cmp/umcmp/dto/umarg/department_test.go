package umarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeptInfoField_StringRepresentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    DeptInfoField
		expected string
	}{
		{DeptFieldName, "name"},
		{DeptFieldParentDeps, "parent_deps"},
		{DeptFieldManagers, "managers"},
	}

	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.field))
		})
	}
}

func TestDeptInfoField_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DeptInfoField("name"), DeptFieldName)
	assert.Equal(t, DeptInfoField("parent_deps"), DeptFieldParentDeps)
	assert.Equal(t, DeptInfoField("managers"), DeptFieldManagers)
}

func TestDeptFields_ToStrings(t *testing.T) {
	t.Parallel()

	fields := DeptFields{
		DeptFieldName,
		DeptFieldParentDeps,
		DeptFieldManagers,
	}
	result := fields.ToStrings()

	assert.Len(t, result, 3)
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "parent_deps", result[1])
	assert.Equal(t, "managers", result[2])
}

func TestDeptFields_ToStrings_Empty(t *testing.T) {
	t.Parallel()

	fields := DeptFields{}
	result := fields.ToStrings()

	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestDeptFields_ToStrings_SingleField(t *testing.T) {
	t.Parallel()

	fields := DeptFields{
		DeptFieldName,
	}
	result := fields.ToStrings()

	assert.Len(t, result, 1)
	assert.Equal(t, "name", result[0])
}

func TestDeptFields_ToStrings_DuplicateFields(t *testing.T) {
	t.Parallel()

	fields := DeptFields{
		DeptFieldName,
		DeptFieldName,
		DeptFieldManagers,
	}
	result := fields.ToStrings()

	assert.Len(t, result, 3)
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "name", result[1])
	assert.Equal(t, "managers", result[2])
}

func TestGetDeptInfoArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"dept-1", "dept-2"},
		Fields:  DeptFields{DeptFieldName, DeptFieldManagers},
	}

	assert.Len(t, dto.DeptIds, 2)
	assert.Len(t, dto.Fields, 2)
	assert.Equal(t, "dept-1", dto.DeptIds[0])
	assert.Equal(t, "name", string(dto.Fields[0]))
}

func TestGetDeptInfoArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{}

	assert.Nil(t, dto.DeptIds)
	assert.Nil(t, dto.Fields)
}

func TestGetDeptInfoArgDto_WithOnlyDeptIds(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"dept-1", "dept-2", "dept-3"},
	}

	assert.Len(t, dto.DeptIds, 3)
	assert.Nil(t, dto.Fields)
}

func TestGetDeptInfoArgDto_WithOnlyFields(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		Fields: DeptFields{DeptFieldName, DeptFieldParentDeps},
	}

	assert.Nil(t, dto.DeptIds)
	assert.Len(t, dto.Fields, 2)
}

func TestGetDeptInfoArgDto_WithChineseDeptIds(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"部门-1", "部门-2"},
	}

	assert.Equal(t, "部门-1", dto.DeptIds[0])
	assert.Equal(t, "部门-2", dto.DeptIds[1])
}

func TestGetDeptInfoArgDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{},
		Fields:  DeptFields{},
	}

	assert.NotNil(t, dto.DeptIds)
	assert.NotNil(t, dto.Fields)
	assert.Len(t, dto.DeptIds, 0)
	assert.Len(t, dto.Fields, 0)
}

func TestGetDeptInfoArgDto_AppendDeptIds(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{}
	dto.DeptIds = append(dto.DeptIds, "dept-1")
	dto.DeptIds = append(dto.DeptIds, "dept-2")

	assert.Len(t, dto.DeptIds, 2)
	assert.Equal(t, "dept-1", dto.DeptIds[0])
	assert.Equal(t, "dept-2", dto.DeptIds[1])
}

func TestGetDeptInfoArgDto_AppendFields(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{}
	dto.Fields = append(dto.Fields, DeptFieldName)
	dto.Fields = append(dto.Fields, DeptFieldManagers)

	assert.Len(t, dto.Fields, 2)
	assert.Equal(t, DeptFieldName, dto.Fields[0])
	assert.Equal(t, DeptFieldManagers, dto.Fields[1])
}

func TestGetDeptInfoArgDto_WithMultipleDeptIds(t *testing.T) {
	t.Parallel()

	deptIds := make([]string, 50)
	for i := 0; i < 50; i++ {
		deptIds[i] = "dept-" + string(rune(i))
	}

	dto := GetDeptInfoArgDto{
		DeptIds: deptIds,
	}

	assert.Len(t, dto.DeptIds, 50)
}

func TestGetDeptInfoArgDto_WithAllFields(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"dept-1"},
		Fields: DeptFields{
			DeptFieldName,
			DeptFieldParentDeps,
			DeptFieldManagers,
		},
	}

	result := dto.Fields.ToStrings()
	assert.Len(t, result, 3)
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "parent_deps", result[1])
	assert.Equal(t, "managers", result[2])
}

func TestGetDeptInfoArgDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"dept-1", "dept-2", "dept-3"},
		Fields:  DeptFields{DeptFieldName, DeptFieldManagers},
	}

	count := 0

	for _, deptId := range dto.DeptIds {
		assert.NotEmpty(t, deptId)

		count++
	}

	assert.Equal(t, 3, count)

	fieldCount := 0

	for _, field := range dto.Fields {
		assert.NotEmpty(t, string(field))

		fieldCount++
	}

	assert.Equal(t, 2, fieldCount)
}

func TestDeptFields_SliceOperations(t *testing.T) {
	t.Parallel()

	fields := DeptFields{
		DeptFieldName,
		DeptFieldParentDeps,
		DeptFieldManagers,
	}

	// Test slicing
	subFields := fields[1:3]
	assert.Len(t, subFields, 2)
	assert.Equal(t, DeptFieldParentDeps, subFields[0])
}

func TestGetDeptInfoArgDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := GetDeptInfoArgDto{
		DeptIds: []string{"dept-1", "dept-2", "dept-3"},
	}

	// Test slicing
	subDeptIds := dto.DeptIds[1:3]
	assert.Len(t, subDeptIds, 2)
	assert.Equal(t, "dept-2", subDeptIds[0])
}

func TestDeptFields_AllFieldTypes(t *testing.T) {
	t.Parallel()

	allFields := []DeptInfoField{
		DeptFieldName,
		DeptFieldParentDeps,
		DeptFieldManagers,
	}

	fields := DeptFields(allFields)
	result := fields.ToStrings()

	assert.Len(t, result, 3)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "parent_deps")
	assert.Contains(t, result, "managers")
}
