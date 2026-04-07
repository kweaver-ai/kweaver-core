package comvalobj

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleInfo_New(t *testing.T) {
	t.Parallel()

	roleInfo := &RoleInfo{
		RoleID:   "role-123",
		RoleName: "Administrator",
	}

	assert.NotNil(t, roleInfo)
	assert.Equal(t, "role-123", roleInfo.RoleID)
	assert.Equal(t, "Administrator", roleInfo.RoleName)
}

func TestUserInfo_New(t *testing.T) {
	t.Parallel()

	userInfo := &UserInfo{
		UserID:   "user-456",
		Username: "john_doe",
	}

	assert.NotNil(t, userInfo)
	assert.Equal(t, "user-456", userInfo.UserID)
	assert.Equal(t, "john_doe", userInfo.Username)
}

func TestUserGroupInfo_New(t *testing.T) {
	t.Parallel()

	groupInfo := &UserGroupInfo{
		UserGroupID:   "group-789",
		UserGroupName: "Developers",
	}

	assert.NotNil(t, groupInfo)
	assert.Equal(t, "group-789", groupInfo.UserGroupID)
	assert.Equal(t, "Developers", groupInfo.UserGroupName)
}

func TestDepartmentInfo_New(t *testing.T) {
	t.Parallel()

	deptInfo := &DepartmentInfo{
		DepartmentID:   "dept-101",
		DepartmentName: "Engineering",
	}

	assert.NotNil(t, deptInfo)
	assert.Equal(t, "dept-101", deptInfo.DepartmentID)
	assert.Equal(t, "Engineering", deptInfo.DepartmentName)
}

func TestAppAccountInfo_New(t *testing.T) {
	t.Parallel()

	appInfo := &AppAccountInfo{
		AppAccountID:   "app-202",
		AppAccountName: "Production App",
	}

	assert.NotNil(t, appInfo)
	assert.Equal(t, "app-202", appInfo.AppAccountID)
	assert.Equal(t, "Production App", appInfo.AppAccountName)
}

func TestRoleInfo_JSONSerialization(t *testing.T) {
	t.Parallel()

	roleInfo := &RoleInfo{
		RoleID:   "role-001",
		RoleName: "Editor",
	}

	jsonBytes, err := json.Marshal(roleInfo)
	require.NoError(t, err)

	var deserialized RoleInfo
	err = json.Unmarshal(jsonBytes, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, roleInfo.RoleID, deserialized.RoleID)
	assert.Equal(t, roleInfo.RoleName, deserialized.RoleName)
}

func TestUserInfo_JSONSerialization(t *testing.T) {
	t.Parallel()

	userInfo := &UserInfo{
		UserID:   "user-002",
		Username: "jane_smith",
	}

	jsonBytes, err := json.Marshal(userInfo)
	require.NoError(t, err)

	var deserialized UserInfo
	err = json.Unmarshal(jsonBytes, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, userInfo.UserID, deserialized.UserID)
	assert.Equal(t, userInfo.Username, deserialized.Username)
}

func TestDepartmentInfo_JSONTags(t *testing.T) {
	t.Parallel()

	deptInfo := &DepartmentInfo{
		DepartmentID:   "dept-003",
		DepartmentName: "Sales",
	}

	jsonBytes, err := json.Marshal(deptInfo)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"department_id"`)
	assert.Contains(t, jsonStr, `"department_name"`)
}

func TestAppAccountInfo_JSONTags(t *testing.T) {
	t.Parallel()

	appInfo := &AppAccountInfo{
		AppAccountID:   "app-004",
		AppAccountName: "Staging App",
	}

	jsonBytes, err := json.Marshal(appInfo)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"app_account_id"`)
	assert.Contains(t, jsonStr, `"app_account_name"`)
}

func TestRoleInfo_EmptyFields(t *testing.T) {
	t.Parallel()

	roleInfo := &RoleInfo{}

	assert.NotNil(t, roleInfo)
	assert.Empty(t, roleInfo.RoleID)
	assert.Empty(t, roleInfo.RoleName)
}

func TestUserInfo_EmptyFields(t *testing.T) {
	t.Parallel()

	userInfo := &UserInfo{}

	assert.NotNil(t, userInfo)
	assert.Empty(t, userInfo.UserID)
	assert.Empty(t, userInfo.Username)
}

func TestUserGroupInfo_EmptyFields(t *testing.T) {
	t.Parallel()

	groupInfo := &UserGroupInfo{}

	assert.NotNil(t, groupInfo)
	assert.Empty(t, groupInfo.UserGroupID)
	assert.Empty(t, groupInfo.UserGroupName)
}

func TestDepartmentInfo_EmptyFields(t *testing.T) {
	t.Parallel()

	deptInfo := &DepartmentInfo{}

	assert.NotNil(t, deptInfo)
	assert.Empty(t, deptInfo.DepartmentID)
	assert.Empty(t, deptInfo.DepartmentName)
}

func TestAppAccountInfo_EmptyFields(t *testing.T) {
	t.Parallel()

	appInfo := &AppAccountInfo{}

	assert.NotNil(t, appInfo)
	assert.Empty(t, appInfo.AppAccountID)
	assert.Empty(t, appInfo.AppAccountName)
}
