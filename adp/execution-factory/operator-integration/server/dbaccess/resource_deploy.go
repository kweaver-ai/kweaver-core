package dbaccess

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type resourceDeployDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	resourceDeployOnce     sync.Once
	resourceDeployInstance *resourceDeployDB
)

const (
	tbResourceDeploy = "t_resource_deploy"
)

// NewResourceDeployDBSingleton 创建资源部署数据库单例
func NewResourceDeployDBSingleton() model.DBResourceDeploy {
	conf := config.NewConfigLoader()
	dbPool := db.NewDBPool()
	dbName := conf.GetDBName()
	logger := conf.GetLogger()
	resourceDeployOnce.Do(func() {
		resourceDeployInstance = &resourceDeployDB{
			dbPool: dbPool,
			logger: logger,
			dbName: dbName,
			orm:    ormhelper.New(dbPool, dbName),
		}
	})
	return resourceDeployInstance
}

func (r *resourceDeployDB) Insert(ctx context.Context, tx *sql.Tx, resourceDeploy *model.ResourceDeployDB) (ID string, err error) {
	orm := r.orm
	if tx != nil {
		orm = r.orm.WithTx(tx)
	}

	now := time.Now().UnixNano()
	resourceDeploy.CreateTime = now
	resourceDeploy.UpdateTime = now

	var id int64
	id, err = orm.Insert().Into(tbResourceDeploy).Values(map[string]interface{}{
		"f_resource_id": resourceDeploy.ResourceID,
		"f_type":        resourceDeploy.Type,
		"f_version":     resourceDeploy.Version,
		"f_name":        resourceDeploy.Name,
		"f_description": resourceDeploy.Description,
		"f_config":      resourceDeploy.Config,
		"f_status":      resourceDeploy.Status,
		"f_create_user": resourceDeploy.CreateUser,
		"f_create_time": now,
		"f_update_user": resourceDeploy.UpdateUser,
		"f_update_time": now,
	}).ExecuteAndReturnID(ctx)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (r *resourceDeployDB) Update(ctx context.Context, tx *sql.Tx, resourceDeploy *model.ResourceDeployDB) error {
	orm := r.orm
	if tx != nil {
		orm = r.orm.WithTx(tx)
	}

	now := time.Now().UnixNano()
	resourceDeploy.UpdateTime = now

	_, err := orm.Update(tbResourceDeploy).SetData(map[string]interface{}{
		"f_name":        resourceDeploy.Name,
		"f_description": resourceDeploy.Description,
		"f_config":      resourceDeploy.Config,
		"f_status":      resourceDeploy.Status,
		"f_update_user": resourceDeploy.UpdateUser,
		"f_update_time": now,
	}).WhereEq("f_resource_id", resourceDeploy.ResourceID).
		WhereEq("f_type", resourceDeploy.Type).
		WhereEq("f_version", resourceDeploy.Version).
		Execute(ctx)
	return err
}

func (r *resourceDeployDB) Delete(ctx context.Context, tx *sql.Tx, resourceVersion int, resourceType, resourceID string) error {
	orm := r.orm
	if tx != nil {
		orm = r.orm.WithTx(tx)
	}

	_, err := orm.Delete().
		From(tbResourceDeploy).
		WhereEq("f_resource_id", resourceID).
		WhereEq("f_type", resourceType).
		WhereEq("f_version", resourceVersion).
		Execute(ctx)
	return err
}

func (r *resourceDeployDB) SelectList(ctx context.Context, tx *sql.Tx, resourceDeploy *model.ResourceDeployDB) (list []*model.ResourceDeployDB, err error) {
	orm := r.orm
	if tx != nil {
		orm = r.orm.WithTx(tx)
	}

	query := orm.Select().From(tbResourceDeploy)
	if resourceDeploy != nil {
		if resourceDeploy.Type != "" {
			query = query.WhereEq("f_type", resourceDeploy.Type)
		}
		if resourceDeploy.ResourceID != "" {
			query = query.WhereEq("f_resource_id", resourceDeploy.ResourceID)
		}
		if resourceDeploy.Version != 0 {
			query = query.WhereEq("f_version", resourceDeploy.Version)
		}
	}
	list = []*model.ResourceDeployDB{}
	err = query.Get(ctx, &list)
	return list, err
}

func (r *resourceDeployDB) SelectListByResourceID(ctx context.Context, resourceID string) (list []*model.ResourceDeployDB, err error) {
	query := r.orm.Select().From(tbResourceDeploy).WhereEq("f_resource_id", resourceID)
	list = []*model.ResourceDeployDB{}
	err = query.Get(ctx, &list)
	return list, err
}

func (r *resourceDeployDB) DeleteByResourceID(ctx context.Context, tx *sql.Tx, resourceID string) (err error) {
	orm := r.orm
	if tx != nil {
		orm = r.orm.WithTx(tx)
	}
	_, err = orm.Delete().From(tbResourceDeploy).WhereEq("f_resource_id", resourceID).Execute(ctx)
	return err
}

// Exists 查询资源部署是否存在
func (r *resourceDeployDB) Exists(ctx context.Context, resourceID string, version int) (exists bool, err error) {
	query := r.orm.Select().From(tbResourceDeploy).WhereEq("f_resource_id", resourceID).WhereEq("f_version", version)
	exists, err = query.Exists(ctx)
	return exists, err
}
