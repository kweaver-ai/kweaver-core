package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperator_Check_ValidOperators(t *testing.T) {
	t.Parallel()

	validOperators := []Operator{
		OperatorGt,
		OperatorEq,
		OperatorLt,
		OperatorNeq,
		OperatorLike,
		OperatorGte,
		OperatorLte,
		OperatorIn,
		OperatorNotIn,
		OperatorIsNull,
		OperatorIsNotNull,
	}

	for _, op := range validOperators {
		t.Run(string(op), func(t *testing.T) {
			t.Parallel()

			result := op.Check()
			assert.True(t, result, "Operator %s should be valid", op)
		})
	}
}

func TestOperator_Check_InvalidOperators(t *testing.T) {
	t.Parallel()

	invalidOperators := []Operator{
		"",
		"invalid",
		"=",
		">",
		"<",
		" unknown ",
		"NOT_LIKE",
	}

	for _, op := range invalidOperators {
		t.Run(string(op), func(t *testing.T) {
			t.Parallel()

			result := op.Check()
			assert.False(t, result, "Operator %s should be invalid", op)
		})
	}
}

func TestOperator_Check_OperatorNotLike(t *testing.T) {
	t.Parallel()

	// Note: OperatorNotLike is defined but not in the valid list in Check()
	// This test documents the current behavior
	result := OperatorNotLike.Check()
	assert.False(t, result, "OperatorNotLike is not in the valid list")
}

func TestOperator_ConstantValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		operator Operator
		expected string
	}{
		{"OperatorGt", OperatorGt, " > "},
		{"OperatorEq", OperatorEq, " = "},
		{"OperatorLt", OperatorLt, " < "},
		{"OperatorNeq", OperatorNeq, " <> "},
		{"OperatorLike", OperatorLike, " like "},
		{"OperatorNotLike", OperatorNotLike, " not like "},
		{"OperatorGte", OperatorGte, " >= "},
		{"OperatorLte", OperatorLte, " <= "},
		{"OperatorIn", OperatorIn, " in "},
		{"OperatorNotIn", OperatorNotIn, " not in "},
		{"OperatorIsNull", OperatorIsNull, " is null "},
		{"OperatorIsNotNull", OperatorIsNotNull, " is not null "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.operator))
		})
	}
}

func TestOperator_StringRepresentation(t *testing.T) {
	t.Parallel()

	op := OperatorGt
	assert.Equal(t, " > ", string(op))
}

func TestOperator_CustomOperator(t *testing.T) {
	t.Parallel()

	customOp := Operator(" custom_op ")
	result := customOp.Check()
	assert.False(t, result, "Custom operators should not be valid")
}

func TestOperator_CaseSensitivity(t *testing.T) {
	t.Parallel()

	// Operators are case-sensitive
	op := Operator(" LIKE ")
	result := op.Check()
	assert.False(t, result, " lowercase like should not be valid")
}

func TestOperator_Whitespace(t *testing.T) {
	t.Parallel()

	// Test that whitespace matters
	op1 := Operator("=")
	op2 := Operator(OperatorEq)

	assert.False(t, op1.Check(), "Operator without whitespace should be invalid")
	assert.True(t, op2.Check(), "Operator with correct whitespace should be valid")
}
