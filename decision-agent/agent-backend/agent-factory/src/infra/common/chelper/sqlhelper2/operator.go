package sqlhelper2

type Operator string

func (o Operator) Check() bool {
	switch o {
	case OperatorGt, OperatorEq, OperatorLt, OperatorNeq, OperatorLike, OperatorGte,
		OperatorLte, OperatorIn, OperatorNotIn, OperatorIsNull, OperatorIsNotNull:
		return true
	}

	return false
}

const (
	OperatorGt      Operator = " > "    // 大于
	OperatorEq      Operator = " = "    // 等于
	OperatorLt      Operator = " < "    // 小于
	OperatorNeq     Operator = " <> "   // 不等于
	OperatorLike    Operator = " like " // 模糊查询
	OperatorNotLike Operator = " not like "
	OperatorGte     Operator = " >= " // 大于等于
	OperatorLte     Operator = " <= " // 小于等于
	OperatorIn      Operator = " in " // in

	OperatorNotIn Operator = " not in " // not in

	OperatorIsNull    Operator = " is null "     // is null
	OperatorIsNotNull Operator = " is not null " // is not null

)
