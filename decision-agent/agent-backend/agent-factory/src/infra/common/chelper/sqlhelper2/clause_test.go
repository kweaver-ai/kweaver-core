package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClause_Build(t *testing.T) {
	t.Parallel()

	var err error

	// OperatorEq
	clause := Clause{
		Key:      "id",
		Value:    1,
		Operator: OperatorEq,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id = ?", sqlStr)
	assert.Equal(t, []interface{}{1}, args)

	// OperatorEq string
	clause = Clause{
		Key:      "name",
		Value:    "John",
		Operator: OperatorEq,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name = ?", sqlStr)
	assert.Equal(t, []interface{}{"John"}, args)

	// OperatorGt int
	clause = Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorGt,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age > ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)

	// OperatorLt int
	clause = Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorLt,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age < ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)

	// OperatorLike string
	clause = Clause{
		Key:      "email",
		Value:    "test@example.com",
		Operator: OperatorLike,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "email like ?", sqlStr)
	assert.Equal(t, []interface{}{"%test@example.com%"}, args)

	// OperatorGte string
	clause = Clause{
		Key:      "created_at",
		Value:    "2022-01-01",
		Operator: OperatorGte,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "created_at >= ?", sqlStr)
	assert.Equal(t, []interface{}{"2022-01-01"}, args)

	// OperatorLte string
	clause = Clause{
		Key:      "updated_at",
		Value:    "2022-01-01",
		Operator: OperatorLte,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "updated_at <= ?", sqlStr)
	assert.Equal(t, []interface{}{"2022-01-01"}, args)

	// OperatorIsNull
	clause = Clause{
		Key:      "deleted_at",
		Value:    "NULL",
		Operator: OperatorIsNull,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "deleted_at is null ", sqlStr)
	assert.Equal(t, []interface{}{}, args)

	// OperatorIsNotNull
	clause = Clause{
		Key:      "deleted_at",
		Value:    "NULL",
		Operator: OperatorIsNotNull,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "deleted_at is not null ", sqlStr)
	assert.Equal(t, []interface{}{}, args)
}

func TestClause_Build_OperatorIn(t *testing.T) {
	t.Parallel()

	var err error

	// OperatorIn []int
	clause := Clause{
		Key:      "id",
		Value:    []int{1, 2, 3},
		Operator: OperatorIn,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2, 3}, args)

	// OperatorIn []string
	clause = Clause{
		Key:      "id",
		Value:    []string{"1", "2", "3"},
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{"1", "2", "3"}, args)

	// OperatorIn []float64
	clause = Clause{
		Key:      "id",
		Value:    []float64{1.1, 2.2, 3.3},
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{1.1, 2.2, 3.3}, args)

	// OperatorNotIn []int
	clause = Clause{
		Key:      "id",
		Value:    []int{1, 2, 3},
		Operator: OperatorNotIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id not in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2, 3}, args)

	// OperatorNotIn []int empty slice
	clause = Clause{
		Key:      "id",
		Value:    []int{},
		Operator: OperatorNotIn,
	}
	sqlStr, args, err = clause.Build()
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "", sqlStr)
	assert.Equal(t, []interface{}{}, args)
}

