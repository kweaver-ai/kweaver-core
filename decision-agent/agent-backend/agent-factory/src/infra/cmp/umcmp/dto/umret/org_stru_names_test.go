package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDName_StructFields(t *testing.T) {
	t.Parallel()

	idName := IDName{
		ID:   "id-123",
		Name: "Test Name",
	}

	assert.Equal(t, "id-123", idName.ID)
	assert.Equal(t, "Test Name", idName.Name)
}

func TestIDName_Empty(t *testing.T) {
	t.Parallel()

	idName := IDName{}

	assert.Empty(t, idName.ID)
	assert.Empty(t, idName.Name)
}

func TestIDName_WithChineseValues(t *testing.T) {
	t.Parallel()

	idName := IDName{
		ID:   "ID-123",
		Name: "测试名称",
	}

	assert.Equal(t, "ID-123", idName.ID)
	assert.Equal(t, "测试名称", idName.Name)
}

func TestGetOsnRetDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames: []IDName{
			{ID: "user-1", Name: "User 1"},
			{ID: "user-2", Name: "User 2"},
		},
		DepartmentNames: []IDName{
			{ID: "dept-1", Name: "Dept 1"},
		},
		GroupNames: []IDName{
			{ID: "group-1", Name: "Group 1"},
		},
		AppNames: []IDName{
			{ID: "app-1", Name: "App 1"},
		},
	}

	assert.Len(t, dto.UserNames, 2)
	assert.Len(t, dto.DepartmentNames, 1)
	assert.Len(t, dto.GroupNames, 1)
	assert.Len(t, dto.AppNames, 1)
	assert.Equal(t, "user-1", dto.UserNames[0].ID)
	assert.Equal(t, "dept-1", dto.DepartmentNames[0].ID)
	assert.Equal(t, "group-1", dto.GroupNames[0].ID)
	assert.Equal(t, "app-1", dto.AppNames[0].ID)
}

func TestGetOsnRetDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{}

	assert.Nil(t, dto.UserNames)
	assert.Nil(t, dto.DepartmentNames)
	assert.Nil(t, dto.GroupNames)
	assert.Nil(t, dto.AppNames)
}

func TestGetOsnRetDto_WithOnlyUserNames(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames: []IDName{
			{ID: "user-1", Name: "User 1"},
		},
	}

	assert.Len(t, dto.UserNames, 1)
	assert.Nil(t, dto.DepartmentNames)
	assert.Nil(t, dto.GroupNames)
	assert.Nil(t, dto.AppNames)
}

func TestGetOsnRetDto_WithOnlyDepartmentNames(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		DepartmentNames: []IDName{
			{ID: "dept-1", Name: "Dept 1"},
		},
	}

	assert.Nil(t, dto.UserNames)
	assert.Len(t, dto.DepartmentNames, 1)
	assert.Nil(t, dto.GroupNames)
	assert.Nil(t, dto.AppNames)
}

func TestGetOsnRetDto_WithOnlyGroupNames(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		GroupNames: []IDName{
			{ID: "group-1", Name: "Group 1"},
		},
	}

	assert.Nil(t, dto.UserNames)
	assert.Nil(t, dto.DepartmentNames)
	assert.Len(t, dto.GroupNames, 1)
	assert.Nil(t, dto.AppNames)
}

func TestGetOsnRetDto_WithOnlyAppNames(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		AppNames: []IDName{
			{ID: "app-1", Name: "App 1"},
		},
	}

	assert.Nil(t, dto.UserNames)
	assert.Nil(t, dto.DepartmentNames)
	assert.Nil(t, dto.GroupNames)
	assert.Len(t, dto.AppNames, 1)
}

func TestGetOsnRetDto_WithChineseNames(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames: []IDName{
			{ID: "user-1", Name: "用户1"},
		},
		DepartmentNames: []IDName{
			{ID: "dept-1", Name: "部门1"},
		},
		GroupNames: []IDName{
			{ID: "group-1", Name: "组1"},
		},
		AppNames: []IDName{
			{ID: "app-1", Name: "应用1"},
		},
	}

	assert.Equal(t, "用户1", dto.UserNames[0].Name)
	assert.Equal(t, "部门1", dto.DepartmentNames[0].Name)
	assert.Equal(t, "组1", dto.GroupNames[0].Name)
	assert.Equal(t, "应用1", dto.AppNames[0].Name)
}

func TestGetOsnRetDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames:       []IDName{},
		DepartmentNames: []IDName{},
		GroupNames:      []IDName{},
		AppNames:        []IDName{},
	}

	assert.NotNil(t, dto.UserNames)
	assert.NotNil(t, dto.DepartmentNames)
	assert.NotNil(t, dto.GroupNames)
	assert.NotNil(t, dto.AppNames)

	assert.Len(t, dto.UserNames, 0)
	assert.Len(t, dto.DepartmentNames, 0)
	assert.Len(t, dto.GroupNames, 0)
	assert.Len(t, dto.AppNames, 0)
}

