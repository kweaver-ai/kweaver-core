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
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

type operatorReleaseDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	orOnce sync.Once
	or     model.IOperatorReleaseDB
)

const (
	tbOperatorRelease = "t_operator_release"
)

// NewOperatorReleaseDB 创建算子发布数据库
func NewOperatorReleaseDB() model.IOperatorReleaseDB {
	orOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		or = &operatorReleaseDB{
			dbPool: dbPool,
			logger: confLoader.GetLogger(),
			dbName: dbName,
			orm:    orm,
		}
	})
	return or
}

// BatchInsert 批量插入算子发布信息
func (or *operatorReleaseDB) BatchInsert(ctx context.Context, tx *sql.Tx, releaseList []*model.OperatorReleaseDB) (opIDs []string, err error) {
	orm := or.orm
	if tx != nil {
		orm = or.orm.WithTx(tx)
	}
	columns := []string{
		"f_op_id",
		"f_name",
		"f_metadata_version",
		"f_metadata_type",
		"f_status",
		"f_operator_type",
		"f_execution_mode",
		"f_execute_control",
		"f_category",
		"f_source",
		"f_extend_info",
		"f_create_time",
		"f_create_user",
		"f_release_time",
		"f_release_user",
		"f_update_time",
		"f_update_user",
		"f_tag",
		"f_is_internal",
		"f_is_data_source",
	}
	now := time.Now().UnixNano()
	values := [][]interface{}{}
	for _, releaseDB := range releaseList {
		values = append(values, []interface{}{
			releaseDB.OpID,
			releaseDB.Name,
			releaseDB.MetadataVersion,
			releaseDB.MetadataType,
			releaseDB.Status,
			releaseDB.OperatorType,
			releaseDB.ExecutionMode,
			releaseDB.ExecuteControl,
			releaseDB.Category,
			releaseDB.Source,
			releaseDB.ExtendInfo,
			releaseDB.CreateTime,
			releaseDB.CreateUser,
			now,
			releaseDB.ReleaseUser,
			releaseDB.UpdateTime,
			releaseDB.UpdateUser,
			releaseDB.Tag,
			releaseDB.IsInternal,
			releaseDB.IsDataSource,
		})
	}
	row, err := orm.Insert().Into(tbOperatorRelease).BatchValues(columns, values).Execute(ctx)
	if err != nil {
		err = errors.Wrap(err, "insert operator release error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = errors.Wrap(err, "insert operator release check affected error")
		return
	}
	if !ok {
		err = errors.New("insert operator release error, no affected rows")
	}
	return
}

// Insert 插入算子发布信息
func (or *operatorReleaseDB) Insert(ctx context.Context, tx *sql.Tx, releaseDB *model.OperatorReleaseDB) (err error) {
	orm := or.orm
	if tx != nil {
		orm = or.orm.WithTx(tx)
	}
	releaseDB.ReleaseTime = time.Now().UnixNano()
	row, err := orm.Insert().Into(tbOperatorRelease).Values(map[string]interface{}{
		"f_op_id":            releaseDB.OpID,
		"f_name":             releaseDB.Name,
		"f_metadata_version": releaseDB.MetadataVersion,
		"f_metadata_type":    releaseDB.MetadataType,
		"f_status":           releaseDB.Status,
		"f_operator_type":    releaseDB.OperatorType,
		"f_execution_mode":   releaseDB.ExecutionMode,
		"f_execute_control":  releaseDB.ExecuteControl,
		"f_category":         releaseDB.Category,
		"f_source":           releaseDB.Source,
		"f_extend_info":      releaseDB.ExtendInfo,
		"f_create_time":      releaseDB.CreateTime,
		"f_create_user":      releaseDB.CreateUser,
		"f_release_time":     releaseDB.ReleaseTime,
		"f_release_user":     releaseDB.ReleaseUser,
		"f_update_time":      releaseDB.UpdateTime,
		"f_update_user":      releaseDB.UpdateUser,
		"f_tag":              releaseDB.Tag,
		"f_is_internal":      releaseDB.IsInternal,
		"f_is_data_source":   releaseDB.IsDataSource,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert operator release info error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = errors.Wrapf(err, "insert operator release info check affected error")
		return
	}
	if !ok {
		err = errors.New("insert operator release info error, no affected rows")
	}
	return
}

// UpdateByOpID 更新算子发布信息
func (or *operatorReleaseDB) UpdateByOpID(ctx context.Context, tx *sql.Tx, releaseDB *model.OperatorReleaseDB) (err error) {
	orm := or.orm
	if tx != nil {
		orm = or.orm.WithTx(tx)
	}
	releaseDB.ReleaseTime = time.Now().UnixNano()
	_, err = orm.Update(tbOperatorRelease).SetData(map[string]interface{}{
		"f_name":             releaseDB.Name,
		"f_metadata_version": releaseDB.MetadataVersion,
		"f_metadata_type":    releaseDB.MetadataType,
		"f_status":           releaseDB.Status,
		"f_operator_type":    releaseDB.OperatorType,
		"f_execution_mode":   releaseDB.ExecutionMode,
		"f_execute_control":  releaseDB.ExecuteControl,
		"f_category":         releaseDB.Category,
		"f_source":           releaseDB.Source,
		"f_extend_info":      releaseDB.ExtendInfo,
		"f_create_time":      releaseDB.CreateTime,
		"f_create_user":      releaseDB.CreateUser,
		"f_release_time":     releaseDB.ReleaseTime,
		"f_release_user":     releaseDB.ReleaseUser,
		"f_update_time":      releaseDB.UpdateTime,
		"f_update_user":      releaseDB.UpdateUser,
		"f_tag":              releaseDB.Tag,
		"f_is_internal":      releaseDB.IsInternal,
		"f_is_data_source":   releaseDB.IsDataSource,
	}).WhereEq("f_op_id", releaseDB.OpID).Execute(ctx)
	return err
}

// DeleteByOpID 删除算子发布信息
func (or *operatorReleaseDB) DeleteByOpID(ctx context.Context, tx *sql.Tx, opID string) (err error) {
	orm := or.orm
	if tx != nil {
		orm = or.orm.WithTx(tx)
	}
	_, err = orm.Delete().From(tbOperatorRelease).WhereEq("f_op_id", opID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete operator release info error")
	}
	return
}

// SelectByOpID 根据算子ID查询算子发布信息
func (or *operatorReleaseDB) SelectByOpID(ctx context.Context, opID string) (exist bool, releaseDB *model.OperatorReleaseDB, err error) {
	releaseDB = &model.OperatorReleaseDB{}
	orm := or.orm
	err = orm.Select().From(tbOperatorRelease).WhereEq("f_op_id", opID).First(ctx, releaseDB)
	exist, err = checkHasQueryErr(err)
	return
}

// SelectByName 根据算子名称查询算子发布信息
func (or *operatorReleaseDB) SelectByName(ctx context.Context, tx *sql.Tx, name string) (exist bool, releaseDB *model.OperatorReleaseDB, err error) {
	releaseDB = &model.OperatorReleaseDB{}
	orm := or.orm
	if tx != nil {
		orm = or.orm.WithTx(tx)
	}
	err = orm.Select().From(tbOperatorRelease).WhereEq("f_name", name).First(ctx, releaseDB)
	exist, err = checkHasQueryErr(err)
	return
}

// CountByWhereClause 根据条件查询算子发布信息数量
func (or *operatorReleaseDB) CountByWhereClause(ctx context.Context, conditions map[string]interface{}) (count int64, err error) {
	orm := or.orm
	queryBuilder := orm.Select().From(tbOperatorRelease)
	queryBuilder = or.buildQueryConditions(queryBuilder, conditions)
	count, err = queryBuilder.Count(ctx)
	return
}

// SelectByWhereClause 根据条件查询算子发布信息
func (or *operatorReleaseDB) SelectByWhereClause(ctx context.Context, conditions map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (
	releaseList []*model.OperatorReleaseDB, err error) {
	orm := or.orm
	queryBuilder := orm.Select().From(tbOperatorRelease)
	queryBuilder = or.buildQueryConditions(queryBuilder, conditions)
	queryBuilder.Cursor(cursor)
	queryBuilder.Sort(sort)
	// 处理分页
	if conditions["all"] == nil || conditions["all"] == false {
		pageSize, ok := conditions["limit"].(int)
		if ok {
			queryBuilder.Limit(pageSize)
		}
		offset, ok := conditions["offset"].(int)
		if ok {
			queryBuilder.Offset(offset)
		}
	}
	releaseList = []*model.OperatorReleaseDB{}
	err = queryBuilder.Get(ctx, &releaseList)
	if err != nil {
		err = errors.Wrapf(err, "select operator release info error")
	}
	return
}

func (or *operatorReleaseDB) buildQueryConditions(query *ormhelper.SelectBuilder, conditions map[string]interface{}) *ormhelper.SelectBuilder {
	if len(conditions) == 0 {
		return query
	}
	if conditions["create_user"] != nil {
		query = query.WhereEq("f_create_user", conditions["create_user"])
	}
	if conditions["release_user"] != nil {
		query = query.WhereEq("f_release_user", conditions["release_user"])
	}
	if conditions["name"] != nil {
		name := conditions["name"].(string)
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if conditions["category"] != nil {
		query = query.WhereEq("f_category", conditions["category"])
	}
	if conditions["operator_type"] != nil {
		query = query.WhereEq("f_operator_type", conditions["operator_type"])
	}
	if conditions["status"] != nil {
		query = query.WhereEq("f_status", conditions["status"])
	}
	if conditions["is_data_source"] != nil {
		query = query.WhereEq("f_is_data_source", conditions["is_data_source"])
	}
	if conditions["execution_mode"] != nil {
		query = query.WhereEq("f_execution_mode", conditions["execution_mode"])
	}
	if conditions["metadata_type"] != nil {
		query = query.WhereEq("f_metadata_type", conditions["metadata_type"])
	}
	if conditions["in"] != nil {
		operatorIDs := conditions["in"].([]string)
		if len(operatorIDs) == 0 {
			return query
		}
		var arr []interface{}
		for _, id := range operatorIDs {
			if id != "" {
				arr = append(arr, id)
			}
		}
		if len(arr) > 0 {
			query = query.WhereIn("f_op_id", arr...)
		}
	}
	return query
}
