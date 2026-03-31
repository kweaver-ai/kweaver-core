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
	"github.com/pkg/errors"
)

type internalComponentConfigDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	icOnce sync.Once
	ic     model.IInternalComponentConfigDB
)

const (
	// tbInternalComponentConfig 表名
	tbInternalComponentConfig = "t_internal_component_config"
)

// NewInternalComponentConfigDBSingleton 创建组件配置数据库访问对象单例
func NewInternalComponentConfigDBSingleton() model.IInternalComponentConfigDB {
	icOnce.Do(func() {
		dbPool := db.NewDBPool()
		confLoader := config.NewConfigLoader()
		dbName := confLoader.GetDBName()
		ic = &internalComponentConfigDB{
			dbName: dbName,
			dbPool: dbPool,
			logger: confLoader.GetLogger(),
			orm:    ormhelper.New(dbPool, dbName),
		}
	})
	return ic
}

// InsertConfig 添加配置
func (ic *internalComponentConfigDB) InsertConfig(ctx context.Context, tx *sql.Tx, config *model.InternalComponentConfigDB) (err error) {
	orm := ic.orm
	if tx != nil {
		orm = ic.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbInternalComponentConfig).Values(map[string]interface{}{
		"f_component_type": config.ComponentType,
		"f_component_id":   config.ComponentID,
		"f_config_version": config.ConfigVersion,
		"f_config_source":  config.ConfigSource,
		"f_protected_flag": config.ProtectedFlag,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert internal component config error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("insert internal component config error")
	}
	return
}

// UpdateConfig 更新配置
func (ic *internalComponentConfigDB) UpdateConfig(ctx context.Context, tx *sql.Tx, config *model.InternalComponentConfigDB) (err error) {
	orm := ic.orm
	if tx != nil {
		orm = ic.orm.WithTx(tx)
	}

	row, err := orm.Update(tbInternalComponentConfig).SetData(map[string]interface{}{
		"f_config_version": config.ConfigVersion,
		"f_config_source":  config.ConfigSource,
		"f_protected_flag": config.ProtectedFlag,
	}).WhereEq("f_component_type", config.ComponentType).
		WhereEq("f_component_id", config.ComponentID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update internal component config error")
		return
	}
	_, err = checkAffected(row)
	return
}

// DeleteConfig 删除配置
func (ic *internalComponentConfigDB) DeleteConfig(ctx context.Context, tx *sql.Tx, configType, configID string) (err error) {
	orm := ic.orm
	if tx != nil {
		orm = ic.orm.WithTx(tx)
	}
	row, err := orm.Delete().From(tbInternalComponentConfig).
		WhereEq("f_component_type", configType).
		WhereEq("f_component_id", configID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete internal component config failed")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("delete internal component config failed")
	}
	return
}

// SelectConfig 查询配置
func (ic *internalComponentConfigDB) SelectConfig(ctx context.Context, configType, configID string) (exist bool, config *model.InternalComponentConfigDB, err error) {
	config = &model.InternalComponentConfigDB{}
	orm := ic.orm
	err = orm.Select().From(tbInternalComponentConfig).
		WhereEq("f_component_type", configType).
		WhereEq("f_component_id", configID).First(ctx, config)
	exist, err = checkHasQuery("SelectConfig", err)
	return
}
