package example

import (
	"context"
	"database/sql"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// 示例：如何在现有项目中使用ORM Helper

// ExampleUsage 展示ORM Helper的基本用法
func ExampleUsage() {
	// 1. 初始化ORM Helper
	// 假设你已经有了数据库连接池 dbPool 和数据库名 dbName
	var dbPool ormhelper.Executor // 你的数据库连接池，实现了Executor接口
	dbName := "your_database_name"

	orm := ormhelper.New(dbPool, dbName)

	// 2. 基本查询操作
	ctx := context.Background()

	// SELECT查询示例
	var configs []MCPConfigExample
	err := orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		OrderByDesc("f_update_time").
		Limit(10). //nolint:mnd
		Get(ctx, &configs)
	if err != nil {
		return
	}

	// 单个记录查询
	var config MCPConfigExample
	err = orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_id", "some-id").
		First(ctx, &config)
	if err != nil {
		return
	}

	// 统计数量
	count, err := orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		Count(ctx)
	if err != nil {
		return
	}
	// 使用count变量的示例（在实际项目中可以用于返回给前端或日志）
	_ = count // 这里可以使用count变量，比如：fmt.Printf("找到 %d 条记录\n", count)

	// 3. 插入操作
	data := map[string]interface{}{
		"f_id":          "new-config-id",
		"f_name":        "新配置",
		"f_description": "这是一个新的配置",
		"f_status":      "active",
		"f_create_time": time.Now().UnixNano(),
		"f_update_time": time.Now().UnixNano(),
	}

	result, err := orm.Insert().
		Into("t_mcp_server_config").
		Values(data).
		Execute(ctx)
	if err != nil {
		return
	}

	// 获取插入的ID（如果是自增主键）
	lastID, _ := result.LastInsertId()
	_ = lastID

	// 4. 批量插入
	columns := []string{"f_id", "f_name", "f_status", "f_create_time", "f_update_time"}
	now := time.Now().UnixNano()
	values := [][]interface{}{
		{"config-1", "配置1", "active", now, now},
		{"config-2", "配置2", "active", now, now},
		{"config-3", "配置3", "inactive", now, now},
	}

	_, err = orm.Insert().
		Into("t_mcp_server_config").
		BatchValues(columns, values).
		Execute(ctx)
	if err != nil {
		return
	}

	// 5. 更新操作
	_, err = orm.Update("t_mcp_server_config").
		Set("f_status", "inactive").
		Set("f_update_time", time.Now().UnixNano()).
		WhereEq("f_id", "some-id").
		Execute(ctx)
	if err != nil {
		return
	}

	// 6. 删除操作
	_, err = orm.Delete().
		From("t_mcp_server_config").
		WhereEq("f_id", "some-id").
		Execute(ctx)
	if err != nil {
		return
	}
}

// ExampleTransactionUsage 展示事务使用方法
func ExampleTransactionUsage() {
	var dbPool ormhelper.Executor
	dbName := "your_database_name"
	orm := ormhelper.New(dbPool, dbName)

	ctx := context.Background()

	// 方法1：使用现有事务（兼容现有代码）
	var tx *sql.Tx // 你的事务对象
	txORM := orm.WithTx(tx)

	// 在事务中执行操作
	_, err := txORM.Insert().
		Into("t_mcp_server_config").
		Values(map[string]interface{}{
			"f_id":   "tx-config-1",
			"f_name": "事务配置1",
		}).
		Execute(ctx)
	if err != nil {
		// 处理错误，可能需要回滚事务
		return
	}

	// 继续在同一事务中执行其他操作
	_, err = txORM.Update("t_mcp_server_config").
		Set("f_status", "active").
		WhereEq("f_id", "tx-config-1").
		Execute(ctx)
	if err != nil {
		// 处理错误，可能需要回滚事务
		return
	}
}

// ExampleComplexQuery 展示复杂查询的构建
func ExampleComplexQuery() {
	var dbPool ormhelper.Executor
	dbName := "your_database_name"
	orm := ormhelper.New(dbPool, dbName)

	ctx := context.Background()

	// 复杂WHERE条件
	var configs []MCPConfigExample
	err := orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		And(func(w *ormhelper.WhereBuilder) {
			w.Gt("f_create_time", time.Now().UnixMilli()).
				Lt("f_create_time", time.Now().Add(time.Minute).UnixMilli())
		}).
		Or(func(w *ormhelper.WhereBuilder) {
			w.Eq("f_category", "special").
				Like("f_name", "%test%")
		}).
		OrderByDesc("f_update_time").
		Limit(20).  //nolint:mnd
		Offset(10). //nolint:mnd
		Get(ctx, &configs)
	if err != nil {
		return
	}

	// JOIN查询示例
	query, args := orm.Select("c.f_id", "c.f_name", "h.f_version").
		From("t_mcp_server_config c").
		LeftJoin("t_mcp_server_release_history h", "c.f_id = h.f_mcp_id").
		WhereEq("c.f_status", "active").
		OrderByDesc("h.f_create_time").
		Build()

	// 执行原生SQL查询
	rows, err := orm.GetExecutor().QueryContext(ctx, query, args...)
	if err != nil {
		return
	}
	if rows.Err() != nil {
		_ = rows.Close()
		return
	}
	_ = rows.Close()
	// 处理结果...
}

// MCPConfigExample 示例配置结构体
type MCPConfigExample struct {
	ID          string `json:"f_id" db:"f_id"`
	Name        string `json:"f_name" db:"f_name"`
	Description string `json:"f_description" db:"f_description"`
	Status      string `json:"f_status" db:"f_status"`
	Category    string `json:"f_category" db:"f_category"`
	CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
	UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
}

// ConfigDAO 示例DAO实现，展示如何在实际项目中组织代码
type ConfigDAO struct {
	orm *ormhelper.DB
}

// NewConfigDAO 创建DAO实例
func NewConfigDAO(orm *ormhelper.DB) *ConfigDAO {
	return &ConfigDAO{orm: orm}
}

// GetActiveConfigs 获取活跃配置列表
func (dao *ConfigDAO) GetActiveConfigs(ctx context.Context) ([]*MCPConfigExample, error) {
	var configs []*MCPConfigExample
	err := dao.orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_status", "active").
		OrderByDesc("f_update_time").
		Get(ctx, &configs)
	return configs, err
}

// GetConfigByID 根据ID获取配置
func (dao *ConfigDAO) GetConfigByID(ctx context.Context, id string) (*MCPConfigExample, error) {
	config := &MCPConfigExample{}
	err := dao.orm.Select().
		From("t_mcp_server_config").
		WhereEq("f_id", id).
		First(ctx, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// CreateConfig 创建新配置
func (dao *ConfigDAO) CreateConfig(ctx context.Context, config *MCPConfigExample) error {
	now := time.Now().UnixNano()
	data := map[string]interface{}{
		"f_id":          config.ID,
		"f_name":        config.Name,
		"f_description": config.Description,
		"f_status":      config.Status,
		"f_category":    config.Category,
		"f_create_time": now,
		"f_update_time": now,
	}

	_, err := dao.orm.Insert().
		Into("t_mcp_server_config").
		Values(data).
		Execute(ctx)
	return err
}

// UpdateConfigStatus 更新配置状态
func (dao *ConfigDAO) UpdateConfigStatus(ctx context.Context, id, status string) error {
	_, err := dao.orm.Update("t_mcp_server_config").
		Set("f_status", status).
		Set("f_update_time", time.Now().UnixNano()).
		WhereEq("f_id", id).
		Execute(ctx)
	return err
}

// DeleteConfig 删除配置
func (dao *ConfigDAO) DeleteConfig(ctx context.Context, id string) error {
	_, err := dao.orm.Delete().
		From("t_mcp_server_config").
		WhereEq("f_id", id).
		Execute(ctx)
	return err
}

// GetConfigsPage 分页查询配置
func (dao *ConfigDAO) GetConfigsPage(ctx context.Context, page, pageSize int, status string) ([]*MCPConfigExample, int64, error) {
	// 查询总数
	countBuilder := dao.orm.Select().From("t_mcp_server_config")
	if status != "" {
		countBuilder.WhereEq("f_status", status)
	}
	total, err := countBuilder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var configs []*MCPConfigExample
	queryBuilder := dao.orm.Select().
		From("t_mcp_server_config").
		OrderByDesc("f_update_time").
		Limit(pageSize).
		Offset((page - 1) * pageSize)

	if status != "" {
		queryBuilder.WhereEq("f_status", status)
	}

	err = queryBuilder.Get(ctx, &configs)
	return configs, total, err
}

// BatchCreateConfigs 批量创建配置
func (dao *ConfigDAO) BatchCreateConfigs(ctx context.Context, configs []*MCPConfigExample) error {
	if len(configs) == 0 {
		return nil
	}

	columns := []string{"f_id", "f_name", "f_description", "f_status", "f_category", "f_create_time", "f_update_time"}
	values := make([][]interface{}, len(configs))

	now := time.Now().UnixNano()
	for i, config := range configs {
		values[i] = []interface{}{
			config.ID,
			config.Name,
			config.Description,
			config.Status,
			config.Category,
			now,
			now,
		}
	}

	_, err := dao.orm.Insert().
		Into("t_mcp_server_config").
		BatchValues(columns, values).
		Execute(ctx)
	return err
}
