package sqlhelper2

import (
	"strings"
)

type AndClauses []*Clause

// Build 生成sql语句
// 示例：
//
//	clause := AndClauses{
//	 &Clause{
//	   Key:      "id",
//	   Value:    []int{1, 2, 3},
//	   Operator: OperatorIn,
//	 },
//	 &Clause{
//	 Key:      "updated_at",
//	 Value:    "2022-01-01",
//	 Operator: OperatorLtEq,
//	 },
//	}
//
// sqlStr, args := clause.Build()
// sqlStr: "id in (?,?,?) and updated_at <= ?"
// args: []interface{}{1, 2, 3, "2022-01-01"}
func (a *AndClauses) Build() (sqlStr string, args []interface{}, err error) {
	args = make([]interface{}, 0)

	if len(*a) == 0 {
		return
	}

	clauseStrs := make([]string, 0, len(*a))

	for _, clause := range *a {
		_sqlStr, _args, _err := clause.Build()
		if _err != nil {
			err = _err
			return
		}

		clauseStrs = append(clauseStrs, _sqlStr)
		args = append(args, _args...)
	}

	sqlStr = strings.Join(clauseStrs, " and ")

	return
}

type OrClauses []*Clause

// Build 生成sql语句
// 示例：
//
//	clause := OrClauses{
//	 &Clause{
//	   Key:      "id",
//	   Value:    []int{1, 2, 3},
//	   Operator: OperatorIn,
//	 },
//	 &Clause{
//	 Key:      "updated_at",
//	 Value:    "2022-01-01",
//	 Operator: OperatorLtEq,
//	 },
//	}
//
// sqlStr, args := clause.Build()
// sqlStr: "id in (?,?,?) or updated_at <= ?"
// args: []interface{}{1, 2, 3, "2022-01-01"}
func (o *OrClauses) Build() (sqlStr string, args []interface{}, err error) {
	args = make([]interface{}, 0)

	if len(*o) == 0 {
		return
	}

	clauseStrs := make([]string, 0, len(*o))

	for _, clause := range *o {
		_sqlStr, _args, _err := clause.Build()
		if _err != nil {
			err = _err
			return
		}

		clauseStrs = append(clauseStrs, _sqlStr)
		args = append(args, _args...)
	}

	sqlStr = strings.Join(clauseStrs, " or ")

	return
}

type WhereBuilder struct {
	Raws     []string
	RawsArgs []interface{}

	OrRaws     []string
	OrRawsArgs []interface{}

	AndClauses AndClauses
	OrClauses  OrClauses
	// AndGroups  []IClause
	// OrGroups   []IClause
}

func NewWhereBuilder() *WhereBuilder {
	return &WhereBuilder{
		AndClauses: AndClauses{},
		OrClauses:  OrClauses{},
		// AndGroups: make([]IClause, 0),
		// OrGroups:   make([]IClause, 0),
	}
}

func (w *WhereBuilder) Where(key string, op Operator, value interface{}) *WhereBuilder {
	if !op.Check() {
		panic("operator not support")
	}

	clause := &Clause{
		Key:      key,
		Value:    value,
		Operator: op,
	}

	w.AndClauses = append(w.AndClauses, clause)

	return w
}

func (w *WhereBuilder) WhereEqual(key string, value interface{}) *WhereBuilder {
	return w.Where(key, OperatorEq, value)
}

func (w *WhereBuilder) WhereNotEqual(key string, value interface{}) *WhereBuilder {
	return w.Where(key, OperatorNeq, value)
}

func (w *WhereBuilder) Like(key string, value interface{}) *WhereBuilder {
	return w.Where(key, OperatorLike, value)
}

func (w *WhereBuilder) Or(key string, op Operator, value interface{}) *WhereBuilder {
	if !op.Check() {
		panic("operator not support")
	}

	clause := &Clause{
		Key:      key,
		Value:    value,
		Operator: op,
	}

	w.OrClauses = append(w.OrClauses, clause)

	return w
}

func (w *WhereBuilder) OrEqual(key string, value interface{}) *WhereBuilder {
	return w.Or(key, OperatorEq, value)
}

func (w *WhereBuilder) OrNotEqual(key string, value interface{}) *WhereBuilder {
	return w.Or(key, OperatorNeq, value)
}

func (w *WhereBuilder) OrLike(key string, value interface{}) *WhereBuilder {
	return w.Or(key, OperatorLike, value)
}

func (w *WhereBuilder) In(key string, value interface{}) *WhereBuilder {
	clause := &Clause{
		Key:      key,
		Value:    value,
		Operator: OperatorIn,
	}

	w.AndClauses = append(w.AndClauses, clause)

	return w
}

