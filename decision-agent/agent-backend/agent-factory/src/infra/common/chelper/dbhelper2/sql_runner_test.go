package dbhelper2

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/sqlhelper2"
	"go.uber.org/mock/gomock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/stretchr/testify/assert"
)

var (
	db      *sqlx.DB
	sqlMock sqlmock.Sqlmock
)

func init() {
	var err error

	db, sqlMock, err = sqlx.New()
	if err != nil {
		panic(err)
	}
}

func getMockLogger(ctrl *gomock.Controller) (logger icmp.Logger) {
	logger = cmpmock.NewMockLogger(ctrl)
	return
}

func TestQuery_NewQueryWithSqlBuilder(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	assert.Equal(t, sb, q.sb)
}

func TestQuery_Tag(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("json")
	assert.Equal(t, "json", q.tag)
}

func TestQuery_FindOne(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("json")

	var obj struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}

	sb.From("users").Select([]string{"key1", "key2"})

	sqlMock.ExpectQuery("select key1,key2 from users").
		WillReturnRows(sqlmock.NewRows([]string{"key1", "key2"}).
			AddRow("value1", "value2"))

	err := q.FindOne(&obj)
	assert.Nil(t, err)
	assert.Equal(t, "value1", obj.Key1)
	assert.Equal(t, "value2", obj.Key2)
}

// TestQuery_FindOne_selectPartFields 测试select部分字段
func TestQuery_FindOne_selectPartFields(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("json")

	var obj struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}

	sb.From("users").Select([]string{"key1"}).Where("id", sqlhelper2.OperatorEq, 1)

	sqlMock.ExpectQuery("select key1 from users where id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"key1"}).
			AddRow("value1"))

	err := q.FindOne(&obj)
	assert.Nil(t, err)
	assert.Equal(t, "value1", obj.Key1)
	assert.Equal(t, "", obj.Key2)
}

func TestQuery_Find(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("db")

	type user = struct {
		Key1 string `db:"key1"`
		Key2 string `db:"key2"`
	}

	objSlice := make([]user, 0)

	sb.From("users").Select([]string{"key1", "key2"})

	sqlMock.ExpectQuery("select key1,key2 from users").
		WillReturnRows(sqlmock.NewRows([]string{"key1", "key2"}).
			AddRow("value11", "value12").AddRow("value21", "value22"))

	err := q.Find(&objSlice)
	assert.Nil(t, err)
	assert.Equal(t, "value11", objSlice[0].Key1)
	assert.Equal(t, "value12", objSlice[0].Key2)
	assert.Equal(t, "value21", objSlice[1].Key1)
	assert.Equal(t, "value22", objSlice[1].Key2)
}

func TestQuery_Find_selectPartFields(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("db")

	type user = struct {
		Key1 string `db:"key1"`
		Key2 string `db:"key2"`
	}

	objSlice := make([]user, 0)

	sb.From("users").Select([]string{"key1"})

	sqlMock.ExpectQuery("select key1 from users").
		WillReturnRows(sqlmock.NewRows([]string{"key1"}).
			AddRow("value11").AddRow("value21"))

	err := q.Find(&objSlice)
	assert.Nil(t, err)
	assert.Equal(t, "value11", objSlice[0].Key1)
	assert.Equal(t, "", objSlice[0].Key2)
	assert.Equal(t, "value21", objSlice[1].Key1)
	assert.Equal(t, "", objSlice[1].Key2)
}

func TestQuery_NewQuery(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	assert.Equal(t, "db", q.tag)
}

func TestQuery_NewQuery_FullFeature(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1、part1
	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("json")

	type User struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}

	q.From("users").Select([]string{"key1", "key2"}).
		Where("id", sqlhelper2.OperatorEq, 1).
		Or("name", sqlhelper2.OperatorEq, "John").Offset(0).Limit(2)

	_sql, args, err := q.sb.ToSelectSQL()
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "select key1,key2 from users where id = ? or name = ? limit 2 offset 0", _sql)

	assert.Equal(t, []interface{}{1, "John"}, args)

	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"key1", "key2"}).
			AddRow("value1", "value2").AddRow("value3", "value4"))

	objSlice := make([]User, 0)

	err = q.Find(&objSlice)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "value1", objSlice[0].Key1)
	assert.Equal(t, "value2", objSlice[0].Key2)
	assert.Equal(t, "value3", objSlice[1].Key1)
	assert.Equal(t, "value4", objSlice[1].Key2)
}

