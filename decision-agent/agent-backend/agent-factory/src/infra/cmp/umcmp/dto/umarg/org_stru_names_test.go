package umarg

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOsnArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1"},
		GroupIDs:      []string{"group-1", "group-2"},
		AppIDs:        []string{"app-1"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 1)
	assert.Len(t, dto.GroupIDs, 2)
	assert.Len(t, dto.AppIDs, 1)
	assert.Equal(t, "user-1", dto.UserIDs[0])
	assert.Equal(t, "dept-1", dto.DepartmentIDs[0])
	assert.Equal(t, "group-1", dto.GroupIDs[0])
	assert.Equal(t, "app-1", dto.AppIDs[0])
}

func TestGetOsnArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Nil(t, dto.GroupIDs)
	assert.Nil(t, dto.AppIDs)
}

func TestGetOsnArgDto_DeDupl(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "user-2", "user-1", "user-3"},
		DepartmentIDs: []string{"dept-1", "dept-2", "dept-1"},
		GroupIDs:      []string{"group-1", "group-2", "group-2"},
	}

	dto.DeDupl()

	assert.Len(t, dto.UserIDs, 3)
	assert.Len(t, dto.DepartmentIDs, 2)
	assert.Len(t, dto.GroupIDs, 2)
	assert.Contains(t, dto.UserIDs, "user-1")
	assert.Contains(t, dto.UserIDs, "user-2")
	assert.Contains(t, dto.UserIDs, "user-3")
}

func TestGetOsnArgDto_RemoveEmptyStrFromSlice(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "", "user-2", ""},
		DepartmentIDs: []string{"dept-1", "", "dept-2"},
		GroupIDs:      []string{"", "group-1", ""},
	}

	dto.RemoveEmptyStrFromSlice()

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 2)
	assert.Len(t, dto.GroupIDs, 1)
	assert.NotContains(t, dto.UserIDs, "")
	assert.NotContains(t, dto.DepartmentIDs, "")
	assert.NotContains(t, dto.GroupIDs, "")
}

func TestGetOsnArgDto_ToSfgKey(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "user-2", "user-1"},
		DepartmentIDs: []string{"dept-1", ""},
	}

	key, err := dto.ToSfgKey()

	assert.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.Contains(t, key, "user-1")
	assert.Contains(t, key, "user-2")
}

func TestGetOsnArgDto_DeDuplAndRemoveEmptyCombined(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "", "user-2", "user-1"},
		DepartmentIDs: []string{"dept-1", "dept-2", "dept-1", ""},
		GroupIDs:      []string{"group-1", "", "group-1"},
	}

	dto.DeDupl()
	dto.RemoveEmptyStrFromSlice()

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 2)
	assert.Len(t, dto.GroupIDs, 1)
	assert.NotContains(t, dto.UserIDs, "")
	assert.NotContains(t, dto.DepartmentIDs, "")
	assert.NotContains(t, dto.GroupIDs, "")
}

func TestGetOsnArgDto_WithOnlyUserIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs: []string{"user-1", "user-2"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Nil(t, dto.GroupIDs)
	assert.Nil(t, dto.AppIDs)
}

func TestGetOsnArgDto_WithOnlyDepartmentIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		DepartmentIDs: []string{"dept-1", "dept-2"},
	}

	assert.Nil(t, dto.UserIDs)
	assert.Len(t, dto.DepartmentIDs, 2)
	assert.Nil(t, dto.GroupIDs)
	assert.Nil(t, dto.AppIDs)
}

func TestGetOsnArgDto_WithOnlyGroupIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		GroupIDs: []string{"group-1", "group-2"},
	}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Len(t, dto.GroupIDs, 2)
	assert.Nil(t, dto.AppIDs)
}

func TestGetOsnArgDto_WithOnlyAppIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		AppIDs: []string{"app-1", "app-2"},
	}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Nil(t, dto.GroupIDs)
	assert.Len(t, dto.AppIDs, 2)
}

func TestGetOsnArgDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"用户-1", "用户-2"},
		DepartmentIDs: []string{"部门-1"},
		GroupIDs:      []string{"组-1"},
		AppIDs:        []string{"应用-1"},
	}

	assert.Equal(t, "用户-1", dto.UserIDs[0])
	assert.Equal(t, "部门-1", dto.DepartmentIDs[0])
	assert.Equal(t, "组-1", dto.GroupIDs[0])
	assert.Equal(t, "应用-1", dto.AppIDs[0])
}

func TestGetOsnArgDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{},
		DepartmentIDs: []string{},
		GroupIDs:      []string{},
		AppIDs:        []string{},
	}

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.NotNil(t, dto.GroupIDs)
	assert.NotNil(t, dto.AppIDs)

	assert.Len(t, dto.UserIDs, 0)
	assert.Len(t, dto.DepartmentIDs, 0)
	assert.Len(t, dto.GroupIDs, 0)
	assert.Len(t, dto.AppIDs, 0)
}

func TestGetOsnArgDto_AppendUserIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{}
	dto.UserIDs = append(dto.UserIDs, "user-1")
	dto.UserIDs = append(dto.UserIDs, "user-2")

	assert.Len(t, dto.UserIDs, 2)
	assert.Equal(t, "user-1", dto.UserIDs[0])
	assert.Equal(t, "user-2", dto.UserIDs[1])
}

func TestGetOsnArgDto_AppendDepartmentIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{}
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-1")
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-2")

	assert.Len(t, dto.DepartmentIDs, 2)
	assert.Equal(t, "dept-1", dto.DepartmentIDs[0])
	assert.Equal(t, "dept-2", dto.DepartmentIDs[1])
}