func TestClause_Build_OperatorIn_duplicate(t *testing.T) {
	t.Parallel()

	var err error

	// OperatorIn []int has duplicate value
	clause := Clause{
		Key:      "id",
		Value:    []int{1, 2, 2},
		Operator: OperatorIn,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2}, args)

	// OperatorIn []string has duplicate value
	clause = Clause{
		Key:      "id",
		Value:    []string{"1", "2", "2"},
		Operator: OperatorIn,
	}

	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?)", sqlStr)
	assert.Equal(t, []interface{}{"1", "2"}, args)

	// OperatorIn []float64 has duplicate value
	clause = Clause{
		Key:      "id",
		Value:    []float64{1.1, 2.2, 2.2},
		Operator: OperatorIn,
	}

	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id in (?,?)", sqlStr)
	assert.Equal(t, []interface{}{1.1, 2.2}, args)

	// OperatorNotIn []int has duplicate value
	clause = Clause{
		Key:      "id",
		Value:    []int{1, 2, 2},
		Operator: OperatorNotIn,
	}

	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id not in (?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestClause_Build_OperatorNotIn(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "id",
		Value:    []int{1, 2, 3},
		Operator: OperatorNotIn,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "id not in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
}

func TestClause_Build_OperatorLike(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "name",
		Value:    "test",
		Operator: OperatorLike,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name like ?", sqlStr)
	assert.Equal(t, []interface{}{"%test%"}, args)
}

func TestClause_Build_OperatorNotLike(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "name",
		Value:    "test",
		Operator: OperatorNotLike,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name not like ?", sqlStr)
	assert.Equal(t, []interface{}{"%test%"}, args)
}

func TestClause_Build_OperatorIsNull(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "name",
		Value:    nil,
		Operator: OperatorIsNull,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name is null ", sqlStr)
	assert.Equal(t, []interface{}{}, args)
}

func TestClause_Build_OperatorIsNotNull(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "name",
		Value:    nil,
		Operator: OperatorIsNotNull,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name is not null ", sqlStr)
	assert.Equal(t, []interface{}{}, args)
}

func TestClause_Build_OperatorGt(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorGt,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age > ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_Build_OperatorLt(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorLt,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age < ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_Build_OperatorGte(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorGte,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age >= ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_Build_OperatorLte(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorLte,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age <= ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_Build_OperatorNotEq(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorNeq,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age <> ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_Build_OperatorEq(t *testing.T) {
	t.Parallel()

	var err error

	clause := Clause{
		Key:      "age",
		Value:    30,
		Operator: OperatorEq,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age = ?", sqlStr)
	assert.Equal(t, []interface{}{30}, args)
}

func TestClause_isAllElementsSameType(t *testing.T) {
	t.Parallel()

	// 1. 是同一类型
	s1 := []interface{}{1, 2, 3}
	assert.Equal(t, true, isAllElementsSameType(s1))

	// 2. 不是同一类型

	s2 := []interface{}{1, 2, "3"}
	assert.Equal(t, false, isAllElementsSameType(s2))
}

func TestClause_parseInClauseForInterfaceSlice(t *testing.T) {
	t.Parallel()

	// 1. []interface{}元素为int
	s1 := []interface{}{1, 2, 3}
	clause := Clause{
		Key:      "age",
		Value:    s1,
		Operator: OperatorIn,
	}
	sqlStr, args, err := clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{1, 2, 3}, args)

	// 2. []interface{}元素为string
	s2 := []interface{}{"1", "2", "3"}
	clause = Clause{
		Key:      "age",
		Value:    s2,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{"1", "2", "3"}, args)

	// 3. []interface{}元素为int8
	s3 := []interface{}{int8(1), int8(2), int8(3)}
	clause = Clause{
		Key:      "age",
		Value:    s3,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{int8(1), int8(2), int8(3)}, args)

	// 4. []interface{}元素为int16
	s4 := []interface{}{int16(1), int16(2), int16(3)}
	clause = Clause{
		Key:      "age",
		Value:    s4,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{int16(1), int16(2), int16(3)}, args)

	// 5. []interface{}元素为int32
	s5 := []interface{}{int32(1), int32(2), int32(3)}
	clause = Clause{
		Key:      "age",
		Value:    s5,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{int32(1), int32(2), int32(3)}, args)

	// 6. []interface{}元素为int64
	s6 := []interface{}{int64(1), int64(2), int64(3)}
	clause = Clause{
		Key:      "age",
		Value:    s6,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, args)

	// 7. []interface{}元素为uint
	s7 := []interface{}{uint(1), uint(2), uint(3)}
	clause = Clause{
		Key:      "age",
		Value:    s7,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{uint(1), uint(2), uint(3)}, args)

	// 8. []interface{}元素为uint8
	s8 := []interface{}{uint8(1), uint8(2), uint8(3)}
	clause = Clause{
		Key:      "age",
		Value:    s8,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{uint8(1), uint8(2), uint8(3)}, args)

	// 9. []interface{}元素为uint16
	s9 := []interface{}{uint16(1), uint16(2), uint16(3)}
	clause = Clause{
		Key:      "age",
		Value:    s9,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{uint16(1), uint16(2), uint16(3)}, args)

	// 10. []interface{}元素为uint32
	s10 := []interface{}{uint32(1), uint32(2), uint32(3)}
	clause = Clause{
		Key:      "age",
		Value:    s10,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{uint32(1), uint32(2), uint32(3)}, args)

	// 11. []interface{}元素为uint64
	s11 := []interface{}{uint64(1), uint64(2), uint64(3)}
	clause = Clause{
		Key:      "age",
		Value:    s11,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{uint64(1), uint64(2), uint64(3)}, args)

	// 12. []interface{}元素为float32
	s12 := []interface{}{float32(1), float32(2), float32(3)}
	clause = Clause{
		Key:      "age",
		Value:    s12,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{float32(1), float32(2), float32(3)}, args)

	// 13. []interface{}元素为float64
	s13 := []interface{}{float64(1), float64(2), float64(3)}
	clause = Clause{
		Key:      "age",
		Value:    s13,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, args)

	// 14. []interface{}元素为bool
	s14 := []interface{}{true, false}
	clause = Clause{
		Key:      "age",
		Value:    s14,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build() //nolint:staticcheck,ineffassign
	assert.Equal(t, err.Error(), "[parseInClauseForInterfaceSlice]: unsupported type")

	//	15. []interface{}内的元素不是同一类型
	s15 := []interface{}{1, "2", 3}
	clause = Clause{
		Key:      "age",
		Value:    s15,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build() //nolint:staticcheck,ineffassign
	assert.Equal(t, err.Error(), "when operator is OperatorIn or OperatorNotIn, all elements should be of the same type")

	// 16. []interface{}为空
	s16 := []interface{}{}
	clause = Clause{
		Key:      "age",
		Value:    s16,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build() //nolint:staticcheck,ineffassign
	assert.Equal(t, err.Error(), "[parseInClauseForInterfaceSlice]:OperatorNotIn and OperatorIn must have at least one element")

	// 17. []interface{}元素为interface{}
	s17 := []interface{}{interface{}(1), interface{}(2), interface{}(3)}
	clause = Clause{
		Key:      "age",
		Value:    s17,
		Operator: OperatorIn,
	}
	sqlStr, args, err = clause.Build()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age in (?,?,?)", sqlStr)
	assert.Equal(t, []interface{}{interface{}(1), interface{}(2), interface{}(3)}, args)
}

func TestClause_Build_OperatorIn_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Test with unsupported type (not a slice)
	clause := Clause{
		Key:      "id",
		Value:    "not a slice",
		Operator: OperatorIn,
	}
	sqlStr, args, err := clause.Build()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "OperatorNotIn and OperatorIn only support")
	assert.Empty(t, sqlStr)
	assert.Empty(t, args)
}

func TestClause_Build_OperatorNotIn_UnsupportedType(t *testing.T) {
	t.Parallel()

	// Test with unsupported type (not a slice)
	clause := Clause{
		Key:      "id",
		Value:    "not a slice",
		Operator: OperatorNotIn,
	}
	sqlStr, args, err := clause.Build()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "OperatorNotIn and OperatorIn only support")
	assert.Empty(t, sqlStr)
	assert.Empty(t, args)
}
