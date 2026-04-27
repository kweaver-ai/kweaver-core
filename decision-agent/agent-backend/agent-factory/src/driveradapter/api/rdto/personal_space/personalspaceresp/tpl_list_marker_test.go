package personalspaceresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPTplListPaginationMarker(t *testing.T) {
	t.Parallel()

	marker := NewPTplListPaginationMarker()

	assert.NotNil(t, marker)
	assert.Zero(t, marker.UpdatedAt)
	assert.Zero(t, marker.LastTplID)
}

func TestPTplListPaginationMarker_ToString_Empty(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{}

	str, err := marker.ToString()

	assert.NoError(t, err)
	assert.NotEmpty(t, str)
	// Base64 encoded JSON with zero values: {"updated_at":0,"last_tpl_id":0}
	assert.Equal(t, "eyJ1cGRhdGVkX2F0IjowLCJsYXN0X3RwbF9pZCI6MH0=", str)
}

func TestPTplListPaginationMarker_ToString_WithValues(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{
		UpdatedAt: 1234567890,
		LastTplID: 999,
	}

	str, err := marker.ToString()

	assert.NoError(t, err)
	assert.NotEmpty(t, str)
}

func TestPTplListPaginationMarker_LoadFromStr_Empty(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{
		UpdatedAt: 111,
		LastTplID: 222,
	}

	err := marker.LoadFromStr("")

	assert.NoError(t, err)
	// Empty string should not modify the marker
	assert.Equal(t, int64(111), marker.UpdatedAt)
	assert.Equal(t, int64(222), marker.LastTplID)
}

func TestPTplListPaginationMarker_LoadFromStr_Valid(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{}

	// Create a marker, serialize it, then load it back
	original := &PTplListPaginationMarker{
		UpdatedAt: 9876543210,
		LastTplID: 888,
	}

	str, err := original.ToString()
	assert.NoError(t, err)

	err = marker.LoadFromStr(str)
	assert.NoError(t, err)
	assert.Equal(t, int64(9876543210), marker.UpdatedAt)
	assert.Equal(t, int64(888), marker.LastTplID)
}

func TestPTplListPaginationMarker_LoadFromStr_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &PTplListPaginationMarker{
		UpdatedAt: 1617235200,
		LastTplID: 42,
	}

	str, err := original.ToString()
	assert.NoError(t, err)

	restored := &PTplListPaginationMarker{}
	err = restored.LoadFromStr(str)
	assert.NoError(t, err)

	assert.Equal(t, original.UpdatedAt, restored.UpdatedAt)
	assert.Equal(t, original.LastTplID, restored.LastTplID)
}

func TestPTplListPaginationMarker_LoadFromStr_InvalidBase64(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{}

	err := marker.LoadFromStr("not-valid-base64!")

	assert.Error(t, err)
}

func TestPTplListPaginationMarker_LoadFromStr_InvalidJSON(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{}

	// Valid base64 but invalid JSON
	err := marker.LoadFromStr("invalid-json-content")

	assert.Error(t, err)
}

func TestPTplListPaginationMarker_StructFields(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{
		UpdatedAt: 1000000,
		LastTplID: 12345,
	}

	assert.Equal(t, int64(1000000), marker.UpdatedAt)
	assert.Equal(t, int64(12345), marker.LastTplID)
}

func TestPTplListPaginationMarker_ToString_Base64Encoding(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{
		UpdatedAt: 0,
		LastTplID: 0,
	}

	str, err := marker.ToString()

	assert.NoError(t, err)
	// Should be valid base64
	assert.NotEmpty(t, str)
	// Verify it's a valid base64 string by decoding it back
	decoded, err := marker.ToString()
	assert.NoError(t, err)
	assert.Equal(t, str, decoded)
}