func TestGetOsnArgDto_AppendGroupIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{}
	dto.GroupIDs = append(dto.GroupIDs, "group-1")
	dto.GroupIDs = append(dto.GroupIDs, "group-2")

	assert.Len(t, dto.GroupIDs, 2)
	assert.Equal(t, "group-1", dto.GroupIDs[0])
	assert.Equal(t, "group-2", dto.GroupIDs[1])
}

func TestGetOsnArgDto_AppendAppIDs(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{}
	dto.AppIDs = append(dto.AppIDs, "app-1")
	dto.AppIDs = append(dto.AppIDs, "app-2")

	assert.Len(t, dto.AppIDs, 2)
	assert.Equal(t, "app-1", dto.AppIDs[0])
	assert.Equal(t, "app-2", dto.AppIDs[1])
}

func TestGetOsnArgDto_WithMultipleIDs(t *testing.T) {
	t.Parallel()

	userIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		userIDs[i] = "user-" + string(rune(i))
	}

	deptIDs := make([]string, 30)
	for i := 0; i < 30; i++ {
		deptIDs[i] = "dept-" + string(rune(i))
	}

	groupIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		groupIDs[i] = "group-" + string(rune(i))
	}

	appIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		appIDs[i] = "app-" + string(rune(i))
	}

	dto := GetOsnArgDto{
		UserIDs:       userIDs,
		DepartmentIDs: deptIDs,
		GroupIDs:      groupIDs,
		AppIDs:        appIDs,
	}

	assert.Len(t, dto.UserIDs, 50)
	assert.Len(t, dto.DepartmentIDs, 30)
	assert.Len(t, dto.GroupIDs, 20)
	assert.Len(t, dto.AppIDs, 10)
}

func TestGetOsnArgDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1"},
		GroupIDs:      []string{"group-1", "group-2"},
		AppIDs:        []string{"app-1"},
	}

	count := 0

	for _, userID := range dto.UserIDs {
		assert.NotEmpty(t, userID)

		count++
	}

	assert.Equal(t, 2, count)

	deptCount := 0

	for _, deptID := range dto.DepartmentIDs {
		assert.NotEmpty(t, deptID)

		deptCount++
	}

	assert.Equal(t, 1, deptCount)

	groupCount := 0

	for _, groupID := range dto.GroupIDs {
		assert.NotEmpty(t, groupID)

		groupCount++
	}

	assert.Equal(t, 2, groupCount)

	appCount := 0

	for _, appID := range dto.AppIDs {
		assert.NotEmpty(t, appID)

		appCount++
	}

	assert.Equal(t, 1, appCount)
}

func TestGetOsnArgDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs: []string{"user-1", "user-2", "user-3"},
	}

	// Test slicing
	subUserIDs := dto.UserIDs[1:3]
	assert.Len(t, subUserIDs, 2)
	assert.Equal(t, "user-2", subUserIDs[0])
}

func TestNewGetOsnUMArgDto(t *testing.T) {
	t.Parallel()

	argDto := &GetOsnArgDto{
		UserIDs: []string{"user-1", "user-2", "user-1"},
	}

	dto := NewGetOsnUMArgDto(argDto)

	assert.NotNil(t, dto)
	assert.Equal(t, argDto, dto.GetOsnArgDto)
	assert.Equal(t, http.MethodGet, dto.Method)
	// Should be deduplicated
	assert.Len(t, dto.GetOsnArgDto.UserIDs, 2)
}

func TestNewGetOsnUMArgDto_WithNilArgDto(t *testing.T) {
	t.Parallel()

	// NewGetOsnUMArgDto will panic if getOsnArgDto is nil
	// because it calls DeDupl() on the nil pointer
	// This test documents the expected behavior
	assert.Panics(t, func() {
		NewGetOsnUMArgDto(nil)
	})
}

func TestNewGetOsnUMArgDto_WithEmptyStrings(t *testing.T) {
	t.Parallel()

	argDto := &GetOsnArgDto{
		UserIDs:       []string{"user-1", "", "user-2"},
		DepartmentIDs: []string{"dept-1", "", "dept-2"},
	}

	dto := NewGetOsnUMArgDto(argDto)

	assert.NotNil(t, dto)
	// Should have empty strings removed
	assert.NotContains(t, dto.GetOsnArgDto.UserIDs, "")
	assert.NotContains(t, dto.GetOsnArgDto.DepartmentIDs, "")
}

func TestGetOsnUMArgDto_StructFields(t *testing.T) {
	t.Parallel()

	innerDto := &GetOsnArgDto{
		UserIDs: []string{"user-1"},
	}

	dto := &GetOsnUMArgDto{
		GetOsnArgDto: innerDto,
		Method:       http.MethodPost,
	}

	assert.Equal(t, innerDto, dto.GetOsnArgDto)
	assert.Equal(t, http.MethodPost, dto.Method)
}

func TestGetOsnUMArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := &GetOsnUMArgDto{}

	assert.Nil(t, dto.GetOsnArgDto)
	assert.Empty(t, dto.Method)
}

func TestGetOsnUMArgDto_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		dto := &GetOsnUMArgDto{
			Method: method,
		}
		assert.Equal(t, method, dto.Method)
	}
}

func TestGetOsnArgDto_AllFieldsSet(t *testing.T) {
	t.Parallel()

	dto := GetOsnArgDto{
		UserIDs:       []string{"user-1"},
		DepartmentIDs: []string{"dept-1"},
		GroupIDs:      []string{"group-1"},
		AppIDs:        []string{"app-1"},
	}

	assert.Len(t, dto.UserIDs, 1)
	assert.Len(t, dto.DepartmentIDs, 1)
	assert.Len(t, dto.GroupIDs, 1)
	assert.Len(t, dto.AppIDs, 1)
}
