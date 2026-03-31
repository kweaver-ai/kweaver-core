package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

type mcpServerReleaseDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	mcrOnce sync.Once
	mcr     model.DBMCPServerRelease
)

const (
	// tbMCPServerRelease MCP Server发布表名
	tbMCPServerRelease = "t_mcp_server_release"
)

// NewMCPServerReleaseDBSingleton 创建MCP Server发布数据库访问对象单例
func NewMCPServerReleaseDBSingleton() model.DBMCPServerRelease {
	confLoader := config.NewConfigLoader()
	dbPool := db.NewDBPool()
	dbName := confLoader.GetDBName()
	logger := confLoader.GetLogger()

	mcrOnce.Do(func() {
		orm := ormhelper.New(dbPool, dbName)
		mcr = &mcpServerReleaseDB{
			dbPool: dbPool,
			logger: logger,
			dbName: dbName,
			orm:    orm,
		}
	})
	return mcr
}

// Insert 插入MCP Server发布
func (m *mcpServerReleaseDB) Insert(ctx context.Context, tx *sql.Tx, release *model.MCPServerReleaseDB) (err error) {
	now := time.Now().UnixNano()
	release.ReleaseTime = now

	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	// 插入数据
	row, err := orm.Insert().Into(tbMCPServerRelease).Values(map[string]interface{}{
		"f_mcp_id":        release.MCPID,
		"f_creation_type": release.CreationType,
		"f_create_user":   release.CreateUser,
		"f_create_time":   release.CreateTime,
		"f_update_user":   release.UpdateUser,
		"f_update_time":   release.UpdateTime,
		"f_name":          release.Name,
		"f_description":   release.Description,
		"f_mode":          release.Mode,
		"f_url":           release.URL,
		"f_headers":       release.Headers,
		"f_command":       release.Command,
		"f_env":           release.Env,
		"f_args":          release.Args,
		"f_category":      release.Category,
		"f_source":        release.Source,
		"f_is_internal":   release.IsInternal,
		"f_version":       release.Version,
		"f_release_desc":  release.ReleaseDesc,
		"f_release_time":  release.ReleaseTime,
		"f_release_user":  release.ReleaseUser,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert mcp failed")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = fmt.Errorf("insert mcp failed, err: %v", err)
		return
	}
	if !ok {
		err = errors.New("insert mcp failed, no affected rows")
	}
	return
}

// UpdateByMCPID 根据mcp_id更新MCP Server发布
func (m *mcpServerReleaseDB) UpdateByMCPID(ctx context.Context, tx *sql.Tx, release *model.MCPServerReleaseDB) (err error) {
	now := time.Now().UnixNano()
	release.UpdateTime = now
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	_, err = orm.Update(tbMCPServerRelease).SetData(map[string]interface{}{
		"f_update_user":  release.UpdateUser,
		"f_update_time":  release.UpdateTime,
		"f_name":         release.Name,
		"f_description":  release.Description,
		"f_mode":         release.Mode,
		"f_url":          release.URL,
		"f_headers":      release.Headers,
		"f_command":      release.Command,
		"f_env":          release.Env,
		"f_args":         release.Args,
		"f_category":     release.Category,
		"f_source":       release.Source,
		"f_version":      release.Version,
		"f_release_desc": release.ReleaseDesc,
		"f_release_time": release.ReleaseTime,
		"f_release_user": release.ReleaseUser,
	}).WhereEq("f_mcp_id", release.MCPID).Execute(ctx)
	return
}

// DeleteByMCPID 根据mcp_id删除MCP Server发布
func (m *mcpServerReleaseDB) DeleteByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) (err error) {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	_, err = orm.Delete().From(tbMCPServerRelease).WhereEq("f_mcp_id", mcpID).Execute(ctx)
	return
}

// SelectListPage 分页查询mcp server发布列表
func (m *mcpServerReleaseDB) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams,
	cursor *ormhelper.CursorParams) (releaseList []*model.MCPServerReleaseDB, err error) {
	query := m.orm.Select().From(tbMCPServerRelease)

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
	releaseList = []*model.MCPServerReleaseDB{}
	err = query.Get(ctx, &releaseList)
	return releaseList, err
}

// SelectByMCPIDs 根据mcp_id列表查询MCP Server发布
func (m *mcpServerReleaseDB) SelectByMCPIDs(ctx context.Context, tx *sql.Tx, mcpIDs, fields []string) (releaseList []*model.MCPServerReleaseDB, err error) {
	query := m.orm.Select(fields...).From(tbMCPServerRelease).WhereIn("f_mcp_id", utils.SliceToInterface(mcpIDs)...)
	err = query.Get(ctx, &releaseList)
	return releaseList, err
}

// SelectByMCPID 根据mcp_id查询MCP Server发布
func (m *mcpServerReleaseDB) SelectByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) (release *model.MCPServerReleaseDB, err error) {
	release = &model.MCPServerReleaseDB{}
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	err = orm.Select().From(tbMCPServerRelease).WhereEq("f_mcp_id", mcpID).First(ctx, release)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return release, nil
}

// CountByWhereClause 根据条件统计数量
func (m *mcpServerReleaseDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error) {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	query := orm.Select().From(tbMCPServerRelease)

	query = m.applyFilterConditions(query, filter)
	count, err = query.Count(ctx)
	return count, err
}

func (m *mcpServerReleaseDB) applyFilterConditions(query *ormhelper.SelectBuilder, filter map[string]interface{}) *ormhelper.SelectBuilder {
	if filter == nil {
		return query
	}
	// 支持的条件查询
	if filter["name"] != nil {
		name := filter["name"].(string)
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if filter["category"] != nil {
		query = query.WhereEq("f_category", filter["category"])
	}
	if filter["source"] != nil {
		query = query.WhereEq("f_source", filter["source"])
	}
	if filter["create_user"] != nil {
		query = query.WhereEq("f_create_user", filter["create_user"])
	}
	if filter["release_user"] != nil {
		query = query.WhereEq("f_release_user", filter["release_user"])
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
