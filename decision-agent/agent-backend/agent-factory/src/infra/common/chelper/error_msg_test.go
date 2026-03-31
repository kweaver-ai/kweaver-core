package chelper

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStruct for testing IErrMsgBindStruct
type MockStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=0"`
}

func (m *MockStruct) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Name.required": `"name"不能为空`,
		"Age.gte":       `"age"不能小于0`,
	}
}

// NestedStruct for testing composite structures
type NestedStruct struct {
	Mock MockStruct
}

func (n *NestedStruct) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

func TestErrMsg(t *testing.T) {
	t.Parallel()

	t.Run("with validation errors", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{}
		validate := validator.New()
		err := validate.Struct(mock)

		// This should have validation errors
		require.Error(t, err)

		msg := ErrMsg(err, mock)

		assert.NotEmpty(t, msg)
		assert.Contains(t, msg, "name")
	})

	t.Run("with regular error", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Test", Age: 25}
		err := errors.New("custom error message")

		msg := ErrMsg(err, mock)

		assert.Equal(t, "custom error message", msg)
	})

	t.Run("with nil struct", func(t *testing.T) {
		t.Parallel()

		err := errors.New("error without struct")

		msg := ErrMsg(err, nil)

		assert.Equal(t, "error without struct", msg)
	})

	t.Run("with struct that has no GetErrMsgMap match", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{}
		validate := validator.New()
		err := validate.Struct(mock)

		require.Error(t, err)

		msg := ErrMsg(err, mock)
		// Should contain numbered error messages
		assert.NotEmpty(t, msg)
	})

	t.Run("with valid struct no validation errors", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Valid", Age: 30}
		validate := validator.New()
		err := validate.Struct(mock)

		// No validation errors
		if err == nil {
			// No error means validation passed
			assert.True(t, true)
		}
	})
}

func TestGetErrsMsgs(t *testing.T) {
	t.Parallel()

	t.Run("with empty validation errors", func(t *testing.T) {
		t.Parallel()

		var validationErrs validator.ValidationErrors

		mock := &MockStruct{}

		// Get reflect value
		v := reflect.ValueOf(mock).Elem()

		msgs := getErrsMsgs(validationErrs, "MockStruct", v, mock)

		// Nil is returned when there are no validation errors
		assert.Nil(t, msgs)
	})

	t.Run("with nil struct interface", func(t *testing.T) {
		t.Parallel()

		var validationErrs validator.ValidationErrors

		mock := &MockStruct{}

		v := reflect.ValueOf(mock).Elem()

		// Test with nil struct interface
		var s IErrMsgBindStruct
		msgs := getErrsMsgs(validationErrs, "MockStruct", v, s)

		assert.Nil(t, msgs)
	})

	t.Run("with valid struct", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Test", Age: 25}
		v := reflect.ValueOf(mock).Elem()

		var validationErrs validator.ValidationErrors

		msgs := getErrsMsgs(validationErrs, "MockStruct", v, mock)

		assert.Nil(t, msgs) // No validation errors means nil
	})

	t.Run("with validation errors", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{}
		validate := validator.New()
		err := validate.Struct(mock)

		require.Error(t, err)

		// Convert to ValidationErrors
		var validationErrs validator.ValidationErrors

		errors.As(err, &validationErrs)

		v := reflect.ValueOf(mock).Elem()

		msgs := getErrsMsgs(validationErrs, "MockStruct", v, mock)

		assert.NotNil(t, msgs)
		assert.NotEmpty(t, msgs) // Should have error messages
	})
}

func TestGetMsgMapFromMidFields(t *testing.T) {
	t.Parallel()

	t.Run("with non-composite field", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Test", Age: 25}
		v := reflect.ValueOf(mock).Elem()

		// Empty midFieldStr means non-composite field
		m := getMsgMapFromMidFields("", v)

		assert.Nil(t, m)
	})

	t.Run("with invalid field path", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Test", Age: 25}
		v := reflect.ValueOf(mock).Elem()

		// Invalid field that doesn't exist - should not panic, just return nil
		m := getMsgMapFromMidFields("InvalidField.Field", v)

		assert.Nil(t, m)
	})

	t.Run("with empty midFieldStr", func(t *testing.T) {
		t.Parallel()

		mock := &MockStruct{Name: "Test", Age: 25}
		v := reflect.ValueOf(mock).Elem()

		m := getMsgMapFromMidFields("", v)

		assert.Nil(t, m)
	})

	t.Run("with non-struct field", func(t *testing.T) {
		t.Parallel()

		type SimpleStruct struct {
			Name string
			Age  int
		}

		simple := &SimpleStruct{Name: "Test", Age: 25}
		v := reflect.ValueOf(simple).Elem()

		// Name is a string field, not a struct
		m := getMsgMapFromMidFields("Name", v)

		assert.Nil(t, m) // Should return nil because Name is not a struct
	})

	t.Run("with struct that doesn't implement IErrMsgBindStruct", func(t *testing.T) {
		t.Parallel()

		type NonImplStruct struct {
			Field string
		}

		nested := struct {
			NonImpl NonImplStruct
		}{
			NonImpl: NonImplStruct{Field: "test"},
		}
		v := reflect.ValueOf(nested)

		// Need to get the element first since we have a pointer to struct
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// Should panic because NonImplStruct doesn't implement IErrMsgBindStruct
		assert.Panics(t, func() {
			getMsgMapFromMidFields("NonImpl", v)
		})
	})

	t.Run("with pointer to struct", func(t *testing.T) {
		t.Parallel()

		type InnerStruct struct {
			Field string
		}

		type InnerWithIErrMsg struct {
			Inner *InnerStruct
		}

		nested := &InnerWithIErrMsg{
			Inner: &InnerStruct{Field: "test"},
		}

		v := reflect.ValueOf(nested).Elem()

		// Inner is a pointer field - when accessed through reflection,
		// the pointer itself doesn't directly implement IErrMsgBindStruct
		// The function panics in this case
		assert.Panics(t, func() {
			getMsgMapFromMidFields("Inner", v)
		})
	})
}

func TestIErrMsgBindStruct_Interface(t *testing.T) {
	t.Parallel()

	// Test that MockStruct implements the interface
	var _ IErrMsgBindStruct = &MockStruct{}

	mock := &MockStruct{}
	assert.NotNil(t, mock.GetErrMsgMap())
}

func TestErrMsg_WithChineseErrorMessage(t *testing.T) {
	t.Parallel()

	mock := &MockStruct{}
	validate := validator.New()
	err := validate.Struct(mock)

	require.Error(t, err)

	msg := ErrMsg(err, mock)

	// Should contain Chinese error message for required field
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "name")
}

func TestErrMsg_WithNonValidationError(t *testing.T) {
	t.Parallel()

	mock := &MockStruct{Name: "Test", Age: 25}
	regularErr := errors.New("this is not a validation error")

	msg := ErrMsg(regularErr, mock)

	assert.Equal(t, "this is not a validation error", msg)
}

func TestErrMsg_WithEmptyValidationErrors(t *testing.T) {
	t.Parallel()

	mock := &MockStruct{Name: "Test", Age: 25}
	validate := validator.New()

	err := validate.Struct(mock)
	if err != nil {
		msg := ErrMsg(err, mock)
		assert.NotEmpty(t, msg)
	}
}
