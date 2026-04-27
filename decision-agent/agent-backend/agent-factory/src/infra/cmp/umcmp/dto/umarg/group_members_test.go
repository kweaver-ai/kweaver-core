package umarg

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGroupMembersArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{
		GroupIDs: []string{"group-1", "group-2"},
	}

	assert.Len(t, dto.GroupIDs, 2)
	assert.Equal(t, "group-1", dto.GroupIDs[0])
}

func TestGetGroupMembersArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{}

	assert.Nil(t, dto.GroupIDs)
}

func TestNewGetGroupMembersUMArgDto(t *testing.T) {
	t.Parallel()

	argDto := &GetGroupMembersArgDto{
		GroupIDs: []string{"group-1"},
	}

	dto := NewGetGroupMembersUMArgDto(argDto)

	assert.NotNil(t, dto)
	assert.Equal(t, argDto, dto.GetGroupMembersArgDto)
	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestNewGetGroupMembersUMArgDto_WithNilArgDto(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersUMArgDto(nil)

	assert.NotNil(t, dto)
	assert.Nil(t, dto.GetGroupMembersArgDto)
	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestGetGroupMembersUMArgDto_StructFields(t *testing.T) {
	t.Parallel()

	innerDto := &GetGroupMembersArgDto{
		GroupIDs: []string{"group-1", "group-2"},
	}

	dto := &GetGroupMembersUMArgDto{
		GetGroupMembersArgDto: innerDto,
		Method:                http.MethodPost,
	}

	assert.Equal(t, innerDto, dto.GetGroupMembersArgDto)
	assert.Equal(t, http.MethodPost, dto.Method)
}

func TestGetGroupMembersUMArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersUMArgDto{}

	assert.Nil(t, dto.GetGroupMembersArgDto)
	assert.Empty(t, dto.Method)
}

func TestGetGroupMembersArgDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{
		GroupIDs: []string{"组-1", "组-2"},
	}

	assert.Len(t, dto.GroupIDs, 2)
	assert.Equal(t, "组-1", dto.GroupIDs[0])
}

func TestGetGroupMembersArgDto_SliceOperations(t *testing.T) {
	t.Parallel()

	groupIDs := []string{"group-1", "group-2", "group-3"}

	dto := GetGroupMembersArgDto{
		GroupIDs: groupIDs,
	}

	// Test slicing
	subGroupIDs := dto.GroupIDs[1:3]
	assert.Len(t, subGroupIDs, 2)
	assert.Equal(t, "group-2", subGroupIDs[0])
}

func TestGetGroupMembersArgDto_Append(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{}
	dto.GroupIDs = append(dto.GroupIDs, "group-1")
	dto.GroupIDs = append(dto.GroupIDs, "group-2")

	assert.Len(t, dto.GroupIDs, 2)
}

func TestGetGroupMembersUMArgDto_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		dto := &GetGroupMembersUMArgDto{
			Method: method,
		}
		assert.Equal(t, method, dto.Method)
	}
}

func TestGetGroupMembersArgDto_WithMultipleGroups(t *testing.T) {
	t.Parallel()

	groupIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		groupIDs[i] = "group-" + string(rune(i))
	}

	dto := GetGroupMembersArgDto{
		GroupIDs: groupIDs,
	}

	assert.Len(t, dto.GroupIDs, 50)
}

func TestGetGroupMembersUMArgDto_WithGetGroupMembersArgDto(t *testing.T) {
	t.Parallel()

	argDto := &GetGroupMembersArgDto{
		GroupIDs: []string{"group-1"},
	}

	dto := NewGetGroupMembersUMArgDto(argDto)

	assert.NotNil(t, dto.GetGroupMembersArgDto)
	assert.Equal(t, "group-1", dto.GetGroupMembersArgDto.GroupIDs[0])
}

func TestGetGroupMembersArgDto_EmptySlice(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{
		GroupIDs: []string{},
	}

	assert.NotNil(t, dto.GroupIDs)
	assert.Len(t, dto.GroupIDs, 0)
}

func TestGetGroupMembersUMArgDto_DefaultMethod(t *testing.T) {
	t.Parallel()

	argDto := &GetGroupMembersArgDto{}
	dto := NewGetGroupMembersUMArgDto(argDto)

	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestGetGroupMembersArgDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersArgDto{
		GroupIDs: []string{"group-1", "group-2", "group-3"},
	}

	count := 0

	for _, groupID := range dto.GroupIDs {
		assert.NotEmpty(t, groupID)

		count++
	}

	assert.Equal(t, 3, count)
}
