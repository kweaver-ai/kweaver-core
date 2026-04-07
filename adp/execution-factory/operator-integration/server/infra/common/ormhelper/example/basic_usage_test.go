package example

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// TestBasicUsage 基本使用示例测试
func TestBasicUsage(t *testing.T) {
	// 创建模拟的数据库连接（在实际测试中，您可能需要使用真实的数据库或mock）
	var db *sql.DB
	orm := ormhelper.New(db, "example_db")

	// 由于没有真实数据库连接，这些测试主要验证SQL构建是否正确
	t.Run("Insert", func(t *testing.T) {
		testInsert(t, orm)
	})

	t.Run("Query", func(t *testing.T) {
		testQuery(t, orm)
	})

	t.Run("Update", func(t *testing.T) {
		testUpdate(t, orm)
	})

	t.Run("Delete", func(t *testing.T) {
		testDelete(t, orm)
	})

	t.Run("Pagination", func(t *testing.T) {
		testPagination(t, orm)
	})

	t.Run("Transaction", func(t *testing.T) {
		testTransaction(t, orm)
	})

	t.Run("ComplexQuery", func(t *testing.T) {
		testComplexQuery(t, orm)
	})
}

// testInsert 插入数据示例测试
func testInsert(t *testing.T, orm *ormhelper.DB) {
	// 测试单条插入SQL构建
	insertBuilder := orm.Insert().Into("users").Values(map[string]interface{}{
		"f_id":          "user-001",
		"f_name":        "张三",
		"f_email":       "zhangsan@example.com",
		"f_create_time": time.Now().Unix(),
	})

	query, args := insertBuilder.Build()
	if query == "" {
		t.Error("Insert query should not be empty")
	}
	if len(args) != 4 {
		t.Errorf("Expected 4 args, got %d", len(args))
	}
	t.Logf("Insert SQL: %s, Args: %v", query, args)

	// 测试批量插入SQL构建
	columns := []string{"f_id", "f_name", "f_email", "f_create_time"}
	values := [][]interface{}{
		{"user-002", "李四", "lisi@example.com", time.Now().Unix()},
		{"user-003", "王五", "wangwu@example.com", time.Now().Unix()},
	}
	batchBuilder := orm.Insert().Into("users").BatchValues(columns, values)
	query, args = batchBuilder.Build()
	if query == "" {
		t.Error("Batch insert query should not be empty")
	}
	t.Logf("Batch Insert SQL: %s, Args: %v", query, args)
}

// testQuery 查询数据示例测试
func testQuery(t *testing.T, orm *ormhelper.DB) {
	// 测试单条查询SQL构建
	selectBuilder := orm.Select().From("users").WhereEq("f_id", "user-001")
	query, args := selectBuilder.Build()
	if query == "" {
		t.Error("Select query should not be empty")
	}
	t.Logf("Select Single SQL: %s, Args: %v", query, args)

	// 测试多条查询SQL构建
	multiBuilder := orm.Select().From("users").
		WhereLike("f_name", "%张%").
		OrderByDesc("f_create_time").
		Limit(10)
	query, args = multiBuilder.Build()
	if query == "" {
		t.Error("Multi select query should not be empty")
	}
	t.Logf("Select Multi SQL: %s, Args: %v", query, args)

	// 测试统计查询SQL构建
	countBuilder := orm.Select().From("users").WhereEq("f_status", "active")
	query, args = countBuilder.Build()
	if query == "" {
		t.Error("Count query should not be empty")
	}
	t.Logf("Count SQL: %s, Args: %v", query, args)
}

// testUpdate 更新数据示例测试
func testUpdate(t *testing.T, orm *ormhelper.DB) {
	// 测试单字段更新SQL构建
	updateBuilder := orm.Update("users").
		Set("f_name", "张三丰").
		WhereEq("f_id", "user-001")
	query, args := updateBuilder.Build()
	if query == "" {
		t.Error("Update query should not be empty")
	}
	t.Logf("Update Single SQL: %s, Args: %v", query, args)

	// 测试批量更新SQL构建
	batchUpdateBuilder := orm.Update("users").SetData(map[string]interface{}{
		"f_status":      "inactive",
		"f_update_time": time.Now().Unix(),
	}).WhereLike("f_email", "%@old-domain.com")
	query, args = batchUpdateBuilder.Build()
	if query == "" {
		t.Error("Batch update query should not be empty")
	}
	t.Logf("Batch Update SQL: %s, Args: %v", query, args)

	// 测试字段自增SQL构建
	incrementBuilder := orm.Update("users").
		Increment("f_login_count", 1).
		WhereEq("f_id", "user-001")
	query, args = incrementBuilder.Build()
	if query == "" {
		t.Error("Increment query should not be empty")
	}
	t.Logf("Increment SQL: %s, Args: %v", query, args)
}

// testDelete 删除数据示例测试
func testDelete(t *testing.T, orm *ormhelper.DB) {
	// 测试单条删除SQL构建
	deleteBuilder := orm.Delete().From("users").WhereEq("f_id", "user-001")
	query, args := deleteBuilder.Build()
	if query == "" {
		t.Error("Delete query should not be empty")
	}
	t.Logf("Delete Single SQL: %s, Args: %v", query, args)

	// 测试批量删除SQL构建
	batchDeleteBuilder := orm.Delete().From("users").
		WhereEq("f_status", "inactive").
		WhereLt("f_create_time", time.Now().AddDate(0, 0, -30).Unix())
	query, args = batchDeleteBuilder.Build()
	if query == "" {
		t.Error("Batch delete query should not be empty")
	}
	t.Logf("Batch Delete SQL: %s, Args: %v", query, args)
}

