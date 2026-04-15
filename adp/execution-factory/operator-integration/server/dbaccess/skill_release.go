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

type skillReleaseDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillReleaseOnce sync.Once
	skillReleaseInst model.ISkillReleaseDB
)

const tbSkillRelease = "t_skill_release"

func NewSkillReleaseDB() model.ISkillReleaseDB {
	skillReleaseOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillReleaseInst = &skillReleaseDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillReleaseInst
}

func (s *skillReleaseDB) Insert(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	now := time.Now().UnixNano()
	release.ReleaseTime = now
	row, err := orm.Insert().Into(tbSkillRelease).Values(map[string]interface{}{
		"f_skill_id":      release.SkillID,
		"f_name":          release.Name,
		"f_description":   release.Description,
		"f_skill_content": release.SkillContent,
		"f_version":       release.Version,
		"f_category":      release.Category,
		"f_source":        release.Source,
		"f_extend_info":   release.ExtendInfo,
		"f_dependencies":  release.Dependencies,
		"f_file_manifest": release.FileManifest,
		"f_status":        release.Status,
		"f_create_user":   release.CreateUser,
		"f_create_time":   release.CreateTime,
		"f_update_user":   release.UpdateUser,
		"f_update_time":   release.UpdateTime,
		"f_release_user":  release.ReleaseUser,
		"f_release_time":  release.ReleaseTime,
		"f_release_desc":  release.ReleaseDesc,
	}).Execute(ctx)
	if err != nil {
		return err
	}
	ok, err := checkAffected(row)
	if err != nil {
		return err
	}
	if !ok {
		return sql.ErrNoRows
	}
	return nil
}

func (s *skillReleaseDB) UpdateBySkillID(ctx context.Context, tx *sql.Tx, release *model.SkillReleaseDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	release.ReleaseTime = time.Now().UnixNano()
	_, err := orm.Update(tbSkillRelease).SetData(map[string]interface{}{
		"f_name":          release.Name,
		"f_description":   release.Description,
		"f_skill_content": release.SkillContent,
		"f_version":       release.Version,
		"f_category":      release.Category,
		"f_source":        release.Source,
		"f_extend_info":   release.ExtendInfo,
		"f_dependencies":  release.Dependencies,
		"f_file_manifest": release.FileManifest,
		"f_status":        release.Status,
		"f_create_user":   release.CreateUser,
		"f_create_time":   release.CreateTime,
		"f_update_user":   release.UpdateUser,
		"f_update_time":   release.UpdateTime,
		"f_release_user":  release.ReleaseUser,
		"f_release_time":  release.ReleaseTime,
		"f_release_desc":  release.ReleaseDesc,
	}).WhereEq("f_skill_id", release.SkillID).Execute(ctx)
	return err
}

func (s *skillReleaseDB) SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) (release *model.SkillReleaseDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	release = &model.SkillReleaseDB{}
	err = orm.Select().From(tbSkillRelease).WhereEq("f_skill_id", skillID).First(ctx, release)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return release, nil
}

func (s *skillReleaseDB) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
	sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (releases []*model.SkillReleaseDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillRelease)
	query = s.applyFilterConditions(query, filter).Cursor(cursor).Sort(sort)
	if filter["all"] == nil || filter["all"] == false {
		if limit, ok := filter["limit"].(int); ok {
			query = query.Limit(limit)
		}
		if offset, ok := filter["offset"].(int); ok {
			query = query.Offset(offset)
		}
	}
	releases = []*model.SkillReleaseDB{}
	err = query.Get(ctx, &releases)
	return releases, err
}

func (s *skillReleaseDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillRelease)
	query = s.applyFilterConditions(query, filter)
	count, err = query.Count(ctx)
	return count, err
}

func (s *skillReleaseDB) DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillRelease).WhereEq("f_skill_id", skillID).Execute(ctx)
	return err
}

func (s *skillReleaseDB) applyFilterConditions(query *ormhelper.SelectBuilder, filter map[string]interface{}) *ormhelper.SelectBuilder {
	if filter == nil {
		return query
	}
	if name, ok := filter["name"].(string); ok && name != "" {
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query = query.WhereEq("f_status", status)
	}
	if category, ok := filter["category"].(string); ok && category != "" {
		query = query.WhereEq("f_category", category)
	}
	if createUser, ok := filter["create_user"].(string); ok && createUser != "" {
		query = query.WhereEq("f_create_user", createUser)
	}
	if releaseUser, ok := filter["release_user"].(string); ok && releaseUser != "" {
		query = query.WhereEq("f_release_user", releaseUser)
	}
	if source, ok := filter["source"].(string); ok && source != "" {
		query = query.WhereEq("f_source", source)
	}
	if in, ok := filter["in"].([]string); ok && len(in) > 0 {
		arr := make([]interface{}, 0, len(in))
		for _, id := range in {
			if id != "" {
				arr = append(arr, id)
			}
		}
		if len(arr) > 0 {
			query = query.WhereIn("f_skill_id", arr...)
		}
	}
	return query
}
