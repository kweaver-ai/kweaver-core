package sqlhelper2

type IClause interface {
	Build() (sqlStr string, args []interface{}, err error)
}