// testPagination 分页查询示例测试
func testPagination(t *testing.T, orm *ormhelper.DB) {
	// 测试分页参数
	pagination := &ormhelper.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	// 测试排序参数
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: "f_create_time", Order: ormhelper.SortOrderDesc},
			{Field: "f_name", Order: ormhelper.SortOrderAsc},
		},
	}

	// 测试带分页和排序的查询SQL构建
	paginationBuilder := orm.Select().From("users").
		WhereEq("f_status", "active").
		Sort(sort).
		Pagination(pagination)
	query, args := paginationBuilder.Build()
	if query == "" {
		t.Error("Pagination query should not be empty")
	}
	t.Logf("Pagination SQL: %s, Args: %v", query, args)

	// 测试总数查询SQL构建
	countBuilder := orm.Select().From("users").WhereEq("f_status", "active")
	query, args = countBuilder.Build()
	if query == "" {
		t.Error("Count for pagination query should not be empty")
	}
	t.Logf("Count for Pagination SQL: %s, Args: %v", query, args)

	// 测试分页结果计算
	total := int64(100)
	result := ormhelper.CalculateQueryResult(total, pagination)
	if result.TotalPages != 10 {
		t.Errorf("Expected total pages 10, got %d", result.TotalPages)
	}
	t.Logf("Pagination Result: %+v", result)
}

// testTransaction 事务使用示例测试
func testTransaction(t *testing.T, orm *ormhelper.DB) {
	// 模拟事务
	var tx *sql.Tx // 在实际测试中，您可能需要创建真实的事务或使用mock

	// 测试事务中的ORM使用
	txOrm := orm.WithTx(tx)
	if txOrm == nil {
		t.Error("Transaction ORM should not be nil")
	}

	// 测试在事务中构建插入SQL
	insertBuilder := txOrm.Insert().Into("users").Values(map[string]interface{}{
		"f_id":   "user-tx-001",
		"f_name": "事务用户",
	})
	query, args := insertBuilder.Build()
	if query == "" {
		t.Error("Transaction insert query should not be empty")
	}
	t.Logf("Transaction Insert SQL: %s, Args: %v", query, args)

	// 测试在事务中构建更新SQL
	updateBuilder := txOrm.Update("users").
		Set("f_status", "verified").
		WhereEq("f_id", "user-tx-001")
	query, args = updateBuilder.Build()
	if query == "" {
		t.Error("Transaction update query should not be empty")
	}
	t.Logf("Transaction Update SQL: %s, Args: %v", query, args)
}

// testComplexQuery 复杂查询示例测试
func testComplexQuery(t *testing.T, orm *ormhelper.DB) {
	// 测试复杂查询SQL构建
	complexBuilder := orm.Select("u.f_id as f_user_id", "u.f_name as f_user_name", "p.f_id as f_profile_id", "p.f_avatar", "u.f_create_time").
		From("users u").
		LeftJoin("user_profiles p", "u.f_id = p.f_user_id").
		Where("u.f_status", "=", "active").
		And(func(w *ormhelper.WhereBuilder) {
			w.Gt("u.f_create_time", time.Now().AddDate(0, -1, 0).Unix()).
				Or(func(w2 *ormhelper.WhereBuilder) {
					w2.Eq("u.f_vip_level", "premium").
						Eq("u.f_verified", true)
				})
		}).
		OrderByDesc("u.f_create_time").
		Limit(50)

	query, args := complexBuilder.Build()
	if query == "" {
		t.Error("Complex query should not be empty")
	}
	t.Logf("Complex Query SQL: %s, Args: %v", query, args)

	// 验证SQL包含期望的关键字
	expectedKeywords := []string{"SELECT", "FROM", "LEFT JOIN", "WHERE", "ORDER BY", "LIMIT"}
	for _, keyword := range expectedKeywords {
		if !contains(query, keyword) {
			t.Errorf("Complex query should contain keyword: %s", keyword)
		}
	}
}

// contains 检查字符串是否包含子字符串（简单的辅助函数）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSQLBuilder 专门测试SQL构建器的功能
func TestSQLBuilder(t *testing.T) {
	var db *sql.DB
	orm := ormhelper.New(db, "test_db")

	t.Run("SelectBuilder", func(t *testing.T) {
		query, args := orm.Select("id", "name").
			From("users").
			WhereEq("status", "active").
			WhereGt("age", 18).
			OrderBy("name").
			Limit(10).
			Offset(20).
			Build()

		t.Logf("Select Query: %s", query)
		t.Logf("Select Args: %v", args)

		if query == "" {
			t.Error("Query should not be empty")
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
	})

	t.Run("UpdateBuilder", func(t *testing.T) {
		query, args := orm.Update("users").
			Set("name", "新名称").
			Set("age", 25).
			WhereEq("id", "user-001").
			Build()

		t.Logf("Update Query: %s", query)
		t.Logf("Update Args: %v", args)

		if query == "" {
			t.Error("Query should not be empty")
		}
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
	})

	t.Run("DeleteBuilder", func(t *testing.T) {
		query, args := orm.Delete().
			From("users").
			WhereEq("status", "inactive").
			WhereLt("last_login", time.Now().AddDate(0, 0, -30).Unix()).
			Build()

		t.Logf("Delete Query: %s", query)
		t.Logf("Delete Args: %v", args)

		if query == "" {
			t.Error("Query should not be empty")
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
	})
}
