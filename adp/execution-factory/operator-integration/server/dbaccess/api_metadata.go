// Package dbaccess
// @file api_metadata.go
// @description: 实现API元数据数据库访问接口
package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type apiMetadataDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	amOnce sync.Once
	am     model.IAPIMetadataDB
)

// NewAPIMetadataDB 初始化API元数据DB
func NewAPIMetadataDB() model.IAPIMetadataDB {
	amOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		am = &apiMetadataDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return am
}

const tableAPIMetadata = "t_metadata_api"

var metadataFields = "`f_summary`, `f_version`, `f_description`, " +
	"`f_path`, `f_svc_url`, `f_method`, `f_api_spec`, `f_create_user`, `f_create_time`,`f_update_user`,`f_update_time`"

// InsertAPIMetadata 插入API元数据
func (a *apiMetadataDB) InsertAPIMetadata(ctx context.Context, tx *sql.Tx, metadata *model.APIMetadataDB) (version string, err error) {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	if metadata.Version == "" {
		metadata.Version = uuid.New().String()
	}
	version = metadata.Version
	now := time.Now().UnixNano()
	metadata.CreateTime = now
	metadata.UpdateTime = now
	query := "INSERT INTO `%s`.`t_metadata_api`(%s)VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	query = fmt.Sprintf(query, a.dbName, metadataFields)
	row, err := exec(ctx, query, metadata.Summary, metadata.Version, metadata.Description, metadata.Path, metadata.ServerURL, metadata.Method,
		metadata.APISpec, metadata.CreateUser, metadata.CreateTime, metadata.UpdateUser, metadata.UpdateTime)
	if err != nil {
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert operator failed, err: %v", err)
	}
	return
}

// InsertAPIMetadatas 批量插入API元数据
func (a *apiMetadataDB) InsertAPIMetadatas(ctx context.Context, tx *sql.Tx, metadatas []*model.APIMetadataDB) (versions []string, err error) {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	query := "INSERT INTO `%s`.`t_metadata_api`(%s)VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	query = fmt.Sprintf(query, a.dbName, metadataFields)
	l := len(metadatas)
	if l > 1 {
		query += strings.Repeat(", (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", l-1)
	}
	now := time.Now().UnixNano()
	versions = []string{}
	args := []interface{}{}
	for _, metadata := range metadatas {
		if metadata.Version == "" {
			metadata.Version = uuid.New().String()
		}
		versions = append(versions, metadata.Version)
		metadata.CreateTime = now
		metadata.UpdateTime = now
		args = append(args, metadata.Summary, metadata.Version, metadata.Description, metadata.Path, metadata.ServerURL, metadata.Method,
			metadata.APISpec, metadata.CreateUser, metadata.CreateTime, metadata.UpdateUser, metadata.UpdateTime)
	}
	row, err := exec(ctx, query, args...)
	if err != nil {
		err = errors.Wrapf(err, "insert operator failed, err: %v", err)
		return
	}
	_, err = checkAffected(row)
	return
}

// SelectByVersion 根据版本获取API元数据
func (a *apiMetadataDB) SelectByVersion(ctx context.Context, version string) (has bool, metadata *model.APIMetadataDB, err error) {
	query := "SELECT `f_summary`, `f_version`, `f_description`, `f_path`, `f_svc_url`, `f_method`, `f_api_spec`, " +
		"`f_create_user`, `f_create_time`, `f_update_user`, `f_update_time` FROM `%s`.`t_metadata_api` WHERE `f_version` = ?"
	query = fmt.Sprintf(query, a.dbName)
	metadata = &model.APIMetadataDB{}
	err = a.dbPool.QueryRowContext(ctx, query, version).Scan(&metadata.Summary, &metadata.Version, &metadata.Description, &metadata.Path, &metadata.ServerURL, &metadata.Method,
		&metadata.APISpec, &metadata.CreateUser, &metadata.CreateTime, &metadata.UpdateUser, &metadata.UpdateTime)
	has, err = checkHasQuery(query, err)
	return
}

// UpdateByVersion 根据版本更新API元数据
func (a *apiMetadataDB) UpdateByVersion(ctx context.Context, tx *sql.Tx, version string, metadata *model.APIMetadataDB) error {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	query := "UPDATE `%s`.`t_metadata_api` SET `f_summary` = ?, `f_description` = ?, `f_path` = ?, `f_svc_url` = ?, `f_method` = ?, `f_api_spec` = ?, " +
		"`f_update_user` = ?, `f_update_time` = ? WHERE `f_version` = ?"
	query = fmt.Sprintf(query, a.dbName)
	_, err := exec(ctx, query, metadata.Summary, metadata.Description, metadata.Path, metadata.ServerURL, metadata.Method, metadata.APISpec,
		metadata.UpdateUser, time.Now().UnixNano(), version)
	return err
}

func (a *apiMetadataDB) UpdateByID(ctx context.Context, tx *sql.Tx, id int64, metadata *model.APIMetadataDB) error {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	query := "UPDATE `%s`.`t_metadata_api` SET `f_version` =?, `f_summary` =?, `f_description` =?, `f_path` =?, `f_svc_url` =?, `f_method` =?, `f_api_spec` =?, " +
		"`f_update_user` =?, `f_update_time` =? WHERE `f_id` =?"
	query = fmt.Sprintf(query, a.dbName)
	_, err := exec(ctx, query, metadata.Version, metadata.Summary, metadata.Description, metadata.Path, metadata.ServerURL, metadata.Method, metadata.APISpec,
		metadata.UpdateUser, time.Now().UnixNano(), id)
	return err
}

func (a *apiMetadataDB) DeleteByVersion(ctx context.Context, tx *sql.Tx, version string) error {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	query := "DELETE FROM `%s`.`t_metadata_api` WHERE `f_version` =?"
	query = fmt.Sprintf(query, a.dbName)
	_, err := exec(ctx, query, version)
	return err
}

// DeleteByVersions 根据版本删除API元数据
func (a *apiMetadataDB) DeleteByVersions(ctx context.Context, tx *sql.Tx, versions []string) (err error) {
	exec := a.dbPool.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	query := "DELETE FROM `%s`.`t_metadata_api` WHERE `f_version` IN (?"
	query = fmt.Sprintf(query, a.dbName)
	if len(versions) > 1 {
		query += strings.Repeat(",?", len(versions)-1)
	}
	query += strings.Repeat(")", 1)
	var args []interface{}
	for _, version := range versions {
		args = append(args, version)
	}
	row, err := exec(ctx, query, args...)
	if err != nil {
		err = errors.Wrapf(err, "delete metadata err")
		return
	}
	_, err = checkAffected(row)
	return
}

// SelectListByVersion 查询多个版本
func (a *apiMetadataDB) SelectListByVersion(ctx context.Context, versions []string) (list []*model.APIMetadataDB, err error) {
	orm := a.orm
	values := []interface{}{}
	for _, version := range versions {
		values = append(values, version)
	}
	list = []*model.APIMetadataDB{}
	if len(values) == 0 {
		return
	}
	err = orm.Select().From(tableAPIMetadata).WhereIn("f_version", values...).Get(ctx, &list)
	if err != nil {
		err = errors.Wrapf(err, "select api metadata list by version error")
	}
	return
}
