package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

// functionMetadataDB 函数元数据数据库
type functionMetadataDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	fmOnce                    sync.Once
	functionMetadataDBService model.IFunctionMetadataDB
	tbFunctionMetadata        = "t_metadata_function"
)

// NewFunctionMetadataDB 创建函数元数据数据库
func NewFunctionMetadataDB() model.IFunctionMetadataDB {
	fmOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		functionMetadataDBService = &functionMetadataDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return functionMetadataDBService
}

// InsertFuncMetadata 插入函数元数据
func (fm *functionMetadataDB) InsertFuncMetadata(ctx context.Context, tx *sql.Tx, metadata *model.FunctionMetadataDB) (version string, err error) {
	if metadata.Version == "" {
		metadata.Version = uuid.New().String()
	}
	version = metadata.Version

	orm := fm.orm
	if tx != nil {
		orm = fm.orm.WithTx(tx)
	}

	row, err := orm.Insert().Into(tbFunctionMetadata).Values(map[string]interface{}{
		"f_version":          metadata.Version,
		"f_summary":          metadata.Summary,
		"f_description":      metadata.Description,
		"f_path":             metadata.Path,
		"f_method":           metadata.Method,
		"f_svc_url":          metadata.ServerURL,
		"f_api_spec":         metadata.APISpec,
		"f_script_type":      metadata.ScriptType,
		"f_dependencies":     metadata.Dependencies,
		"f_dependencies_url": metadata.DependenciesURL,
		"f_code":             metadata.Code,
		"f_create_user":      metadata.CreateUser,
		"f_create_time":      metadata.CreateTime,
		"f_update_user":      metadata.UpdateUser,
		"f_update_time":      metadata.UpdateTime,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert function metadata error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert function metadata failed, err: %v", err)
	}
	return
}

// UpdateByVersion 更新函数元数据
func (fm *functionMetadataDB) UpdateByVersion(ctx context.Context, tx *sql.Tx, metadata *model.FunctionMetadataDB) (err error) {
	now := time.Now().UnixNano()
	metadata.UpdateTime = now
	orm := fm.orm
	if tx != nil {
		orm = fm.orm.WithTx(tx)
	}
	row, err := orm.Update(tbFunctionMetadata).SetData(map[string]interface{}{
		"f_summary":          metadata.Summary,
		"f_description":      metadata.Description,
		"f_path":             metadata.Path,
		"f_method":           metadata.Method,
		"f_svc_url":          metadata.ServerURL,
		"f_api_spec":         metadata.APISpec,
		"f_script_type":      metadata.ScriptType,
		"f_dependencies":     metadata.Dependencies,
		"f_dependencies_url": metadata.DependenciesURL,
		"f_code":             metadata.Code,
		"f_update_user":      metadata.UpdateUser,
		"f_update_time":      metadata.UpdateTime,
	}).WhereEq("f_version", metadata.Version).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update function metadata error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("update function metadata failed, err: %v", err)
	}
	return
}

// SelectByVersion 查询函数元数据
func (fm *functionMetadataDB) SelectByVersion(ctx context.Context, version string) (exist bool, metadata *model.FunctionMetadataDB, err error) {
	metadata = &model.FunctionMetadataDB{}
	orm := fm.orm
	err = orm.Select().From(tbFunctionMetadata).WhereEq("f_version", version).First(ctx, metadata)
	exist, err = checkHasQueryErr(err)
	return
}

// DeleteByVersion 删除函数元数据
func (fm *functionMetadataDB) DeleteByVersion(ctx context.Context, tx *sql.Tx, version string) (err error) {
	orm := fm.orm
	if tx != nil {
		orm = fm.orm.WithTx(tx)
	}
	row, err := orm.Delete().From(tbFunctionMetadata).WhereIn("f_version", version).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete function metadata error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("delete function metadata failed, err: %v", err)
	}
	return
}

// DeleteByVersions 删除函数元数据
func (fm *functionMetadataDB) DeleteByVersions(ctx context.Context, tx *sql.Tx, versions []string) (err error) {
	orm := fm.orm
	if tx != nil {
		orm = fm.orm.WithTx(tx)
	}
	args := []interface{}{}
	for _, version := range versions {
		args = append(args, version)
	}
	row, err := orm.Delete().From(tbFunctionMetadata).WhereIn("f_version", args...).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete function metadata error")
		return
	}
	_, err = checkAffected(row)
	return
}

// InsertFuncMetadatas 插入函数元数据列表
func (fm *functionMetadataDB) InsertFuncMetadatas(ctx context.Context, tx *sql.Tx, metadatas []*model.FunctionMetadataDB) (versions []string, err error) {
	orm := fm.orm
	if tx != nil {
		orm = fm.orm.WithTx(tx)
	}
	columns := []string{
		"f_version",
		"f_summary",
		"f_description",
		"f_path",
		"f_method",
		"f_svc_url",
		"f_api_spec",
		"f_script_type",
		"f_dependencies",
		"f_dependencies_url",
		"f_code",
		"f_create_user",
		"f_create_time",
		"f_update_user",
		"f_update_time",
	}
	now := time.Now().UnixNano()
	versions = []string{}
	values := [][]interface{}{}
	for _, metadata := range metadatas {
		if metadata.Version == "" {
			metadata.Version = uuid.New().String()
		}
		versions = append(versions, metadata.Version)
		values = append(values, []interface{}{
			metadata.Version,
			metadata.Summary,
			metadata.Description,
			metadata.Path,
			metadata.Method,
			metadata.ServerURL,
			metadata.APISpec,
			metadata.ScriptType,
			metadata.Dependencies,
			metadata.DependenciesURL,
			metadata.Code,
			metadata.CreateUser,
			now,
			metadata.UpdateUser,
			now,
		})
	}
	row, err := orm.Insert().Into(tbFunctionMetadata).BatchValues(columns, values).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert function metadata error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert function metadata failed, err: %v", err)
	}
	return
}

// SelectListByVersion 查询函数元数据列表
func (fm *functionMetadataDB) SelectListByVersion(ctx context.Context, versions []string) (metadataDBs []*model.FunctionMetadataDB, err error) {
	metadataDBs = []*model.FunctionMetadataDB{}
	orm := fm.orm
	args := []interface{}{}
	for _, version := range versions {
		args = append(args, version)
	}
	err = orm.Select().From(tbFunctionMetadata).WhereIn("f_version", args...).Get(ctx, &metadataDBs)
	if err != nil {
		err = errors.Wrapf(err, "select function metadata list by version error")
	}
	return
}
