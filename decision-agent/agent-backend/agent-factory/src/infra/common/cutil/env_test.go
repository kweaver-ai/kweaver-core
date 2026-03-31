package cutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_WithExistingEnv(t *testing.T) {
	t.Parallel()

	// Set up test environment variable
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result := GetEnv("TEST_ENV_VAR", "default_value")
	assert.Equal(t, "test_value", result)
}

func TestGetEnv_WithNonExistingEnv(t *testing.T) {
	t.Parallel()

	// Make sure the env var doesn't exist
	os.Unsetenv("NON_EXISTING_VAR")

	result := GetEnv("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetEnv_WithEmptyEnv(t *testing.T) {
	t.Parallel()

	// Set env var to empty string
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result := GetEnv("EMPTY_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestGetEnvMustInt_WithValidInt(t *testing.T) {
	t.Parallel()

	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	result := GetEnvMustInt("TEST_INT_VAR", 10)
	assert.Equal(t, 42, result)
}

func TestGetEnvMustInt_WithNonExistingEnv(t *testing.T) {
	t.Parallel()

	os.Unsetenv("NON_EXISTING_INT_VAR")

	result := GetEnvMustInt("NON_EXISTING_INT_VAR", 10)
	assert.Equal(t, 10, result)
}

func TestGetEnvMustInt_WithDefaultValue(t *testing.T) {
	t.Parallel()

	os.Unsetenv("DEFAULT_INT_VAR")

	result := GetEnvMustInt("DEFAULT_INT_VAR", 99)
	assert.Equal(t, 99, result)
}

func TestGetEnvMustInt_WithInvalidInt(t *testing.T) {
	t.Parallel()

	os.Setenv("INVALID_INT_VAR", "not_a_number")
	defer os.Unsetenv("INVALID_INT_VAR")

	assert.Panics(t, func() {
		GetEnvMustInt("INVALID_INT_VAR", 10)
	})
}

func TestGetEnvMustInt_WithNegativeInt(t *testing.T) {
	t.Parallel()

	os.Setenv("NEGATIVE_INT_VAR", "-42")
	defer os.Unsetenv("NEGATIVE_INT_VAR")

	result := GetEnvMustInt("NEGATIVE_INT_VAR", 0)
	assert.Equal(t, -42, result)
}

func TestGetEnvMustInt_WithZero(t *testing.T) {
	t.Parallel()

	os.Setenv("ZERO_INT_VAR", "0")
	defer os.Unsetenv("ZERO_INT_VAR")

	result := GetEnvMustInt("ZERO_INT_VAR", 10)
	assert.Equal(t, 0, result)
}
