package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt8Ptr(t *testing.T) {
	t.Parallel()

	value := int8(42)
	ptr := Int8Ptr(value)

	assert.NotNil(t, ptr, "Int8Ptr() should return non-nil pointer")
	assert.Equal(t, value, *ptr, "Dereferenced pointer should equal input value")
}

func TestInt8PtrZero(t *testing.T) {
	t.Parallel()

	value := int8(0)
	ptr := Int8Ptr(value)

	assert.NotNil(t, ptr, "Int8Ptr() should return non-nil pointer for zero value")
	assert.Equal(t, value, *ptr, "Dereferenced pointer should equal zero value")
}

func TestInt8PtrNegative(t *testing.T) {
	t.Parallel()

	value := int8(-42)
	ptr := Int8Ptr(value)

	assert.NotNil(t, ptr, "Int8Ptr() should return non-nil pointer for negative value")
	assert.Equal(t, value, *ptr, "Dereferenced pointer should equal negative value")
}
