package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserEnabledMap_TypeDefinition(t *testing.T) {
	t.Parallel()

	// Test that UserEnabledMap is a map type
	var m UserEnabledMap

	assert.Nil(t, m)
}

func TestUserEnabledMap_WithValues(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
		"user-2": false,
		"user-3": true,
	}

	assert.Len(t, m, 3)
	assert.True(t, m["user-1"])
	assert.False(t, m["user-2"])
	assert.True(t, m["user-3"])
}

func TestUserEnabledMap_Empty(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{}
	assert.Len(t, m, 0)
}

func TestUserEnabledMap_NewMap(t *testing.T) {
	t.Parallel()

	m := make(UserEnabledMap)
	assert.NotNil(t, m)

	m["user-1"] = true
	m["user-2"] = false

	assert.Len(t, m, 2)
}

func TestUserEnabledMap_AddValues(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{}

	m["user-1"] = true
	assert.True(t, m["user-1"])
	assert.Len(t, m, 1)

	m["user-2"] = false
	assert.False(t, m["user-2"])
	assert.Len(t, m, 2)
}

func TestUserEnabledMap_UpdateValue(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
	}

	assert.True(t, m["user-1"])

	// Update value
	m["user-1"] = false
	assert.False(t, m["user-1"])
}

func TestUserEnabledMap_DeleteValue(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
		"user-2": false,
	}

	assert.Len(t, m, 2)

	delete(m, "user-1")
	assert.Len(t, m, 1)
	// user-1 no longer exists
	_, exists := m["user-1"]
	assert.False(t, exists)
	assert.False(t, m["user-2"])
}

func TestUserEnabledMap_WithChineseUserID(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"用户-1": true,
		"用户-2": false,
	}

	assert.True(t, m["用户-1"])
	assert.False(t, m["用户-2"])
}

func TestUserEnabledMap_Iteration(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
		"user-2": false,
		"user-3": true,
	}

	count := 0
	enabledCount := 0

	for userID, enabled := range m {
		assert.NotEmpty(t, userID)

		count++

		if enabled {
			enabledCount++
		}
	}

	assert.Equal(t, 3, count)
	assert.Equal(t, 2, enabledCount)
}

func TestUserEnabledMap_CheckValue(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
		"user-2": false,
	}

	// Check existing user
	val, exists := m["user-1"]
	assert.True(t, exists)
	assert.True(t, val)

	// Check non-existing user
	val, exists = m["user-999"]
	assert.False(t, exists)
	assert.False(t, val)
}

func TestUserEnabledMap_WithDifferentValueTypes(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-1": true,
		"user-2": false,
	}

	// All values should be bool
	for _, v := range m {
		assert.IsType(t, false, v)
	}
}

func TestUserEnabledMap_Copy(t *testing.T) {
	t.Parallel()

	original := UserEnabledMap{
		"user-1": true,
		"user-2": false,
	}

	// Create a copy
	copy := make(UserEnabledMap)
	for k, v := range original {
		copy[k] = v
	}

	assert.Equal(t, original, copy)
	assert.Len(t, copy, 2)

	// Modify copy shouldn't affect original
	copy["user-3"] = true

	assert.Len(t, original, 2)
	assert.Len(t, copy, 3)
}

func TestUserEnabledMap_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	m := UserEnabledMap{
		"user-with-dash":       true,
		"user.with.dot":        false,
		"user_with_underscore": true,
	}

	assert.Len(t, m, 3)
	assert.True(t, m["user-with-dash"])
	assert.False(t, m["user.with.dot"])
	assert.True(t, m["user_with_underscore"])
}
