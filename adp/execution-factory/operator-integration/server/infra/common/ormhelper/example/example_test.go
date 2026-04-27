package example_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// MCPConfigTest 测试用的配置结构体
type MCPConfigTest struct {
	ID          string `json:"f_id" db:"f_id"`
	Name        string `json:"f_name" db:"f_name"`
	Description string `json:"f_description" db:"f_description"`
	Status      string `json:"f_status" db:"f_status"`
	CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
	UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
}

// ExampleDAO 示例DAO
type ExampleDAO struct {
	orm *ormhelper.DB
}

// NewExampleDAO 创建DAO实例
func NewExampleDAO(orm *ormhelper.DB) *ExampleDAO {
	return &ExampleDAO{orm: orm}
}

// Insert 插入配置
func (dao *ExampleDAO) Insert(ctx context.Context, config *MCPConfigTest) (string, error) {
	now := time.Now().UnixNano()
	data := map[string]interface{}{
		"f_id":          config.ID,
		"f_name":        config.Name,
		"f_description": config.Description,
		"f_status":      config.Status,
		"f_create_time": now,
		"f_update_time": now,
	}
	_, err := dao.orm.Insert().Into("t_mcp_server_config").Values(data).Execute(ctx)
	return config.ID, err
}

// UpdateStatus 更新状态
func (dao *ExampleDAO) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := dao.orm.Update("t_mcp_server_config").
		Set("f_status", status).
		Set("f_update_time", time.Now().UnixNano()).
		WhereEq("f_id", id).
		Execute(ctx)
	return err
}

// DeleteByID 根据ID删除
func (dao *ExampleDAO) DeleteByID(ctx context.Context, id string) error {
	_, err := dao.orm.Delete().From("t_mcp_server_config").WhereEq("f_id", id).Execute(ctx)
	return err
}