func Test_FindColumn(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("json")

	// 1、字段为string类型
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"key1"}).
		AddRow("value1").
		AddRow("value2"))

	var key1s []string

	err := q.From("users").FindColumn("key1", &key1s)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []string{"value1", "value2"}, key1s)

	//  2、字段为其他类型

	// int32类型
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"key1"}).
		AddRow(1).
		AddRow(2))

	var key1s2 []int32

	err = q.From("users").FindColumn("key1", &key1s2)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []int32{1, 2}, key1s2)

	// int64类型
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"key1"}).
		AddRow(1).
		AddRow(2))

	var key1s3 []int64

	err = q.From("users").FindColumn("key1", &key1s3)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []int64{1, 2}, key1s3)

	//  int类型
	sqlMock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"key1"}).
		AddRow(1).
		AddRow(2))

	var key1s4 []int

	err = q.From("users").FindColumn("key1", &key1s4)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, []int{1, 2}, key1s4)
}

func Test_struct2ScanArgsByTag(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	type args struct {
		s    interface{}
		tags []string //nolint:unused
	}

	type UserInfo = struct {
		Age    int `db:"age"`
		Height int `db:"height"`
	}

	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "test1",
			args: args{
				s: &struct {
					ID   int    `db:"id"`
					Name string `db:"name"`
				}{},
			},
			want: []interface{}{new(int), new(string)},
		},
		{
			name: "包含匿名结构体",
			args: args{
				s: &struct {
					ID   int    `db:"id2"`
					Name string `db:"name2"`
					UserInfo
				}{},
			},
			want: []interface{}{new(int), new(string), new(int), new(int)},
		},
		{
			name: "包含匿名结构体指针",
			args: args{
				s: &struct {
					ID   int    `db:"id3"`
					Name string `db:"name3"`
					*UserInfo
				}{
					UserInfo: &UserInfo{},
				},
			},
			want: []interface{}{new(int), new(string), new(int), new(int)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
			if got := q.struct2ScanArgsByTag(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("struct2ScanArgsByTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_struct2ScanArgsByTag_WithOption(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	type args struct {
		s      interface{}
		fields []string
	}

	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "test1",
			args: args{
				s: &struct {
					ID   int    `db:"id"`
					Name string `db:"name"`
				}{},
				fields: []string{"id"},
			},
			want: []interface{}{new(int) /*, new(string)*/},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
			opt := Option{
				SelectFields: tt.args.fields,
			}
			if got := q.struct2ScanArgsByTag(tt.args.s, opt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("struct2ScanArgsByTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Exec(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. Update
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := q.Update(map[string]interface{}{"key1": "value1"})
	assert.Nil(t, err)

	// 2. UpdateByStruct
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = q.UpdateByStruct(struct {
		Key1 string `db:"key1"`
	}{Key1: "value1"})
	assert.Nil(t, err)

	//	3. Delete
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = q.From("user").
		NotIn("id", []int{1, 2, 3}).
		WhereEqual("key1", "value1").Delete()
	assert.Nil(t, err)

	//	4. Insert
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = q.Insert(map[string]interface{}{"key1": "value1"})
	assert.Nil(t, err)

	//	5. InsertStruct
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = q.InsertStruct(struct {
		Key1 string `db:"key1"`
	}{
		Key1: "value1",
	})
	assert.Nil(t, err)

	//	6. InsertStructs
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(0, 2))

	_, err = q.InsertStructs([]struct {
		Key1 string `db:"key1"`
	}{
		{
			Key1: "value1",
		},
		{
			Key1: "value2",
		},
	})

	assert.Nil(t, err)
}

func Test_Exists(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. 存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := q.Exists()
	assert.Nil(t, err)
	assert.True(t, exists)

	// 2. 不存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"})).WillReturnError(sql.ErrNoRows)

	exists, err = q.Exists()
	assert.Nil(t, err)
	assert.False(t, exists)
}

func Test_Count(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. 存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	count, err := q.Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	// 2. 不存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"})).WillReturnError(sql.ErrNoRows)

	count, err = q.Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), count)
}

func Test_Count_Raw(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. 存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	count, err := q.Raw("select count(*) from users").Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)

	// 2. 不存在
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"1"})).WillReturnError(sql.ErrNoRows)

	count, err = q.Raw("select count(*) from users").Count()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), count)
}

type userT struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (u *userT) TableName() string {
	return "users"
}