func (w *WhereBuilder) NotIn(key string, value interface{}) *WhereBuilder {
	clause := &Clause{
		Key:      key,
		Value:    value,
		Operator: OperatorNotIn,
	}

	w.AndClauses = append(w.AndClauses, clause)

	return w
}

func (w *WhereBuilder) WhereRaw(condition string, arg ...interface{}) *WhereBuilder {
	if condition == "" {
		panic("condition cannot be empty")
	}

	w.Raws = append(w.Raws, w.hlCondition(condition))
	w.RawsArgs = append(w.RawsArgs, arg...)

	return w
}

func (w *WhereBuilder) WhereOrRaw(condition string, arg ...interface{}) *WhereBuilder {
	if condition == "" {
		panic("condition cannot be empty")
	}

	w.OrRaws = append(w.OrRaws, w.hlCondition(condition))
	w.OrRawsArgs = append(w.OrRawsArgs, arg...)

	return w
}

// 判断condition是否以"("开头，以")"结尾，如果不是，则添加
func (w *WhereBuilder) hlCondition(condition string) (newCondition string) {
	//newCondition = condition
	//这种不对 因为：当condition 如“(a=b and c=d) or (e=f)”时，也需要添加括号
	//if !strings.HasPrefix(condition, "(") || !strings.HasSuffix(condition, ")") {
	//	newCondition = "(" + condition + ")"
	//}
	// 空字符串或已经是完整括号包围的情况
	if condition == "" {
		return condition
	}

	// 检查是否最外层已有完整的括号
	if w.hasOuterParentheses(condition) {
		return condition
	}

	// 没有完整的外层括号，添加括号
	newCondition = "(" + condition + ")"

	return
}

// 检查字符串是否最外层有完整的括号
func (w *WhereBuilder) hasOuterParentheses(s string) bool {
	if len(s) < 2 {
		return false
	}

	// 必须以 ( 开头和 ) 结尾
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return false
	}

	// 检查括号匹配，确保最外层括号包围整个表达式
	level := 0

	for i, char := range s {
		switch char {
		case '(':
			level++
		case ')':
			level--
			// 如果在最后一个字符之前括号就平衡了，说明不是完整的外层括号
			if level == 0 && i < len(s)-1 {
				return false
			}
		}
	}

	// 最终括号应该平衡
	return level == 0
}

// ToWhereSQL 生成sql语句
// 子句优先级：and子句 > and分组 > or子句 > or分组
// and子句和and分组之间是and关系
// and分组 和 or子句 之间是or关系
// or子句和or分组之间是or关系
// 示例：见common/sqlhelper/where_builder_test.go
func (w *WhereBuilder) ToWhereSQL() (whereStr string, args []interface{}, err error) {
	args = make([]interface{}, 0)

	// and子句
	andSQLStr, andArgs, err := w.AndClauses.Build()
	if err != nil {
		return
	}

	// andsqlStr切片，用于存放and子句和and分组
	andSQLStrSlice := make([]string, 0)

	// 1.1: 先拼接and子句，加入到andSqlStrSlice
	if andSQLStr != "" {
		andSQLStrSlice = append(andSQLStrSlice, andSQLStr)
		args = append(args, andArgs...)
	}

	// 将and子句和and分组拼接起来
	if len(andSQLStrSlice) > 0 {
		whereStr = strings.Join(andSQLStrSlice, " and ")
	}

	// 2.1: 再拼接原生sql
	if len(w.Raws) > 0 {
		if whereStr != "" {
			whereStr = whereStr + " and " + strings.Join(w.Raws, " and ")
		} else {
			whereStr = strings.Join(w.Raws, " and ")
		}

		args = append(args, w.RawsArgs...)
	}

	// 2.2: 再拼接or原生sql
	if len(w.OrRaws) > 0 {
		tmpOrRaws := make([]string, len(w.OrRaws))
		for i := 0; i < len(w.OrRaws); i++ {
			tmpOrRaws[i] = w.hlCondition(w.OrRaws[i])
		}

		_tmpWhereStr := strings.Join(tmpOrRaws, " or ")

		if whereStr != "" {
			whereStr = whereStr + " or " + _tmpWhereStr
		} else {
			whereStr = _tmpWhereStr
		}

		args = append(args, w.OrRawsArgs...)
	}

	// 3.1: 再拼接or子句
	//【注意】：因为or的优先级低于and，所以or子句需要放到最后拼接
	// or子句
	orSQLStr, orArgs, err := w.OrClauses.Build()
	if err != nil {
		return
	}

	if orSQLStr != "" {
		if whereStr != "" {
			whereStr = whereStr + " or " + orSQLStr
		} else {
			whereStr = orSQLStr
		}

		args = append(args, orArgs...)
	}

	return
}