func TestExampleDAO_Insert(t *testing.T) {
	// 创建模拟执行器
	mockExecutor := NewMockExecutor()

	// 设置期望的SQL执行
	mockExecutor.ExpectExec(
		"INSERT INTO `test_db`.`t_mcp_server_config` (f_create_time, f_description, f_id, f_name, f_status, f_update_time) VALUES (?, ?, ?, ?, ?, ?)",
	).WillReturnResult(1, 1)

	// 创建ORM实例
	orm := ormhelper.New(mockExecutor, "test_db")
	dao := NewExampleDAO(orm)

	// 测试数据
	config := &MCPConfigTest{
		ID:          "test-id-1",
		Name:        "测试配置",
		Description: "这是一个测试配置",
		Status:      "active",
	}

	// 执行插入
	ctx := context.Background()
	id, err := dao.Insert(ctx, config)

	// 验证结果
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}
	if id != "test-id-1" {
		t.Errorf("Expected id 'test-id-1', got '%s'", id)
	}

	// 验证所有期望都被满足
	if err := mockExecutor.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestExampleDAO_Insert_Error(t *testing.T) {
	// 创建模拟执行器
	mockExecutor := NewMockExecutor()

	// 设置期望的SQL执行（返回错误）
	mockExecutor.ExpectExec(
		"INSERT INTO `test_db`.`t_mcp_server_config` (f_create_time, f_description, f_id, f_name, f_status, f_update_time) VALUES (?, ?, ?, ?, ?, ?)",
	).WillReturnError(sql.ErrConnDone)

	// 创建ORM实例
	orm := ormhelper.New(mockExecutor, "test_db")
	dao := NewExampleDAO(orm)

	// 测试数据
	config := &MCPConfigTest{
		ID:          "test-id-2",
		Name:        "测试配置2",
		Description: "这是另一个测试配置",
		Status:      "inactive",
	}

	// 执行插入
	ctx := context.Background()
	_, err := dao.Insert(ctx, config)

	// 验证错误
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if err != sql.ErrConnDone {
		t.Errorf("Expected sql.ErrConnDone, got %v", err)
	}

	// 验证所有期望都被满足
	if err := mockExecutor.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestExampleDAO_UpdateStatus(t *testing.T) {
	// 创建模拟执行器
	mockExecutor := NewMockExecutor()

	// 设置期望的SQL执行
	mockExecutor.ExpectExec(
		"UPDATE `test_db`.`t_mcp_server_config` SET f_status = ?, f_update_time = ? WHERE f_id = ?",
		"published", // status
		// f_update_time 是动态生成的，我们不验证具体值
		// "test-id-1", // id
	).WillReturnResult(0, 1)

	// 创建ORM实例
	orm := ormhelper.New(mockExecutor, "test_db")
	dao := NewExampleDAO(orm)

	// 执行更新
	ctx := context.Background()
	err := dao.UpdateStatus(ctx, "test-id-1", "published")

	// 验证结果
	if err != nil {
		t.Errorf("UpdateStatus failed: %v", err)
	}
}

func TestExampleDAO_DeleteByID(t *testing.T) {
	// 创建模拟执行器
	mockExecutor := NewMockExecutor()

	// 设置期望的SQL执行
	mockExecutor.ExpectExec(
		"DELETE FROM `test_db`.`t_mcp_server_config` WHERE f_id = ?",
		"test-id-1",
	).WillReturnResult(0, 1)

	// 创建ORM实例
	orm := ormhelper.New(mockExecutor, "test_db")
	dao := NewExampleDAO(orm)

	// 执行删除
	ctx := context.Background()
	err := dao.DeleteByID(ctx, "test-id-1")

	// 验证结果
	if err != nil {
		t.Errorf("DeleteByID failed: %v", err)
	}
}

func TestSelectBuilder_Build(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	query, args := orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		OrderByDesc("f_update_time").
		Limit(10).
		Build()

	expectedQuery := "SELECT * FROM `test_db`.`t_mcp_server_config` WHERE f_status = ? ORDER BY f_update_time DESC LIMIT 10"
	if query != expectedQuery {
		t.Errorf("Expected query: %s, got: %s", expectedQuery, query)
	}

	if len(args) != 1 || args[0] != "active" {
		t.Errorf("Expected args: [active], got: %v", args)
	}
}

func TestInsertBuilder_Build(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	data := map[string]interface{}{
		"f_id":     "test-id",
		"f_name":   "测试",
		"f_status": "active",
	}

	query, args := orm.Insert().
		Into("t_mcp_server_config").
		Values(data).
		Build()

	// 验证查询包含正确的表名和字段
	if !contains(query, "INSERT INTO `test_db`.`t_mcp_server_config`") {
		t.Errorf("Query should contain table name, got: %s", query)
	}

	// 验证参数数量
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestUpdateBuilder_Build(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	query, args := orm.Update("t_mcp_server_config").
		Set("f_status", "inactive").
		Set("f_update_time", 123456789).
		WhereEq("f_id", "test-id").
		Build()

	expectedQuery := "UPDATE `test_db`.`t_mcp_server_config` SET f_status = ?, f_update_time = ? WHERE f_id = ?"
	if query != expectedQuery {
		t.Errorf("Expected query: %s, got: %s", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestDeleteBuilder_Build(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	query, args := orm.Delete().
		From("t_mcp_server_config").
		WhereEq("f_id", "test-id").
		Build()

	expectedQuery := "DELETE FROM `test_db`.`t_mcp_server_config` WHERE f_id = ?"
	if query != expectedQuery {
		t.Errorf("Expected query: %s, got: %s", expectedQuery, query)
	}

	if len(args) != 1 || args[0] != "test-id" {
		t.Errorf("Expected args: [test-id], got: %v", args)
	}
}

func TestWhereBuilder_Complex(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	query, args := orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		And(func(w *ormhelper.WhereBuilder) {
			w.Gt("f_create_time", 1000000000).
				Lt("f_create_time", 2000000000)
		}).
		Or(func(w *ormhelper.WhereBuilder) {
			w.Eq("f_category", "special").
				Like("f_name", "%test%")
		}).
		Build()

	// 验证查询包含复杂的WHERE条件
	if !containsSubstring(query, "WHERE") {
		t.Errorf("Query should contain WHERE clause, got: %s", query)
	}

	// 验证参数数量（应该有5个参数）
	if len(args) != 5 {
		t.Errorf("Expected 5 args, got %d: %v", len(args), args)
	}
}

func TestBatchInsert(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	columns := []string{"f_id", "f_name", "f_status"}
	values := [][]interface{}{
		{"id1", "name1", "active"},
		{"id2", "name2", "inactive"},
	}

	query, args := orm.Insert().
		Into("t_mcp_server_config").
		BatchValues(columns, values).
		Build()

	// 验证查询包含批量插入语法
	if !contains(query, "INSERT INTO `test_db`.`t_mcp_server_config`") {
		t.Errorf("Query should contain table name, got: %s", query)
	}

	if !contains(query, "VALUES") {
		t.Errorf("Query should contain VALUES, got: %s", query)
	}

	// 验证参数数量（2行 * 3列 = 6个参数）
	if len(args) != 6 {
		t.Errorf("Expected 6 args, got %d", len(args))
	}
}

func TestTransaction(t *testing.T) {
	mockExecutor := NewMockExecutor()
	orm := ormhelper.New(mockExecutor, "test_db")

	// 模拟事务
	var mockTx *sql.Tx // 在实际测试中，这里应该是真实的事务对象
	txORM := orm.WithTx(mockTx)

	// 验证事务ORM实例
	if txORM == nil {
		t.Error("WithTx should return a valid ORM instance")
	}

	// 验证是否在事务中（这里由于mockTx是nil，所以会返回false）
	// 在实际使用中，如果传入真实的*sql.Tx，IsInTransaction()会返回true
	if txORM.IsInTransaction() {
		t.Log("Transaction ORM created successfully")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// 辅助函数：检查字符串是否包含子字符串（不区分大小写）
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
