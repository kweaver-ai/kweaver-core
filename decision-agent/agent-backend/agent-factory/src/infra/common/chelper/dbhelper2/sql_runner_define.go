package dbhelper2

import (
	"database/sql"
	"reflect"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/sqlhelper2"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type SQLRunner struct {
	logger icmp.Logger

	sb            *sqlhelper2.SelectBuilder
	deleteBuilder *sqlhelper2.DeleteBuilder
	ib            *sqlhelper2.InsertBuilder
	ub            *sqlhelper2.UpdateBuilder
	db            ISQLRunner
	tag           string
	rawSQL        string // 原始sql
	rawSQLArgs    []interface{}
	po            ITable

	lastSql  string
	lastArgs []interface{}
}

func NewSQLRunner(db ISQLRunner, logger icmp.Logger) *SQLRunner {
	tmp := &SQLRunner{
		logger: logger,
		db:     db,
		sb:     sqlhelper2.NewSelectBuilder(),
		tag:    "db",
	}
	tmp.updateBuilders()

	return tmp
}

// TxSr 获取SQLRunner，用于事务
func TxSr(tx *sql.Tx, logger icmp.Logger) *SQLRunner {
	return NewSQLRunner(tx, logger)
}

func SrByISQLRunner(db ISQLRunner, logger icmp.Logger) *SQLRunner {
	return NewSQLRunner(db, logger)
}

func NewQueryWithSQLBuilder(db ISQLRunner, sb *sqlhelper2.SelectBuilder, logger icmp.Logger) *SQLRunner {
	tmp := &SQLRunner{
		logger: logger,

		db:  db,
		sb:  sb,
		tag: "db",
	}
	tmp.updateBuilders()

	return tmp
}

func (q *SQLRunner) updateBuilders() {
	q.deleteBuilder = sqlhelper2.NewDeleteBuilder()
	q.deleteBuilder.SetWhereBuilder(q.sb.GetWhereBuilder())

	q.ub = sqlhelper2.NewUpdateBuilder()
	q.ub.SetWhereBuilder(q.sb.GetWhereBuilder()).
		Tag(q.tag)

	q.ib = sqlhelper2.NewInsertBuilder()
	q.ib.Tag(q.tag)
}

func (q *SQLRunner) SetSQLBuilder(sb *sqlhelper2.SelectBuilder) *SQLRunner {
	q.sb = sb
	q.updateBuilders()

	return q
}

func (q *SQLRunner) Tag(tag string) *SQLRunner {
	q.tag = tag
	q.ib.Tag(tag)
	q.ub.Tag(tag)

	return q
}

// struct2ScanArgsByTag 将结构体中的字段按照tag的值，转换为scan所需的参数
// obj 必须是一个结构体或者结构体指针
// 注意：
// 1.如果结构体中有匿名字段，会递归处理
// 2.匿名字段不能为指针类型
func (q *SQLRunner) struct2ScanArgsByTag(obj interface{}, opts ...Option) (args []interface{}) {
	v := reflect.ValueOf(obj)
	t := v.Type()

	for t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("obj must be a struct or a pointer to struct")
	}

	selectFields := make([]string, 0)
	// 如果指定了rawSQL，没有此逻辑
	if q.rawSQL == "" && len(opts) > 0 {
		selectFields = opts[0].SelectFields
	}

	for i := 0; i < v.NumField(); i++ {
		// 支持匿名字段
		if t.Field(i).Anonymous {
			tmpFieldType := t.Field(i).Type
			tmpFieldVal := v.Field(i)

			if tmpFieldType.Kind() == reflect.Ptr {
				tmpFieldType = tmpFieldType.Elem()
				tmpFieldVal = tmpFieldVal.Elem()
			}

			if tmpFieldType.Kind() == reflect.Struct {
				args = append(args, q.struct2ScanArgsByTag(tmpFieldVal.Addr().Interface(), opts...)...)
			}

			continue
		}

		if key, ok := t.Field(i).Tag.Lookup(q.tag); ok {
			if key == "" || key == "-" {
				continue
			}

			// 如果指定了selectFields，只处理selectFields中的字段
			if len(selectFields) > 0 && !cutil.Exists(selectFields, key) {
				continue
			}

			args = append(args, v.Field(i).Addr().Interface())
		}
	}

	return
}

func (q *SQLRunner) FromPo(po ITable) *SQLRunner {
	q.From(po.TableName())
	q.po = po

	return q
}

func (q *SQLRunner) From(table string) *SQLRunner {
	q.sb.From(table)
	q.ib.From(table)
	q.deleteBuilder.From(table)
	q.ub.From(table)

	return q
}

func (q *SQLRunner) Select(columns []string) *SQLRunner {
	q.sb.Select(columns)
	return q
}

// ResetSelect 重置（清空）select字段
// 比如使用Count()后，可清空select字段，后面Find()时发现没有select字段，会自动select所有字段
func (q *SQLRunner) ResetSelect() *SQLRunner {
	q.sb.Select([]string{})
	return q
}

// Raw 原始sql
// 需要注意字段的顺序问题 xx.Raw(rawSQL).Find(&obj) obj的字段顺序要和rawSQL中的字段顺序一致
func (q *SQLRunner) Raw(_sql string, args ...interface{}) *SQLRunner {
	q.rawSQL = _sql
	q.rawSQLArgs = make([]interface{}, 0)

	if len(args) > 0 {
		q.rawSQLArgs = args
	}

	return q
}

func (q *SQLRunner) Limit(limit int) *SQLRunner {
	q.sb.Limit(limit)
	return q
}

func (q *SQLRunner) Offset(offset int) *SQLRunner {
	q.sb.Offset(offset)
	return q
}

func (q *SQLRunner) Page(page, size int) *SQLRunner {
	q.sb.Offset((page - 1) * size).Limit(size)
	return q
}

func (q *SQLRunner) SetUpdateFields(fields []string) *SQLRunner {
	q.ub.SetUpdateFields(fields)
	return q
}

func (q *SQLRunner) Order(order string) *SQLRunner {
	q.sb.Order(order)
	return q
}