func Test_FromPo(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. 不使用FromPo，使用From，此时必须使用Select来指定查询的字段
	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "name1"))

	users := make([]userT, 0)
	err := q.From("users").
		Select([]string{"id", "name"}).
		Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))

	//	2. 使用FromPo，Select可选 但是如果不指定Select，那么就会查询所有字段
	q = NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")
	sqlMock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "name1"))

	users = make([]userT, 0)
	err = q.FromPo(&userT{}).
		Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func Test_InsertStructsInBatches(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. 插入成功
	//	1.1 第一次插入 2条
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(1, 1))
	//	1.2 第二次插入 1条
	sqlMock.ExpectExec("").
		WillReturnResult(sqlmock.NewResult(1, 1))

	type user = struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	users := make([]user, 0)
	users = append(users, user{ID: 1, Name: "name1"}, user{ID: 2, Name: "name2"}, user{ID: 3, Name: "name3"})

	err := q.InsertStructsInBatches(users, 2)
	assert.Nil(t, err)

	// 2. 插入失败
	sqlMock.ExpectExec("").
		WillReturnError(sql.ErrNoRows)

	err = q.InsertStructsInBatches(users, 2)
	assert.NotNil(t, err)
}

func Test_WhereByWhereBuilder(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. WhereByWhereBuilder
	sqlMock.ExpectQuery("select id,name from users where").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "name1"))

	users := make([]userT, 0)

	wb := sqlhelper2.NewWhereBuilder().
		WhereEqual("id", 1)
	err := q.FromPo(&userT{}).
		WhereByWhereBuilder(wb)

	assert.Nil(t, err)

	err = q.Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func Test_RawExists(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	// 1. 存在记录的情况
	sqlMock.ExpectQuery("select 1 from users where id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := q.RawExists("select 1 from users where id = ?", 1)
	assert.Nil(t, err)
	assert.True(t, exists)

	// 2. 不存在记录的情况
	sqlMock.ExpectQuery("select 1 from users where id = ?").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	exists, err = q.RawExists("select 1 from users where id = ?", 999)
	assert.Nil(t, err)
	assert.False(t, exists)

	// 3. 数据库错误的情况
	sqlMock.ExpectQuery("select 1 from users where id = ?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	exists, err = q.RawExists("select 1 from users where id = ?", 1)
	assert.NotNil(t, err)
	assert.False(t, exists)
	assert.Equal(t, sql.ErrConnDone, err)

	// 4. 返回值为0的情况
	sqlMock.ExpectQuery("select 0 from users where id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"0"}).AddRow(0))

	exists, err = q.RawExists("select 0 from users where id = ?", 1)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestSQLRunner_BuilderMethods(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("ResetSelect", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))
		sb := sqlhelper2.NewSelectBuilder()
		sb.Select([]string{"id", "name"})
		q.sb = sb

		result := q.ResetSelect()
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})

	t.Run("Page", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))

		result := q.Page(2, 10)
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})

	t.Run("SetUpdateFields", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))

		fields := []string{"name", "age"}
		result := q.SetUpdateFields(fields)
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})

	t.Run("Order", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))

		result := q.Order("id DESC")
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})

	t.Run("SetSQLBuilder", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))
		newSb := sqlhelper2.NewSelectBuilder()

		result := q.SetSQLBuilder(newSb)
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
		assert.Equal(t, newSb, q.sb)
	})

	t.Run("Limit", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))

		result := q.Limit(10)
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})

	t.Run("Offset", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q := NewSQLRunner(db, getMockLogger(ctrl))

		result := q.Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, q, result)
	})
}

func TestSQLRunner_TxSr(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx := &sql.Tx{}
	result := TxSr(tx, getMockLogger(ctrl))
	assert.NotNil(t, result)
}

func TestSQLRunner_SrByISQLRunner(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	result := SrByISQLRunner(db, getMockLogger(ctrl))
	assert.NotNil(t, result)
}

