package dbaccess

import (
	"context"
	"database/sql"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

type operatorReleaseHistoryDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	rhOnce sync.Once
	rh     model.IOperatorReleaseHistoryDB
)

const tbOperatorReleaseHistory = "t_operator_release_history"

func NewOperatorReleaseHistoryDB() model.IOperatorReleaseHistoryDB {
	rhOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		rh = &operatorReleaseHistoryDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    ormhelper.New(dbPool, dbName),
		}
	})
	return rh
}

// Insert 添加
func (rh *operatorReleaseHistoryDB) Insert(ctx context.Context, tx *sql.Tx, historyDB *model.OperatorReleaseHistoryDB) (err error) {
	orm := rh.orm
	if tx != nil {
		orm = rh.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbOperatorReleaseHistory).Values(map[string]interface{}{
		"f_op_id":            historyDB.OpID,
		"f_metadata_version": historyDB.MetadataVersion,
		"f_metadata_type":    historyDB.MetadataType,
		"f_op_release":       historyDB.OpRelease,
		"f_tag":              historyDB.Tag,
		"f_create_time":      historyDB.CreateTime,
		"f_create_user":      historyDB.CreateUser,
		"f_update_time":      historyDB.UpdateTime,
		"f_update_user":      historyDB.UpdateUser,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert operator release history error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = errors.Wrapf(err, "insert operator release history check addected error")
		return
	}
	if !ok {
		err = errors.Wrapf(err, "insert operator release history check addected error, affected rows is 0")
	}
	return
}

// DeleteByOpID 根据ID删除记录
func (rh *operatorReleaseHistoryDB) DeleteByOpID(ctx context.Context, tx *sql.Tx, opID string) (err error) {
	orm := rh.orm
	if tx != nil {
		orm = rh.orm.WithTx(tx)
	}
	_, err = orm.Delete().From(tbOperatorReleaseHistory).WhereEq("f_op_id", opID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete operator release history info error")
	}
	return
}

// SelectByOpID 根据算子ID查询发布历史
func (rh *operatorReleaseHistoryDB) SelectByOpID(ctx context.Context, opID string) (histories []*model.OperatorReleaseHistoryDB, err error) {
	orm := rh.orm
	histories = []*model.OperatorReleaseHistoryDB{}
	err = orm.Select().From(tbOperatorReleaseHistory).WhereEq("f_op_id", opID).OrderByDesc("f_create_time").
		Get(ctx, &histories)
	return
}

// BatchDeleteByID 批量删除
func (rh *operatorReleaseHistoryDB) BatchDeleteByID(ctx context.Context, tx *sql.Tx, ids []int64) (err error) {
	orm := rh.orm
	if tx != nil {
		orm = rh.orm.WithTx(tx)
	}
	var values []interface{}
	for _, id := range ids {
		values = append(values, id)
	}
	_, err = orm.Delete().From(tbOperatorReleaseHistory).WhereIn("f_id", values...).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "batch delete operator release history info error")
	}
	return
}

// SelectByOpIDAndMetdata 根据算子ID和元数据版本查询发布历史
func (rh *operatorReleaseHistoryDB) SelectByOpIDAndMetdata(ctx context.Context, opID, metadataVersion string) (exist bool,
	historyDB *model.OperatorReleaseHistoryDB, err error) {
	orm := rh.orm
	historyDB = &model.OperatorReleaseHistoryDB{}
	err = orm.Select().From(tbOperatorReleaseHistory).WhereEq("f_op_id", opID).
		WhereEq("f_metadata_version", metadataVersion).OrderByDesc("f_update_time").Limit(1).First(ctx, historyDB)
	exist, err = checkHasQueryErr(err)
	return
}

// SelectByOpIDAndTag 根据算子ID和版本查询发布历史
func (rh *operatorReleaseHistoryDB) SelectByOpIDAndTag(ctx context.Context, opID string,
	tag int) (exist bool, historyDB *model.OperatorReleaseHistoryDB, err error) {
	orm := rh.orm
	historyDB = &model.OperatorReleaseHistoryDB{}
	err = orm.Select().From(tbOperatorReleaseHistory).WhereEq("f_op_id", opID).
		WhereEq("f_tag", tag).OrderByDesc("f_update_time").Limit(1).First(ctx, historyDB)
	exist, err = checkHasQueryErr(err)
	return
}

// UpdateReleaseHistoryByID  更新发布历史
func (rh *operatorReleaseHistoryDB) UpdateReleaseHistoryByID(ctx context.Context, tx *sql.Tx, historyDB *model.OperatorReleaseHistoryDB) (err error) {
	orm := rh.orm
	if tx != nil {
		orm = rh.orm.WithTx(tx)
	}
	_, err = orm.Update(tbOperatorReleaseHistory).SetData(map[string]interface{}{
		"f_op_id":            historyDB.OpID,
		"f_metadata_version": historyDB.MetadataVersion,
		"f_metadata_type":    historyDB.MetadataType,
		"f_op_release":       historyDB.OpRelease,
		"f_tag":              historyDB.Tag,
		"f_create_time":      historyDB.CreateTime,
		"f_create_user":      historyDB.CreateUser,
		"f_update_time":      historyDB.UpdateTime,
		"f_update_user":      historyDB.UpdateUser,
	}).WhereEq("f_id", historyDB.ID).Execute(ctx)
	return
}
