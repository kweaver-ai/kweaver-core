package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/google/uuid"
)

type mcpServerConfigDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	mcOnce sync.Once
	mc     model.DBMCPServerConfig
)

const (
	// tbMCPServerConfig MCP Server配置表名
	tbMCPServerConfig = "t_mcp_server_config"
)

// NewMCPServerConfigDBSingleton 创建MCP Server配置数据库访问对象单例
func NewMCPServerConfigDBSingleton() model.DBMCPServerConfig {
	confLoader := config.NewConfigLoader()
	dbPool := db.NewDBPool()
	dbName := confLoader.GetDBName()
	logger := confLoader.GetLogger()

	mcOnce.Do(func() {
		// 使用基本的ORM实例，不包含日志功能
		orm := ormhelper.New(dbPool, dbName)

		mc = &mcpServerConfigDB{
			dbPool: dbPool,
			logger: logger,
			dbName: dbName,
			orm:    orm,
		}
	})
	return mc
}

// Insert 插入MCP Server配置
func (m *mcpServerConfigDB) Insert(ctx context.Context, tx *sql.Tx, config *model.MCPServerConfigDB) (id string, err error) {
	now := time.Now().UnixNano()
	MCPID := uuid.New().String()
	if config.MCPID != "" {
		MCPID = config.MCPID
	}

	// 默认版本号为1
	if config.Version == 0 {
		config.Version = 1
	}
	config.MCPID = MCPID
	config.CreateTime = now
	config.UpdateTime = now

	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	// 使用ORM Helper插入数据
	_, err = orm.Insert().Into(tbMCPServerConfig).Values(map[string]interface{}{
		"f_mcp_id":        config.MCPID,
		"f_name":          config.Name,
		"f_description":   config.Description,
		"f_mode":          config.Mode,
		"f_url":           config.URL,
		"f_headers":       config.Headers,
		"f_command":       config.Command,
		"f_env":           config.Env,
		"f_args":          config.Args,
		"f_status":        config.Status,
		"f_category":      config.Category,
		"f_source":        config.Source,
		"f_create_user":   config.CreateUser,
		"f_create_time":   config.CreateTime,
		"f_update_user":   config.UpdateUser,
		"f_update_time":   config.UpdateTime,
		"f_is_internal":   config.IsInternal,
		"f_creation_type": config.CreationType,
		"f_version":       config.Version,
	}).Execute(ctx)

	if err != nil {
		return "", err
	}
	return config.MCPID, nil
}

// SelectByID 根据ID查询MCP Server配置
func (m *mcpServerConfigDB) SelectByID(ctx context.Context, tx *sql.Tx, mcpID string) (config *model.MCPServerConfigDB, err error) {
	config = &model.MCPServerConfigDB{}

	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	err = orm.Select().From(tbMCPServerConfig).WhereEq("f_mcp_id", mcpID).First(ctx, config)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

// UpdateByID 根据ID更新MCP Server配置
func (m *mcpServerConfigDB) UpdateByID(ctx context.Context, tx *sql.Tx, config *model.MCPServerConfigDB) error {
	config.UpdateTime = time.Now().UnixNano()

	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	_, err := orm.Update(tbMCPServerConfig).SetData(map[string]interface{}{
		"f_name":        config.Name,
		"f_description": config.Description,
		"f_mode":        config.Mode,
		"f_url":         config.URL,
		"f_headers":     config.Headers,
		"f_command":     config.Command,
		"f_env":         config.Env,
		"f_args":        config.Args,
		"f_status":      config.Status,
		"f_category":    config.Category,
		"f_source":      config.Source,
		"f_update_user": config.UpdateUser,
		"f_update_time": config.UpdateTime,
		"f_version":     config.Version,
	}).WhereEq("f_mcp_id", config.MCPID).Execute(ctx)
	return err
}

// UpdateStatus 更新MCP Server配置状态
func (m *mcpServerConfigDB) UpdateStatus(ctx context.Context, tx *sql.Tx, mcpID, status, updateUser string, version int) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	_, err := orm.Update(tbMCPServerConfig).SetData(map[string]interface{}{
		"f_status":      status,
		"f_update_user": updateUser,
		"f_update_time": time.Now().UnixNano(),
		"f_version":     version,
	}).WhereEq("f_mcp_id", mcpID).Execute(ctx)
	return err
}

