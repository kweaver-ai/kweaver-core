package dbhelper2

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 定义一些测试用的自定义字符串类型
type (
	UserID       string
	OrderStatus  string
	CategoryName string
)

func TestFillSQL(t *testing.T) {
	t.Parallel()

	// 1. 正常情况 - 混合类型参数
	sql := "select * from user where id = ? and name = ?"
	args := []interface{}{1, "test"}
	filledSql := FillSQL(sql, args...)
	assert.Equal(t, "select * from user where id = 1 and name = 'test'", filledSql)

	// 2. 字符串类型参数
	t.Run("字符串参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where name = ? and email = ?"
		args := []interface{}{"张三", "test@example.com"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where name = '张三' and email = 'test@example.com'", filledSql)
	})

	// 3. 数字类型参数
	t.Run("数字参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where id = ? and age = ? and salary = ? and score = ?"
		args := []interface{}{123, int64(25), 5000.50, float32(98.5)}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where id = 123 and age = 25 and salary = 5000.5 and score = 98.5", filledSql)
	})

	// 4. 布尔类型参数
	t.Run("布尔参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where is_active = ? and is_deleted = ?"
		args := []interface{}{true, false}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where is_active = true and is_deleted = false", filledSql)
	})

	// 5. 时间类型参数
	t.Run("时间参数", func(t *testing.T) {
		t.Parallel()

		testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		sql := "select * from user where created_at = ?"
		args := []interface{}{testTime}
		filledSql := FillSQL(sql, args...)
		expected := "select * from user where created_at = " + testTime.String()
		assert.Equal(t, expected, filledSql)
	})

	// 6. nil 参数 - 应该panic
	t.Run("nil参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where deleted_at = ?"
		args := []interface{}{nil}

		assert.Panics(t, func() {
			FillSQL(sql, args...)
		}, "传入nil参数应该panic")
	})

	// 7. 空字符串参数
	t.Run("空字符串参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where name = ? and description = ?"
		args := []interface{}{"", "non-empty"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where name = '' and description = 'non-empty'", filledSql)
	})

	// 8. 没有参数的情况
	t.Run("无参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user"
		filledSql := FillSQL(sql)
		assert.Equal(t, "select * from user", filledSql)
	})

	// 9. 没有占位符但有参数的情况
	t.Run("无占位符但有参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user"
		args := []interface{}{1, "test"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user", filledSql) // 参数被忽略
	})

	// 10. 占位符多于参数的情况
	t.Run("占位符多于参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where id = ? and name = ? and age = ?"
		args := []interface{}{1, "test"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where id = 1 and name = 'test' and age = ?", filledSql)
	})

	// 11. 参数多于占位符的情况
	t.Run("参数多于占位符", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where id = ?"
		args := []interface{}{1, "test", 25}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where id = 1", filledSql) // 多余参数被忽略
	})

	// 12. 包含特殊字符的字符串
	t.Run("特殊字符字符串", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where name = ? and comment = ?"
		args := []interface{}{"O'Connor", "It's a \"test\" string"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where name = 'O'Connor' and comment = 'It's a \"test\" string'", filledSql)
	})

	// 13. 零值参数
	t.Run("零值参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where id = ? and count = ? and rate = ?"
		args := []interface{}{0, int64(0), 0.0}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where id = 0 and count = 0 and rate = 0", filledSql)
	})

	// 14. 结构体参数
	t.Run("结构体参数", func(t *testing.T) {
		t.Parallel()

		type User struct {
			ID   int
			Name string
		}

		user := User{ID: 1, Name: "test"}
		sql := "select * from user where data = ?"
		args := []interface{}{user}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where data = {1 test}", filledSql)
	})

	// 15. 切片参数
	t.Run("切片参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where ids = ?"
		args := []interface{}{[]int{1, 2, 3}}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where ids = [1 2 3]", filledSql)
	})

	// 16. 空 SQL 语句
	t.Run("空SQL语句", func(t *testing.T) {
		t.Parallel()

		sql := ""
		args := []interface{}{1, "test"}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "", filledSql)
	})

	// 17. 复杂 SQL 语句
	t.Run("复杂SQL语句", func(t *testing.T) {
		t.Parallel()

		sql := `INSERT INTO user (id, name, email, age, is_active, created_at) 
				VALUES (?, ?, ?, ?, ?, ?) 
				ON DUPLICATE KEY UPDATE name = ?, email = ?`
		testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		args := []interface{}{1, "张三", "zhangsan@example.com", 25, true, testTime, "李四", "lisi@example.com"}
		filledSql := FillSQL(sql, args...)
		expected := fmt.Sprintf(`INSERT INTO user (id, name, email, age, is_active, created_at) 
				VALUES (1, '张三', 'zhangsan@example.com', 25, true, %s) 
				ON DUPLICATE KEY UPDATE name = '李四', email = 'lisi@example.com'`, testTime.String())
		assert.Equal(t, expected, filledSql)
	})

	// 18. 负数参数
	t.Run("负数参数", func(t *testing.T) {
		t.Parallel()

		sql := "select * from user where balance = ? and score = ?"
		args := []interface{}{-100, -99.5}
		filledSql := FillSQL(sql, args...)
		assert.Equal(t, "select * from user where balance = -100 and score = -99.5", filledSql)
	})

	// 19. 自定义字符串类型测试
	t.Run("自定义字符串类型", func(t *testing.T) {
		t.Parallel()
		// UserID 类型
		sql1 := "select * from users where id = ?"
		args1 := []interface{}{UserID("user123")}
		filledSql1 := FillSQL(sql1, args1...)
		assert.Equal(t, "select * from users where id = 'user123'", filledSql1)

		// OrderStatus 类型
		sql2 := "select * from orders where status = ?"
		args2 := []interface{}{OrderStatus("pending")}
		filledSql2 := FillSQL(sql2, args2...)
		assert.Equal(t, "select * from orders where status = 'pending'", filledSql2)

		// 混合使用自定义字符串类型和其他类型
		sql3 := "select * from orders where user_id = ? and status = ? and amount > ?"
		args3 := []interface{}{UserID("user456"), OrderStatus("completed"), 100}
		filledSql3 := FillSQL(sql3, args3...)
		assert.Equal(t, "select * from orders where user_id = 'user456' and status = 'completed' and amount > 100", filledSql3)

		// 多个自定义字符串类型
		sql4 := "select * from products where category = ? and vendor_id = ?"
		args4 := []interface{}{CategoryName("electronics"), UserID("vendor789")}
		filledSql4 := FillSQL(sql4, args4...)
		assert.Equal(t, "select * from products where category = 'electronics' and vendor_id = 'vendor789'", filledSql4)
	})
}

