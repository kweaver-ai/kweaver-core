package umarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfoField_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, UserInfoField("name"), FieldName)
	assert.Equal(t, UserInfoField("parent_deps"), FieldParentDeps)
	assert.Equal(t, UserInfoField("enabled"), FieldEnabled)
	assert.Equal(t, UserInfoField("roles"), FieldRoles)
	assert.Equal(t, UserInfoField("account"), FieldAccount)
	assert.Equal(t, UserInfoField("groups"), FieldGroups)
}

func TestFields_ToStrings(t *testing.T) {
	t.Parallel()

	fields := Fields{
		FieldName,
		FieldRoles,
		FieldAccount,
	}

	result := fields.ToStrings()

	assert.Len(t, result, 3)
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "roles", result[1])
	assert.Equal(t, "account", result[2])
}

func TestFields_ToStrings_Empty(t *testing.T) {
	t.Parallel()

	fields := Fields{}

	result := fields.ToStrings()

	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestGetUserInfoArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoArgDto{
		UserIds: []string{"user-1", "user-2"},
		Fields:  Fields{FieldName, FieldRoles},
	}

	assert.Len(t, dto.UserIds, 2)
	assert.Len(t, dto.Fields, 2)
}

func TestGetUserInfoArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoArgDto{}

	assert.Nil(t, dto.UserIds)
	assert.Nil(t, dto.Fields)
}

func TestGetUserEnableStatusArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetUserEnableStatusArgDto{
		UserIds: []string{"user-1", "user-2", "user-3"},
	}

	assert.Len(t, dto.UserIds, 3)
	assert.Equal(t, "user-1", dto.UserIds[0])
}

func TestGetUserEnableStatusArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetUserEnableStatusArgDto{}

	assert.Nil(t, dto.UserIds)
}

func TestGetUserInfoSingleArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoSingleArgDto{
		UserID: "user-123",
		Fields: Fields{FieldName, FieldAccount},
	}

	assert.Equal(t, "user-123", dto.UserID)
	assert.Len(t, dto.Fields, 2)
}

func TestGetUserInfoSingleArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoSingleArgDto{}

	assert.Empty(t, dto.UserID)
	assert.Nil(t, dto.Fields)
}

func TestFields_WithAllFields(t *testing.T) {
	t.Parallel()

	fields := Fields{
		FieldName,
		FieldParentDeps,
		FieldEnabled,
		FieldRoles,
		FieldAccount,
		FieldGroups,
	}

	assert.Len(t, fields, 6)
}

func TestFields_Append(t *testing.T) {
	t.Parallel()

	fields := Fields{}
	fields = append(fields, FieldName)
	fields = append(fields, FieldRoles)

	assert.Len(t, fields, 2)
}

func TestFields_Iteration(t *testing.T) {
	t.Parallel()

	fields := Fields{FieldName, FieldEnabled, FieldRoles}

	count := 0

	for _, field := range fields {
		assert.NotEmpty(t, string(field))

		count++
	}

	assert.Equal(t, 3, count)
}

func TestUserInfoField_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    UserInfoField
		expected string
	}{
		{name: "FieldName", field: FieldName, expected: "name"},
		{name: "FieldParentDeps", field: FieldParentDeps, expected: "parent_deps"},
		{name: "FieldEnabled", field: FieldEnabled, expected: "enabled"},
		{name: "FieldRoles", field: FieldRoles, expected: "roles"},
		{name: "FieldAccount", field: FieldAccount, expected: "account"},
		{name: "FieldGroups", field: FieldGroups, expected: "groups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserInfoArgDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoArgDto{
		UserIds: []string{"用户-1", "用户-2"},
		Fields:  Fields{FieldName},
	}

	assert.Len(t, dto.UserIds, 2)
	assert.Equal(t, "用户-1", dto.UserIds[0])
}

func TestGetUserEnableStatusArgDto_WithMultipleUsers(t *testing.T) {
	t.Parallel()

	userIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		userIDs[i] = "user-" + string(rune(i))
	}

	dto := GetUserEnableStatusArgDto{
		UserIds: userIDs,
	}

	assert.Len(t, dto.UserIds, 100)
}

func TestGetUserInfoSingleArgDto_WithAllFields(t *testing.T) {
	t.Parallel()

	dto := GetUserInfoSingleArgDto{
		UserID: "user-123",
		Fields: Fields{
			FieldName,
			FieldParentDeps,
			FieldEnabled,
			FieldRoles,
			FieldAccount,
			FieldGroups,
		},
	}

	assert.Len(t, dto.Fields, 6)
}