// DeleteByID 根据ID删除MCP Server配置
func (m *mcpServerConfigDB) DeleteByID(ctx context.Context, tx *sql.Tx, mcpID string) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	_, err := orm.Delete().From(tbMCPServerConfig).WhereEq("f_mcp_id", mcpID).Execute(ctx)
	return err
}

// BatchDelete 批量删除MCP Server配置
func (m *mcpServerConfigDB) BatchDelete(ctx context.Context, tx *sql.Tx, ids []string) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbMCPServerConfig).WhereIn("f_id", utils.SliceToInterface(ids)...).Execute(ctx)
	return err
}

// SelectListPage 分页查询MCP Server配置列表
func (m *mcpServerConfigDB) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (configList []*model.MCPServerConfigDB, err error) {
	query := m.orm.Select().From(tbMCPServerConfig)
	query = m.applyFilterConditions(query, filter)
	if cursor != nil {
		query = query.Cursor(cursor)
	}
	// 处理排序和分页
	query = query.Sort(sort)
	if filter["all"] == nil || filter["all"] == false {
		pageSize, ok := filter["limit"].(int)
		if ok {
			query.Limit(pageSize)
		}
		offset, ok := filter["offset"].(int)
		if ok {
			query.Offset(offset)
		}
	}
	// 执行查询
	configList = []*model.MCPServerConfigDB{}
	err = query.Get(ctx, &configList)
	return configList, err
}

// SelectByName 根据名称查询MCP Server配置
func (m *mcpServerConfigDB) SelectByName(ctx context.Context, tx *sql.Tx, name string, status []string) (config *model.MCPServerConfigDB, err error) {
	config = &model.MCPServerConfigDB{}

	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	args := []interface{}{}
	for _, v := range status {
		args = append(args, v)
	}
	err = orm.Select().From(tbMCPServerConfig).WhereEq("f_name", name).WhereIn("f_status", args...).First(ctx, config)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

// CountByWhereClause 根据条件统计数量
func (m *mcpServerConfigDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error) {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	query := orm.Select().From(tbMCPServerConfig)
	query = m.applyFilterConditions(query, filter)

	count, err = query.Count(ctx)
	return count, err
}

// applyFilterConditions 应用过滤条件到查询
func (m *mcpServerConfigDB) applyFilterConditions(query *ormhelper.SelectBuilder, filter map[string]interface{}) *ormhelper.SelectBuilder {
	if filter == nil {
		return query
	}
	// 支持的条件查询
	if filter["name"] != nil {
		name := filter["name"].(string)
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if filter["status"] != nil {
		query = query.WhereEq("f_status", filter["status"])
	}
	if filter["category"] != nil {
		query = query.WhereEq("f_category", filter["category"])
	}
	if filter["source"] != nil {
		query = query.WhereEq("f_source", filter["source"])
	}
	if filter["createUser"] != nil {
		query = query.WhereEq("f_create_user", filter["createUser"])
	}
	if filter["mode"] != nil {
		query = query.WhereEq("f_mode", filter["mode"])
	}
	if filter["in"] != nil {
		queryInParams, ok := filter["in"].([]string)
		if !ok || len(queryInParams) == 0 {
			return query
		}
		arrs := []interface{}{}
		for _, v := range queryInParams {
			if v != "" {
				arrs = append(arrs, v)
			}
		}
		if len(arrs) > 0 {
			query = query.WhereIn("f_mcp_id", arrs...)
		}
	}
	return query
}

// SelectByIDs 根据ID列表查询MCP Server配置列表
func (m *mcpServerConfigDB) SelectByMCPIDs(ctx context.Context, mcpIDs []string) (configList []*model.MCPServerConfigDB, err error) {
	orm := m.orm
	configList = []*model.MCPServerConfigDB{}
	err = orm.Select().From(tbMCPServerConfig).WhereIn("f_mcp_id", utils.SliceToInterface(mcpIDs)...).Get(ctx, &configList)
	return configList, err
}

// SelectListByNamesAndStatus 根据名字及状态批量获取列表
func (m *mcpServerConfigDB) SelectListByNamesAndStatus(ctx context.Context, names []string, status ...string) (configList []*model.MCPServerConfigDB, err error) {
	configList = []*model.MCPServerConfigDB{}
	query := m.orm.Select().From(tbMCPServerConfig).WhereIn("f_name", utils.SliceToInterface(names)...)
	if len(status) > 0 {
		query = query.WhereIn("f_status", utils.SliceToInterface(status)...)
	}
	err = query.Get(ctx, &configList)
	return configList, err
}
