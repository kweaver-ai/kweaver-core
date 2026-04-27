package efastarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIbField_StringRepresentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field    IbField
		expected string
	}{
		{IbFieldName, "names"},
		{IbFieldDocLibTypes, "doc_lib_types"},
		{IbFieldPaths, "paths"},
	}

	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.field))
		})
	}
}

func TestIbField_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, IbField("names"), IbFieldName)
	assert.Equal(t, IbField("doc_lib_types"), IbFieldDocLibTypes)
	assert.Equal(t, IbField("paths"), IbFieldPaths)
}

func TestIbFields_ToPathString(t *testing.T) {
	t.Parallel()

	fields := IbFields{
		IbFieldName,
		IbFieldDocLibTypes,
		IbFieldPaths,
	}
	result := fields.ToPathString()

	assert.Equal(t, "names,doc_lib_types,paths", result)
}

func TestIbFields_ToPathString_Empty(t *testing.T) {
	t.Parallel()

	fields := IbFields{}
	result := fields.ToPathString()

	assert.Equal(t, "", result)
}

func TestIbFields_ToPathString_SingleField(t *testing.T) {
	t.Parallel()

	fields := IbFields{
		IbFieldName,
	}
	result := fields.ToPathString()

	assert.Equal(t, "names", result)
}

func TestIbFields_ToPathString_DuplicateFields(t *testing.T) {
	t.Parallel()

	fields := IbFields{
		IbFieldName,
		IbFieldName,
		IbFieldPaths,
	}
	result := fields.ToPathString()

	assert.Equal(t, "names,names,paths", result)
}

func TestGetFsMetadataArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs:    []string{"id-1", "id-2"},
		ObjIDs: []string{"obj-1"},
		Fields: IbFields{IbFieldName, IbFieldPaths},
	}

	assert.Len(t, dto.IDs, 2)
	assert.Len(t, dto.ObjIDs, 1)
	assert.Len(t, dto.Fields, 2)
	assert.Equal(t, "id-1", dto.IDs[0])
	assert.Equal(t, "obj-1", dto.ObjIDs[0])
	assert.Equal(t, IbFieldName, dto.Fields[0])
}

func TestGetFsMetadataArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{}

	assert.Nil(t, dto.IDs)
	assert.Nil(t, dto.ObjIDs)
	assert.Nil(t, dto.Fields)
}

func TestGetFsMetadataArgDto_WithOnlyIDs(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs: []string{"id-1", "id-2", "id-3"},
	}

	assert.Len(t, dto.IDs, 3)
	assert.Nil(t, dto.ObjIDs)
	assert.Nil(t, dto.Fields)
}

func TestGetFsMetadataArgDto_WithOnlyObjIDs(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		ObjIDs: []string{"obj-1", "obj-2"},
	}

	assert.Nil(t, dto.IDs)
	assert.Len(t, dto.ObjIDs, 2)
	assert.Nil(t, dto.Fields)
}

func TestGetFsMetadataArgDto_WithOnlyFields(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		Fields: IbFields{IbFieldName, IbFieldDocLibTypes},
	}

	assert.Nil(t, dto.IDs)
	assert.Nil(t, dto.ObjIDs)
	assert.Len(t, dto.Fields, 2)
}

func TestGetFsMetadataArgDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs:    []string{"ID-1", "ID-2"},
		ObjIDs: []string{"对象-1"},
	}

	assert.Equal(t, "ID-1", dto.IDs[0])
	assert.Equal(t, "对象-1", dto.ObjIDs[0])
}

func TestGetFsMetadataArgDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs:    []string{},
		ObjIDs: []string{},
		Fields: IbFields{},
	}

	assert.NotNil(t, dto.IDs)
	assert.NotNil(t, dto.ObjIDs)
	assert.NotNil(t, dto.Fields)
	assert.Len(t, dto.IDs, 0)
	assert.Len(t, dto.ObjIDs, 0)
	assert.Len(t, dto.Fields, 0)
}

func TestGetFsMetadataArgDto_AppendIDs(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{}
	dto.IDs = append(dto.IDs, "id-1")
	dto.IDs = append(dto.IDs, "id-2")

	assert.Len(t, dto.IDs, 2)
	assert.Equal(t, "id-1", dto.IDs[0])
	assert.Equal(t, "id-2", dto.IDs[1])
}

func TestGetFsMetadataArgDto_AppendObjIDs(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{}
	dto.ObjIDs = append(dto.ObjIDs, "obj-1")
	dto.ObjIDs = append(dto.ObjIDs, "obj-2")

	assert.Len(t, dto.ObjIDs, 2)
	assert.Equal(t, "obj-1", dto.ObjIDs[0])
	assert.Equal(t, "obj-2", dto.ObjIDs[1])
}

func TestGetFsMetadataArgDto_AppendFields(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{}
	dto.Fields = append(dto.Fields, IbFieldName)
	dto.Fields = append(dto.Fields, IbFieldPaths)

	assert.Len(t, dto.Fields, 2)
	assert.Equal(t, IbFieldName, dto.Fields[0])
	assert.Equal(t, IbFieldPaths, dto.Fields[1])
}

func TestGetFsMetadataArgDto_WithMultipleIDs(t *testing.T) {
	t.Parallel()

	ids := make([]string, 50)
	for i := 0; i < 50; i++ {
		ids[i] = "id-" + string(rune(i))
	}

	dto := GetFsMetadataArgDto{
		IDs: ids,
	}

	assert.Len(t, dto.IDs, 50)
}

func TestGetFsMetadataArgDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs:    []string{"id-1", "id-2"},
		ObjIDs: []string{"obj-1"},
		Fields: IbFields{IbFieldName, IbFieldPaths},
	}

	count := 0

	for _, id := range dto.IDs {
		assert.NotEmpty(t, id)

		count++
	}

	assert.Equal(t, 2, count)

	objCount := 0

	for _, objID := range dto.ObjIDs {
		assert.NotEmpty(t, objID)

		objCount++
	}

	assert.Equal(t, 1, objCount)

	fieldCount := 0

	for _, field := range dto.Fields {
		assert.NotEmpty(t, string(field))

		fieldCount++
	}

	assert.Equal(t, 2, fieldCount)
}

func TestNewGetFsMetadataEFArgDto_WithIDs(t *testing.T) {
	t.Parallel()

	argDto := &GetFsMetadataArgDto{
		IDs: []string{"id-1", "id-2", "id-1"},
	}

	efDto := NewGetFsMetadataEFArgDto(argDto)

	assert.NotNil(t, efDto)
	assert.Equal(t, "GET", efDto.Method)
	assert.NotNil(t, efDto.IDs)
	// Should be deduplicated
	assert.Len(t, efDto.IDs, 2)
	assert.Nil(t, efDto.ObjIDs)
}

func TestNewGetFsMetadataEFArgDto_WithObjIDs(t *testing.T) {
	t.Parallel()

	argDto := &GetFsMetadataArgDto{
		ObjIDs: []string{"obj-1", "obj-2", "obj-1"},
	}

	efDto := NewGetFsMetadataEFArgDto(argDto)

	assert.NotNil(t, efDto)
	assert.Equal(t, "GET", efDto.Method)
	assert.NotNil(t, efDto.ObjIDs)
	// Should be deduplicated
	assert.Len(t, efDto.ObjIDs, 2)
	assert.Nil(t, efDto.IDs)
}

func TestNewGetFsMetadataEFArgDto_WithBothIDsAndObjIDs(t *testing.T) {
	t.Parallel()

	argDto := &GetFsMetadataArgDto{
		IDs:    []string{"id-1", "id-2"},
		ObjIDs: []string{"obj-1", "obj-2"},
	}

	efDto := NewGetFsMetadataEFArgDto(argDto)

	assert.NotNil(t, efDto)
	assert.Equal(t, "GET", efDto.Method)
	// Should prioritize IDs
	assert.NotNil(t, efDto.IDs)
	assert.Nil(t, efDto.ObjIDs)
}

func TestNewGetFsMetadataEFArgDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	argDto := &GetFsMetadataArgDto{
		IDs:    []string{},
		ObjIDs: []string{},
	}

	efDto := NewGetFsMetadataEFArgDto(argDto)

	assert.NotNil(t, efDto)
	assert.Equal(t, "", efDto.Method)
	assert.Nil(t, efDto.IDs)
	assert.Nil(t, efDto.ObjIDs)
}

func TestNewGetFsMetadataEFArgDto_WithEmptyDto(t *testing.T) {
	t.Parallel()

	argDto := &GetFsMetadataArgDto{}

	efDto := NewGetFsMetadataEFArgDto(argDto)

	assert.NotNil(t, efDto)
	assert.Equal(t, "", efDto.Method)
	assert.Nil(t, efDto.IDs)
	assert.Nil(t, efDto.ObjIDs)
}

func TestGetFsMetadataEFArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := &GetFsMetadataEFArgDto{
		IDs:    []string{"id-1"},
		ObjIDs: []string{"obj-1"},
		Method: "POST",
	}

	assert.NotNil(t, dto.IDs)
	assert.NotNil(t, dto.ObjIDs)
	assert.Equal(t, "POST", dto.Method)
}

func TestGetFsMetadataEFArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := &GetFsMetadataEFArgDto{}

	assert.Nil(t, dto.IDs)
	assert.Nil(t, dto.ObjIDs)
	assert.Empty(t, dto.Method)
}

func TestGetFsMetadataEFArgDto_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
	}

	for _, method := range methods {
		dto := &GetFsMetadataEFArgDto{
			Method: method,
		}
		assert.Equal(t, method, dto.Method)
	}
}

func TestIbFields_SliceOperations(t *testing.T) {
	t.Parallel()

	fields := IbFields{
		IbFieldName,
		IbFieldDocLibTypes,
		IbFieldPaths,
	}

	// Test slicing
	subFields := fields[1:3]
	assert.Len(t, subFields, 2)
	assert.Equal(t, IbFieldDocLibTypes, subFields[0])
}

func TestGetFsMetadataArgDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := GetFsMetadataArgDto{
		IDs: []string{"id-1", "id-2", "id-3"},
	}

	// Test slicing
	subIDs := dto.IDs[1:3]
	assert.Len(t, subIDs, 2)
	assert.Equal(t, "id-2", subIDs[0])
}

func TestIbFields_AllFieldTypes(t *testing.T) {
	t.Parallel()

	allFields := []IbField{
		IbFieldName,
		IbFieldDocLibTypes,
		IbFieldPaths,
	}

	fields := IbFields(allFields)
	result := fields.ToPathString()

	assert.Len(t, fields, 3)
	assert.Contains(t, result, "names")
	assert.Contains(t, result, "doc_lib_types")
	assert.Contains(t, result, "paths")
}
