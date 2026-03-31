// Package example 一个简单的使用示例
package example

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

var (
	defaultLimit = 10
)

// BasicUsageExample 基本使用示例
func BasicUsageExample() {
	// 假设你已经有了数据库连接
	var orm *ormhelper.DB
	db, _ := sql.Open("mysql", "dsn")

	// 这里使用nil作为示例，实际使用时请传入真实的数据库连接
	// var db *sql.DB
	orm = ormhelper.New(db, "example_db")

	ctx := context.Background()

	// 1. 插入数据示例
	insertExample(ctx, orm)

	// 2. 查询数据示例
	queryExample(ctx, orm)

	// 3. 更新数据示例
	updateExample(ctx, orm)

	// 4. 删除数据示例
	deleteExample(ctx, orm)

	// 5. 分页查询示例
	paginationExample(ctx, orm)

	// 6. 事务使用示例
	transactionExample(ctx, orm)
}

// 插入数据示例
func insertExample(ctx context.Context, orm *ormhelper.DB) {
	// 单条插入
	_, err := orm.Insert().Into("users").Values(map[string]interface{}{
		"f_id":          "user-001",
		"f_name":        "张三",
		"f_email":       "zhangsan@example.com",
		"f_create_time": time.Now().Unix(),
	}).Execute(ctx)
	if err != nil { // 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 批量插入
	columns := []string{"f_id", "f_name", "f_email", "f_create_time"}
	values := [][]interface{}{
		{"user-002", "李四", "lisi@example.com", time.Now().Unix()},
		{"user-003", "王五", "wangwu@example.com", time.Now().Unix()},
	}
	_, err = orm.Insert().Into("users").BatchValues(columns, values).Execute(ctx)
	if err != nil { // 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}
}

// 查询数据示例
func queryExample(ctx context.Context, orm *ormhelper.DB) {
	// 定义结果结构体
	type User struct {
		ID         string `db:"f_id"`
		Name       string `db:"f_name"`
		Email      string `db:"f_email"`
		CreateTime int64  `db:"f_create_time"`
	}

	// 查询单条记录
	var user User
	err := orm.Select().From("users").WhereEq("f_id", "user-001").First(ctx, &user)
	if err != nil { // 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 查询多条记录
	var users []*User
	err = orm.Select().From("users").
		WhereLike("f_name", "%张%").
		OrderByDesc("f_create_time").
		Limit(defaultLimit).
		Get(ctx, &users)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 统计数量
	count, err := orm.Select().From("users").WhereEq("f_status", "active").Count(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}
	_ = count
}

// 更新数据示例
func updateExample(ctx context.Context, orm *ormhelper.DB) {
	// 更新单个字段
	_, err := orm.Update("users").
		Set("f_name", "张三丰").
		WhereEq("f_id", "user-001").
		Execute(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 批量更新
	_, err = orm.Update("users").SetData(map[string]interface{}{
		"f_status":      "inactive",
		"f_update_time": time.Now().Unix(),
	}).WhereLike("f_email", "%@old-domain.com").Execute(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 字段自增
	_, err = orm.Update("users").
		Increment("f_login_count", 1).
		WhereEq("f_id", "user-001").
		Execute(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}
}

// 删除数据示例
func deleteExample(ctx context.Context, orm *ormhelper.DB) {
	// 删除单条记录
	_, err := orm.Delete().From("users").WhereEq("f_id", "user-001").Execute(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 批量删除
	_, err = orm.Delete().From("users").
		WhereEq("f_status", "inactive").
		WhereLt("f_create_time", time.Now().AddDate(0, 0, -30).Unix()).
		Execute(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}
}

// 分页查询示例
func paginationExample(ctx context.Context, orm *ormhelper.DB) {
	type User struct {
		ID         string `db:"f_id"`
		Name       string `db:"f_name"`
		Email      string `db:"f_email"`
		CreateTime int64  `db:"f_create_time"`
	}

	// 使用分页参数
	pagination := &ormhelper.PaginationParams{
		Page:     1,
		PageSize: defaultLimit,
	}

	// 使用排序参数
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: "f_create_time", Order: ormhelper.SortOrderDesc},
			{Field: "f_name", Order: ormhelper.SortOrderAsc},
		},
	}

	var users []*User
	err := orm.Select().From("users").
		WhereEq("f_status", "active").
		Sort(sort).
		Pagination(pagination).
		Get(ctx, &users)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 获取总数用于分页计算
	total, err := orm.Select().From("users").WhereEq("f_status", "active").Count(ctx)
	if err != nil {
		// 处理错误
		fmt.Printf("err : %s\n", err)
		return
	}

	// 计算分页信息
	result := ormhelper.CalculateQueryResult(total, pagination)
	_ = result
}

// 事务使用示例
func transactionExample(ctx context.Context, orm *ormhelper.DB) {
	// 假设你有一个事务
	var tx *sql.Tx // 这里使用nil作为示例

	// 在事务中使用ORM
	txOrm := orm.WithTx(tx)

	// 在事务中执行操作
	_, err := txOrm.Insert().Into("users").Values(map[string]interface{}{
		"f_id":   "user-tx-001",
		"f_name": "事务用户",
	}).Execute(ctx)
	if err != nil {
		// 回滚事务
		_ = tx.Rollback()
		return
	}

	_, err = txOrm.Update("users").
		Set("f_status", "verified").
		WhereEq("f_id", "user-tx-001").
		Execute(ctx)
	if err != nil {
		// 回滚事务
		_ = tx.Rollback()
		return
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		fmt.Printf("err : %s\n", err)
		return
	}
}

// 复杂查询示例
// func complexQueryExample(ctx context.Context, orm *ormhelper.DB) {
// 	type UserProfile struct {
// 		UserID     string `db:"f_user_id"`
// 		UserName   string `db:"f_user_name"`
// 		ProfileID  string `db:"f_profile_id"`
// 		Avatar     string `db:"f_avatar"`
// 		CreateTime int64  `db:"f_create_time"`
// 	}

// 	var profiles []*UserProfile
// 	err := orm.Select("u.f_id as f_user_id", "u.f_name as f_user_name", "p.f_id as f_profile_id", "p.f_avatar", "u.f_create_time").
// 		From("users u").
// 		LeftJoin("user_profiles p", "u.f_id = p.f_user_id").
// 		Where("u.f_status", "=", "active").
// 		And(func(w *ormhelper.WhereBuilder) {
// 			w.Gt("u.f_create_time", time.Now().AddDate(0, -1, 0).Unix()).
// 				Or(func(w2 *ormhelper.WhereBuilder) {
// 					w2.Eq("u.f_vip_level", "premium").
// 						Eq("u.f_verified", true)
// 				})
// 		}).
// 		OrderByDesc("u.f_create_time").
// 		Limit(50).
// 		Get(ctx, &profiles)
// 	if err != nil {
// 		// 处理错误
// 	}
// }