func TestSQLRunner_WhereNotEqual(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	result := q.WhereNotEqual("status", "deleted")
	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestSQLRunner_OrEqual(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	result := q.OrEqual("name", "admin")
	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestSQLRunner_In(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	result := q.In("id", []int{1, 2, 3})
	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestSQLRunner_Like(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	result := q.Like("name", "%test%")
	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestSQLRunner_WhereRaw(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	result := q.WhereRaw("id > ? AND status = ?", 100, "active")
	assert.NotNil(t, result)
	assert.Equal(t, q, result)
}

func TestSQLRunner_WhereByWhereBuilderOr(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))
	q.Tag("db")

	sqlMock.ExpectQuery("select id,name from users where").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "name1"))

	users := make([]userT, 0)

	wb := sqlhelper2.NewWhereBuilder().
		WhereEqual("id", 1)
	err := q.FromPo(&userT{}).
		WhereByWhereBuilderOr(wb)

	assert.Nil(t, err)

	err = q.Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestSQLRunner_RawExec(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	sqlMock.ExpectExec("DELETE FROM users WHERE id = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := q.RawExec("DELETE FROM users WHERE id = ?", 1)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	rowsAffected, err := res.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

func TestSQLRunner_Scan(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	sqlMock.ExpectQuery("select id,name from users where id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test"))

	var id int

	var name string

	err := q.Select([]string{"id", "name"}).
		From("users").
		Where("id", sqlhelper2.OperatorEq, 1).
		Scan(&id, &name)

	assert.Nil(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "test", name)
}

func TestSQLRunner_Scan_PanicWithoutSelect(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	var id int

	var name string

	assert.Panics(t, func() {
		_ = q.From("users").
			Where("id", sqlhelper2.OperatorEq, 1).
			Scan(&id, &name)
	})
}

func TestSQLRunner_Scan_QueryError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	sqlMock.ExpectQuery("select id,name from users where id = ?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	var id int

	var name string

	err := q.Select([]string{"id", "name"}).
		From("users").
		Where("id", sqlhelper2.OperatorEq, 1).
		Scan(&id, &name)

	assert.NotNil(t, err)
}

func TestSQLRunner_FindOne_QueryError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("json")

	var obj struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}

	sb.From("users").Select([]string{"key1", "key2"})

	sqlMock.ExpectQuery("select key1,key2 from users").
		WillReturnError(sql.ErrConnDone)

	err := q.FindOne(&obj)
	assert.NotNil(t, err)
}

func TestSQLRunner_Find_QueryError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("db")

	type user = struct {
		Key1 string `db:"key1"`
		Key2 string `db:"key2"`
	}

	objSlice := make([]user, 0)

	sb.From("users").Select([]string{"key1", "key2"})

	sqlMock.ExpectQuery("select key1,key2 from users").
		WillReturnError(sql.ErrConnDone)

	err := q.Find(&objSlice)
	assert.NotNil(t, err)
}

func TestSQLRunner_Find_ScanError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sb := sqlhelper2.NewSelectBuilder()
	q := NewQueryWithSQLBuilder(db, sb, getMockLogger(ctrl))
	q.Tag("db")

	type user = struct {
		Key1 string `db:"key1"`
		Key2 int    `db:"key2"`
	}

	objSlice := make([]user, 0)

	sb.From("users").Select([]string{"key1", "key2"})

	// Return a string for an int field to cause Scan error
	sqlMock.ExpectQuery("select key1,key2 from users").
		WillReturnRows(sqlmock.NewRows([]string{"key1", "key2"}).
			AddRow("value1", "not-an-int"))

	err := q.Find(&objSlice)
	assert.NotNil(t, err)
}

func TestSQLRunner_FindColumn_QueryError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	sqlMock.ExpectQuery("select key1 from users").
		WillReturnError(sql.ErrConnDone)

	var key1s []string

	err := q.From("users").FindColumn("key1", &key1s)
	assert.NotNil(t, err)
}

func TestSQLRunner_FindColumn_ScanError(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	// Return string for int field to cause Scan error
	sqlMock.ExpectQuery("select key1 from users").
		WillReturnRows(sqlmock.NewRows([]string{"key1"}).
			AddRow("not-an-int"))

	var key1s []int

	err := q.From("users").FindColumn("key1", &key1s)
	assert.NotNil(t, err)
}

func TestSQLRunner_Raw(t *testing.T) {
	// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	q := NewSQLRunner(db, getMockLogger(ctrl))

	// Test Raw with args
	t.Run("with args", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q2 := q.Raw("SELECT * FROM users WHERE id = ?", 123)
		assert.NotNil(t, q2)
		assert.Equal(t, "SELECT * FROM users WHERE id = ?", q2.rawSQL)
		assert.Len(t, q2.rawSQLArgs, 1)
		assert.Equal(t, 123, q2.rawSQLArgs[0])
	})

	// Test Raw without args
	t.Run("without args", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q2 := q.Raw("SELECT * FROM users")
		assert.NotNil(t, q2)
		assert.Equal(t, "SELECT * FROM users", q2.rawSQL)
		assert.Empty(t, q2.rawSQLArgs)
	})

	// Test Raw with multiple args
	t.Run("with multiple args", func(t *testing.T) {
		// 不使用 t.Parallel(): 共享包级别 sqlMock 实例，不能并行
		q2 := q.Raw("SELECT * FROM users WHERE id = ? AND status = ?", 123, "active")
		assert.NotNil(t, q2)
		assert.Equal(t, "SELECT * FROM users WHERE id = ? AND status = ?", q2.rawSQL)
		assert.Len(t, q2.rawSQLArgs, 2)
		assert.Equal(t, 123, q2.rawSQLArgs[0])
		assert.Equal(t, "active", q2.rawSQLArgs[1])
	})
}