func TestGetOsnRetDto_AppendUserNames(t *testing.T) {
	t.Parallel()

	dto := &GetOsnRetDto{}
	dto.UserNames = append(dto.UserNames, IDName{ID: "user-1", Name: "User 1"})
	dto.UserNames = append(dto.UserNames, IDName{ID: "user-2", Name: "User 2"})

	assert.Len(t, dto.UserNames, 2)
	assert.Equal(t, "user-1", dto.UserNames[0].ID)
	assert.Equal(t, "user-2", dto.UserNames[1].ID)
}

func TestGetOsnRetDto_AppendDepartmentNames(t *testing.T) {
	t.Parallel()

	dto := &GetOsnRetDto{}
	dto.DepartmentNames = append(dto.DepartmentNames, IDName{ID: "dept-1", Name: "Dept 1"})
	dto.DepartmentNames = append(dto.DepartmentNames, IDName{ID: "dept-2", Name: "Dept 2"})

	assert.Len(t, dto.DepartmentNames, 2)
	assert.Equal(t, "dept-1", dto.DepartmentNames[0].ID)
	assert.Equal(t, "dept-2", dto.DepartmentNames[1].ID)
}

func TestGetOsnRetDto_AppendGroupNames(t *testing.T) {
	t.Parallel()

	dto := &GetOsnRetDto{}
	dto.GroupNames = append(dto.GroupNames, IDName{ID: "group-1", Name: "Group 1"})
	dto.GroupNames = append(dto.GroupNames, IDName{ID: "group-2", Name: "Group 2"})

	assert.Len(t, dto.GroupNames, 2)
	assert.Equal(t, "group-1", dto.GroupNames[0].ID)
	assert.Equal(t, "group-2", dto.GroupNames[1].ID)
}

func TestGetOsnRetDto_AppendAppNames(t *testing.T) {
	t.Parallel()

	dto := &GetOsnRetDto{}
	dto.AppNames = append(dto.AppNames, IDName{ID: "app-1", Name: "App 1"})
	dto.AppNames = append(dto.AppNames, IDName{ID: "app-2", Name: "App 2"})

	assert.Len(t, dto.AppNames, 2)
	assert.Equal(t, "app-1", dto.AppNames[0].ID)
	assert.Equal(t, "app-2", dto.AppNames[1].ID)
}

func TestGetOsnRetDto_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	userNames := make([]IDName, 50)
	for i := 0; i < 50; i++ {
		userNames[i] = IDName{
			ID:   "user-" + string(rune(i)),
			Name: "User " + string(rune(i)),
		}
	}

	deptNames := make([]IDName, 30)
	for i := 0; i < 30; i++ {
		deptNames[i] = IDName{
			ID:   "dept-" + string(rune(i)),
			Name: "Dept " + string(rune(i)),
		}
	}

	dto := GetOsnRetDto{
		UserNames:       userNames,
		DepartmentNames: deptNames,
	}

	assert.Len(t, dto.UserNames, 50)
	assert.Len(t, dto.DepartmentNames, 30)
}

func TestGetOsnRetDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames: []IDName{
			{ID: "user-1", Name: "User 1"},
			{ID: "user-2", Name: "User 2"},
		},
		DepartmentNames: []IDName{
			{ID: "dept-1", Name: "Dept 1"},
		},
	}

	count := 0

	for _, userName := range dto.UserNames {
		assert.NotEmpty(t, userName.ID)
		assert.NotEmpty(t, userName.Name)

		count++
	}

	assert.Equal(t, 2, count)

	deptCount := 0

	for _, deptName := range dto.DepartmentNames {
		assert.NotEmpty(t, deptName.ID)
		assert.NotEmpty(t, deptName.Name)

		deptCount++
	}

	assert.Equal(t, 1, deptCount)
}

func TestIDName_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	idName := IDName{
		ID:   "id@#$%",
		Name: "Name (Test)",
	}

	assert.Equal(t, "id@#$%", idName.ID)
	assert.Equal(t, "Name (Test)", idName.Name)
}

func TestGetOsnRetDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := GetOsnRetDto{
		UserNames: []IDName{
			{ID: "user-1", Name: "User 1"},
			{ID: "user-2", Name: "User 2"},
			{ID: "user-3", Name: "User 3"},
		},
	}

	// Test slicing
	subUserNames := dto.UserNames[1:3]
	assert.Len(t, subUserNames, 2)
	assert.Equal(t, "user-2", subUserNames[0].ID)
}