func TestGenInClauseGeneric(t *testing.T) {
	t.Parallel()

	// 1. 字符串类型测试
	t.Run("字符串类型", func(t *testing.T) {
		t.Parallel()

		args := []string{"apple", "banana", "orange"}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "'apple','banana','orange'", inClause)
	})

	// 2. 整数类型测试
	t.Run("整数类型", func(t *testing.T) {
		t.Parallel()

		args := []int{1, 2, 3, 4, 5}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "1,2,3,4,5", inClause)
	})

	// 3. 浮点数类型测试
	t.Run("浮点数类型", func(t *testing.T) {
		t.Parallel()

		args := []float64{1.1, 2.2, 3.3}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "1.1,2.2,3.3", inClause)
	})

	// 4. 布尔类型测试
	t.Run("布尔类型", func(t *testing.T) {
		t.Parallel()

		args := []bool{true, false, true}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "true,false,true", inClause)
	})

	// 5. 混合类型结构体测试
	t.Run("结构体类型", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			ID   int
			Name string
		}

		args := []TestStruct{{1, "test1"}, {2, "test2"}}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "{1 test1},{2 test2}", inClause)
	})

	// 6. 空切片测试
	t.Run("空切片", func(t *testing.T) {
		t.Parallel()

		var args []string
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "", inClause)
	})

	// 7. 单元素测试
	t.Run("单元素", func(t *testing.T) {
		t.Parallel()

		args := []int{42}
		inClause := GenInClauseGeneric(args)
		assert.Equal(t, "42", inClause)
	})

	// 8. 自定义字符串类型测试
	t.Run("自定义字符串类型", func(t *testing.T) {
		t.Parallel()
		// UserID 类型
		userIds := []UserID{UserID("user1"), UserID("user2"), UserID("user3")}
		inClause1 := GenInClauseGeneric(userIds)
		assert.Equal(t, "'user1','user2','user3'", inClause1)

		// OrderStatus 类型
		statuses := []OrderStatus{OrderStatus("pending"), OrderStatus("completed"), OrderStatus("cancelled")}
		inClause2 := GenInClauseGeneric(statuses)
		assert.Equal(t, "'pending','completed','cancelled'", inClause2)

		// CategoryName 类型
		categories := []CategoryName{CategoryName("electronics"), CategoryName("books")}
		inClause3 := GenInClauseGeneric(categories)
		assert.Equal(t, "'electronics','books'", inClause3)

		// 单个自定义字符串类型元素
		singleUser := []UserID{UserID("singleuser")}
		inClause4 := GenInClauseGeneric(singleUser)
		assert.Equal(t, "'singleuser'", inClause4)

		// 空的自定义字符串类型切片
		var emptyUsers []UserID
		inClause5 := GenInClauseGeneric(emptyUsers)
		assert.Equal(t, "", inClause5)
	})
}
