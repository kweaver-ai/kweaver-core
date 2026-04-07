package umtypes

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"
	"github.com/stretchr/testify/assert"
)

func TestOsnInfoMapS_StructFields(t *testing.T) {
	t.Parallel()

	m := OsnInfoMapS{
		UserNameMap:       map[string]string{"user-1": "User 1"},
		DepartmentNameMap: map[string]string{"dept-1": "Dept 1"},
		GroupNameMap:      map[string]string{"group-1": "Group 1"},
		AppNameMap:        map[string]string{"app-1": "App 1"},
	}

	assert.Equal(t, "User 1", m.UserNameMap["user-1"])
	assert.Equal(t, "Dept 1", m.DepartmentNameMap["dept-1"])
	assert.Equal(t, "Group 1", m.GroupNameMap["group-1"])
	assert.Equal(t, "App 1", m.AppNameMap["app-1"])
}

func TestOsnInfoMapS_Empty(t *testing.T) {
	t.Parallel()

	m := OsnInfoMapS{}

	assert.Nil(t, m.UserNameMap)
	assert.Nil(t, m.DepartmentNameMap)
	assert.Nil(t, m.GroupNameMap)
	assert.Nil(t, m.AppNameMap)
}

func TestNewOsnInfoMapS(t *testing.T) {
	t.Parallel()

	m := NewOsnInfoMapS()

	assert.NotNil(t, m)
	assert.NotNil(t, m.UserNameMap)
	assert.NotNil(t, m.DepartmentNameMap)
	assert.NotNil(t, m.GroupNameMap)
	assert.NotNil(t, m.AppNameMap)
	assert.Empty(t, m.UserNameMap)
	assert.Empty(t, m.DepartmentNameMap)
	assert.Empty(t, m.GroupNameMap)
	assert.Empty(t, m.AppNameMap)
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithUsers(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		UserNames: []umret.IDName{
			{ID: "user-1", Name: "User 1"},
			{ID: "user-2", Name: "User 2"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.UserNameMap, 2)
	assert.Equal(t, "User 1", m.UserNameMap["user-1"])
	assert.Equal(t, "User 2", m.UserNameMap["user-2"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithDepartments(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		DepartmentNames: []umret.IDName{
			{ID: "dept-1", Name: "Dept 1"},
			{ID: "dept-2", Name: "Dept 2"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.DepartmentNameMap, 2)
	assert.Equal(t, "Dept 1", m.DepartmentNameMap["dept-1"])
	assert.Equal(t, "Dept 2", m.DepartmentNameMap["dept-2"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithGroups(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		GroupNames: []umret.IDName{
			{ID: "group-1", Name: "Group 1"},
			{ID: "group-2", Name: "Group 2"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.GroupNameMap, 2)
	assert.Equal(t, "Group 1", m.GroupNameMap["group-1"])
	assert.Equal(t, "Group 2", m.GroupNameMap["group-2"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithApps(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		AppNames: []umret.IDName{
			{ID: "app-1", Name: "App 1"},
			{ID: "app-2", Name: "App 2"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.AppNameMap, 2)
	assert.Equal(t, "App 1", m.AppNameMap["app-1"])
	assert.Equal(t, "App 2", m.AppNameMap["app-2"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithAllTypes(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		UserNames: []umret.IDName{
			{ID: "user-1", Name: "User 1"},
		},
		DepartmentNames: []umret.IDName{
			{ID: "dept-1", Name: "Dept 1"},
		},
		GroupNames: []umret.IDName{
			{ID: "group-1", Name: "Group 1"},
		},
		AppNames: []umret.IDName{
			{ID: "app-1", Name: "App 1"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.UserNameMap, 1)
	assert.Len(t, m.DepartmentNameMap, 1)
	assert.Len(t, m.GroupNameMap, 1)
	assert.Len(t, m.AppNameMap, 1)
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithChineseNames(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		UserNames: []umret.IDName{
			{ID: "user-1", Name: "用户1"},
		},
		DepartmentNames: []umret.IDName{
			{ID: "dept-1", Name: "部门1"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Equal(t, "用户1", m.UserNameMap["user-1"])
	assert.Equal(t, "部门1", m.DepartmentNameMap["dept-1"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithEmptyDto(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Empty(t, m.UserNameMap)
	assert.Empty(t, m.DepartmentNameMap)
	assert.Empty(t, m.GroupNameMap)
	assert.Empty(t, m.AppNameMap)
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithNilDto(t *testing.T) {
	t.Parallel()

	m := NewOsnInfoMapS()

	// FromGetOsnRetDto will panic if dto is nil
	assert.Panics(t, func() {
		m.FromGetOsnRetDto(nil)
	})
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithDuplicateIDs(t *testing.T) {
	t.Parallel()

	dto := &umret.GetOsnRetDto{
		UserNames: []umret.IDName{
			{ID: "user-1", Name: "User 1"},
			{ID: "user-1", Name: "User 1 Updated"},
		},
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	// Last one should win
	assert.Len(t, m.UserNameMap, 1)
	assert.Equal(t, "User 1 Updated", m.UserNameMap["user-1"])
}

func TestOsnInfoMapS_FromGetOsnRetDto_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	userNames := make([]umret.IDName, 50)
	for i := 0; i < 50; i++ {
		userNames[i] = umret.IDName{
			ID:   "user-" + string(rune(i)),
			Name: "User " + string(rune(i)),
		}
	}

	dto := &umret.GetOsnRetDto{
		UserNames: userNames,
	}

	m := NewOsnInfoMapS()
	m.FromGetOsnRetDto(dto)

	assert.Len(t, m.UserNameMap, 50)
}

func TestOsnInfoMapS_AllMapsInitialized(t *testing.T) {
	t.Parallel()

	m := NewOsnInfoMapS()

	// All maps should be initialized
	assert.NotNil(t, m.UserNameMap)
	assert.NotNil(t, m.DepartmentNameMap)
	assert.NotNil(t, m.GroupNameMap)
	assert.NotNil(t, m.AppNameMap)
}

func TestOsnInfoMapS_SetValueDirectly(t *testing.T) {
	t.Parallel()

	m := NewOsnInfoMapS()

	m.UserNameMap["user-1"] = "User 1"
	m.DepartmentNameMap["dept-1"] = "Dept 1"
	m.GroupNameMap["group-1"] = "Group 1"
	m.AppNameMap["app-1"] = "App 1"

	assert.Equal(t, "User 1", m.UserNameMap["user-1"])
	assert.Equal(t, "Dept 1", m.DepartmentNameMap["dept-1"])
	assert.Equal(t, "Group 1", m.GroupNameMap["group-1"])
	assert.Equal(t, "App 1", m.AppNameMap["app-1"])
}
