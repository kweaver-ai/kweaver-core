package umarg

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchOrgArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1"},
		Scope:         []string{"scope-1", "scope-2"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 1)
	assert.Len(t, dto.Scope, 2)
}

func TestSearchOrgArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Nil(t, dto.Scope)
}

func TestNewSearchOrgUMArgDto(t *testing.T) {
	t.Parallel()

	argDto := &SearchOrgArgDto{
		UserIDs: []string{"user-1"},
	}

	dto := NewSearchOrgUMArgDto(argDto)

	assert.NotNil(t, dto)
	assert.Equal(t, argDto, dto.SearchOrgArgDto)
	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestNewSearchOrgUMArgDto_WithNilArgDto(t *testing.T) {
	t.Parallel()

	dto := NewSearchOrgUMArgDto(nil)

	assert.NotNil(t, dto)
	assert.Nil(t, dto.SearchOrgArgDto)
	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestSearchOrgUMArgDto_StructFields(t *testing.T) {
	t.Parallel()

	innerDto := &SearchOrgArgDto{
		UserIDs:       []string{"user-1"},
		DepartmentIDs: []string{"dept-1"},
		Scope:         []string{"scope-1"},
	}

	dto := &SearchOrgUMArgDto{
		SearchOrgArgDto: innerDto,
		Method:          http.MethodPost,
	}

	assert.Equal(t, innerDto, dto.SearchOrgArgDto)
	assert.Equal(t, http.MethodPost, dto.Method)
}

func TestSearchOrgUMArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := &SearchOrgUMArgDto{}

	assert.Nil(t, dto.SearchOrgArgDto)
	assert.Empty(t, dto.Method)
}

func TestSearchOrgArgDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs:       []string{"用户-1", "用户-2"},
		DepartmentIDs: []string{"部门-1"},
		Scope:         []string{"范围-1"},
	}

	assert.Equal(t, "用户-1", dto.UserIDs[0])
	assert.Equal(t, "部门-1", dto.DepartmentIDs[0])
}

func TestSearchOrgArgDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs: []string{"user-1", "user-2"},
		Scope:   []string{"scope-1", "scope-2"},
	}

	// Test slicing
	subScope := dto.Scope[1:2]
	assert.Len(t, subScope, 1)
	assert.Equal(t, "scope-2", subScope[0])
}

func TestSearchOrgArgDto_Append(t *testing.T) {
	t.Parallel()

	dto := &SearchOrgArgDto{}
	dto.UserIDs = append(dto.UserIDs, "user-1")
	dto.Scope = append(dto.Scope, "scope-1")

	assert.Len(t, dto.UserIDs, 1)
	assert.Len(t, dto.Scope, 1)
}

func TestSearchOrgUMArgDto_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		dto := &SearchOrgUMArgDto{
			Method: method,
		}
		assert.Equal(t, method, dto.Method)
	}
}

func TestSearchOrgArgDto_WithMultipleArrays(t *testing.T) {
	t.Parallel()

	userIDs := make([]string, 30)
	for i := 0; i < 30; i++ {
		userIDs[i] = "user-" + string(rune(i))
	}

	deptIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		deptIDs[i] = "dept-" + string(rune(i))
	}

	scopes := make([]string, 10)
	for i := 0; i < 10; i++ {
		scopes[i] = "scope-" + string(rune(i))
	}

	dto := SearchOrgArgDto{
		UserIDs:       userIDs,
		DepartmentIDs: deptIDs,
		Scope:         scopes,
	}

	assert.Len(t, dto.UserIDs, 30)
	assert.Len(t, dto.DepartmentIDs, 20)
	assert.Len(t, dto.Scope, 10)
}

func TestSearchOrgArgDto_WithOnlyUserIDs(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs: []string{"user-1", "user-2"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Nil(t, dto.Scope)
}

func TestSearchOrgArgDto_WithOnlyScope(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		Scope: []string{"scope-1", "scope-2"},
	}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
	assert.Len(t, dto.Scope, 2)
}

func TestSearchOrgUMArgDto_DefaultMethod(t *testing.T) {
	t.Parallel()

	argDto := &SearchOrgArgDto{}
	dto := NewSearchOrgUMArgDto(argDto)

	assert.Equal(t, http.MethodGet, dto.Method)
}

func TestSearchOrgArgDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs: []string{"user-1", "user-2"},
		Scope:   []string{"scope-1"},
	}

	userCount := 0

	for _, userID := range dto.UserIDs {
		assert.NotEmpty(t, userID)

		userCount++
	}

	assert.Equal(t, 2, userCount)

	scopeCount := 0

	for _, scope := range dto.Scope {
		assert.NotEmpty(t, scope)

		scopeCount++
	}

	assert.Equal(t, 1, scopeCount)
}

func TestSearchOrgArgDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := SearchOrgArgDto{
		UserIDs:       []string{},
		DepartmentIDs: []string{},
		Scope:         []string{},
	}

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.NotNil(t, dto.Scope)

	assert.Len(t, dto.UserIDs, 0)
	assert.Len(t, dto.DepartmentIDs, 0)
	assert.Len(t, dto.Scope, 0)
}
