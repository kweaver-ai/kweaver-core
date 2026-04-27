package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyInterface interface {
	Method()
}

type Implements struct{}

func (Implements) Method() {}

type DoesNotImplement struct{}

func TestIsImplement(t *testing.T) {
	t.Parallel()

	var i MyInterface = Implements{}

	assert.True(t, IsImplement(Implements{}, &i))

	var notImple DoesNotImplement

	assert.False(t, IsImplement(&notImple, &i))
}

func TestMustStrSlice(t *testing.T) {
	t.Parallel()

	input := []interface{}{"a", "b", "c"}
	expected := []string{"a", "b", "c"}
	result := MustStrSlice(input)
	assert.Equal(t, expected, result)

	invalidInput := []interface{}{1, 2, 3}

	assert.Panics(t, func() { MustStrSlice(invalidInput) })
}

func TestMustStrSlice2(t *testing.T) {
	t.Parallel()

	input := []interface{}{"a", "b", "c"}
	expected := []string{"a", "b", "c"}
	result := MustStrSlice2(input)
	assert.Equal(t, expected, result)

	invalidInput := []interface{}{1, 2, 3}

	assert.Panics(t, func() { MustStrSlice2(invalidInput) })
}
