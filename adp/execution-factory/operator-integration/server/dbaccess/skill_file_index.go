package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type skillFileIndexDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillFileOnce sync.Once
	skillFileInst model.ISkillFileIndex
)

const tbSkillFileIndex = "t_skill_file_index"

func NewSkillFileIndexDB() model.ISkillFileIndex {
	skillFileOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillFileInst = &skillFileIndexDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillFileInst
}

func (s *skillFileIndexDB) InsertSkillFile(ctx context.Context, tx *sql.Tx, file *model.SkillFileIndexDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	now := time.Now().UnixNano()
	file.CreateTime = now
	file.UpdateTime = now
	_, err := orm.Insert().Into(tbSkillFileIndex).Values(map[string]interface{}{
		"f_skill_id":       file.SkillID,
		"f_skill_version":  file.SkillVersion,
		"f_rel_path":       file.RelPath,
		"f_path_hash":      file.PathHash,
		"f_storage_id":     file.StorageID,
		"f_storage_key":    file.StorageKey,
		"f_file_type":      file.FileType,
		"f_content_sha256": file.ContentSHA256,
		"f_mime_type":      file.MimeType,
		"f_size":           file.Size,
		"f_create_time":    file.CreateTime,
		"f_update_time":    file.UpdateTime,
	}).Execute(ctx)
	return err
}

func (s *skillFileIndexDB) BatchInsertSkillFiles(ctx context.Context, tx *sql.Tx, files []*model.SkillFileIndexDB) error {
	for _, file := range files {
		if err := s.InsertSkillFile(ctx, tx, file); err != nil {
			return err
		}
	}
	return nil
}

func (s *skillFileIndexDB) UpdateSkillFile(ctx context.Context, tx *sql.Tx, file *model.SkillFileIndexDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	file.UpdateTime = time.Now().UnixNano()
	_, err := orm.Update(tbSkillFileIndex).SetData(map[string]interface{}{
		"f_storage_id":     file.StorageID,
		"f_storage_key":    file.StorageKey,
		"f_file_type":      file.FileType,
		"f_content_sha256": file.ContentSHA256,
		"f_mime_type":      file.MimeType,
		"f_size":           file.Size,
		"f_update_time":    file.UpdateTime,
	}).WhereEq("f_skill_id", file.SkillID).WhereEq("f_skill_version", file.SkillVersion).WhereEq("f_rel_path", file.RelPath).Execute(ctx)
	return err
}

func (s *skillFileIndexDB) SelectSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID, version string) (files []*model.SkillFileIndexDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	files = []*model.SkillFileIndexDB{}
	err = orm.Select().From(tbSkillFileIndex).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).Get(ctx, &files)
	return files, err
}

func (s *skillFileIndexDB) SelectSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, version, relPath string) (file *model.SkillFileIndexDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	file = &model.SkillFileIndexDB{}
	err = orm.Select().From(tbSkillFileIndex).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).WhereEq("f_rel_path", relPath).First(ctx, file)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *skillFileIndexDB) SelectSkillFileByPathHash(ctx context.Context, tx *sql.Tx, skillID, version, pathHash string) (file *model.SkillFileIndexDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	file = &model.SkillFileIndexDB{}
	err = orm.Select().From(tbSkillFileIndex).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).WhereEq("f_path_hash", pathHash).First(ctx, file)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *skillFileIndexDB) DeleteSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID string, version string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillFileIndex).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).Execute(ctx)
	return err
}

func (s *skillFileIndexDB) DeleteSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, version, relPath string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillFileIndex).WhereEq("f_skill_id", skillID).WhereEq("f_skill_version", version).WhereEq("f_rel_path", relPath).Execute(ctx)
	return err
}
