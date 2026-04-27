package dbaccess

import (
	"context"
	"database/sql"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type mcpServerReleaseHistoryDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	mcrhOnce sync.Once
	mcrh     model.DBMCPServerReleaseHistory
)

const (
	tbMCPServerReleaseHistory = "t_mcp_server_release_history"
)

func NewMCPServerReleaseHistoryDBSingleton() model.DBMCPServerReleaseHistory {
	confLoader := config.NewConfigLoader()
	dbPool := db.NewDBPool()
	dbName := confLoader.GetDBName()
	logger := confLoader.GetLogger()

	mcrhOnce.Do(func() {
		orm := ormhelper.New(dbPool, dbName)
		mcrh = &mcpServerReleaseHistoryDB{
			dbPool: dbPool,
			logger: logger,
			dbName: dbName,
			orm:    orm,
		}
	})
	return mcrh
}

// Insert 插入MCP Server发布历史
func (m *mcpServerReleaseHistoryDB) Insert(ctx context.Context, tx *sql.Tx, history *model.MCPServerReleaseHistoryDB) (mcpID string, err error) {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}
	_, err = orm.Insert().Into(tbMCPServerReleaseHistory).Values(map[string]interface{}{
		"f_create_user":  history.CreateUser,
		"f_create_time":  history.CreateTime,
		"f_update_user":  history.UpdateUser,
		"f_update_time":  history.UpdateTime,
		"f_mcp_id":       history.MCPID,
		"f_mcp_release":  history.MCPRelease,
		"f_version":      history.Version,
		"f_release_desc": history.ReleaseDesc,
	}).Execute(ctx)
	if err != nil {
		return
	}
	mcpID = history.MCPID
	return
}

func (m *mcpServerReleaseHistoryDB) SelectByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) (historys []*model.MCPServerReleaseHistoryDB, err error) {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	historys = []*model.MCPServerReleaseHistoryDB{}
	err = orm.Select().From(tbMCPServerReleaseHistory).WhereEq("f_mcp_id", mcpID).OrderByDesc("f_create_time").Get(ctx, &historys)

	return historys, err
}

func (m *mcpServerReleaseHistoryDB) DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	_, err := orm.Delete().From(tbMCPServerReleaseHistory).WhereEq("f_id", id).Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *mcpServerReleaseHistoryDB) DeleteByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	_, err := orm.Delete().From(tbMCPServerReleaseHistory).WhereEq("f_mcp_id", mcpID).Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *mcpServerReleaseHistoryDB) DeleteByMCPIDAndVersion(ctx context.Context, tx *sql.Tx, mcpID string, version int) error {
	orm := m.orm
	if tx != nil {
		orm = m.orm.WithTx(tx)
	}

	_, err := orm.Delete().From(tbMCPServerReleaseHistory).WhereEq("f_mcp_id", mcpID).WhereEq("f_version", version).Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}
