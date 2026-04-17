package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

type skillRepositoryDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillRepoOnce sync.Once
	skillRepoInst model.ISkillRepository
)

const tbSkillRepository = "t_skill_repository"

func NewSkillRepositoryDB() model.ISkillRepository {
	skillRepoOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillRepoInst = &skillRepositoryDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillRepoInst
}

func (s *skillRepositoryDB) InsertSkill(ctx context.Context, tx *sql.Tx, skill *model.SkillRepositoryDB) (skillID string, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	if skill.SkillID == "" {
		skill.SkillID = uuid.NewString()
	}
	now := time.Now().UnixNano()
	skillID = skill.SkillID
	skill.CreateTime = now
	skill.UpdateTime = now
	if skill.Status == "" {
		skill.Status = string(interfaces.BizStatusUnpublish)
	}
	row, err := orm.Insert().Into(tbSkillRepository).Values(map[string]interface{}{
		"f_skill_id":      skill.SkillID,
		"f_name":          skill.Name,
		"f_description":   skill.Description,
		"f_skill_content": skill.SkillContent,
		"f_version":       skill.Version,
		"f_status":        skill.Status,
		"f_source":        skill.Source,
		"f_extend_info":   skill.ExtendInfo,
		"f_dependencies":  skill.Dependencies,
		"f_file_manifest": skill.FileManifest,
		"f_create_time":   skill.CreateTime,
		"f_create_user":   skill.CreateUser,
		"f_update_time":   skill.UpdateTime,
		"f_update_user":   skill.UpdateUser,
		"f_delete_time":   skill.DeleteTime,
		"f_delete_user":   skill.DeleteUser,
		"f_category":      skill.Category,
		"f_is_deleted":    skill.IsDeleted,
	}).Execute(ctx)
	if err != nil {
		return "", err
	}
	ok, err := checkAffected(row)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", sql.ErrNoRows
	}
	return skillID, nil
}

func (s *skillRepositoryDB) UpdateSkill(ctx context.Context, tx *sql.Tx, skill *model.SkillRepositoryDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	skill.UpdateTime = time.Now().UnixNano()
	_, err := orm.Update(tbSkillRepository).SetData(map[string]interface{}{
		"f_name":          skill.Name,
		"f_description":   skill.Description,
		"f_skill_content": skill.SkillContent,
		"f_version":       skill.Version,
		"f_status":        skill.Status,
		"f_source":        skill.Source,
		"f_extend_info":   skill.ExtendInfo,
		"f_dependencies":  skill.Dependencies,
		"f_file_manifest": skill.FileManifest,
		"f_update_time":   skill.UpdateTime,
		"f_update_user":   skill.UpdateUser,
		"f_delete_time":   skill.DeleteTime,
		"f_delete_user":   skill.DeleteUser,
		"f_category":      skill.Category,
		"f_is_deleted":    skill.IsDeleted,
	}).WhereEq("f_skill_id", skill.SkillID).Execute(ctx)
	return err
}

func (s *skillRepositoryDB) UpdateSkillStatus(ctx context.Context, tx *sql.Tx, skillID string, status string, updateUser string) (err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	row, err := orm.Update(tbSkillRepository).SetData(map[string]interface{}{
		"f_status":      status,
		"f_update_time": time.Now().UnixNano(),
		"f_update_user": updateUser,
	}).WhereEq("f_skill_id", skillID).Execute(ctx)
	if err != nil {
		err = errors.Wrap(err, "update skill status error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = errors.Wrap(err, "update skill status check affected error")
		return
	}
	if !ok {
		err = errors.New("update skill status error, no affected rows")
	}
	return
}

func (s *skillRepositoryDB) UpdateSkillDeleted(ctx context.Context, tx *sql.Tx, skillID string, isDeleted bool, updateUser string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	updateData := map[string]interface{}{
		"f_is_deleted":  isDeleted,
		"f_update_time": time.Now().UnixNano(),
		"f_update_user": updateUser,
	}
	_, err := orm.Update(tbSkillRepository).SetData(updateData).WhereEq("f_skill_id", skillID).Execute(ctx)
	return err
}

func (s *skillRepositoryDB) SelectSkillByID(ctx context.Context, tx *sql.Tx, skillID string) (skill *model.SkillRepositoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	skill = &model.SkillRepositoryDB{}
	err = orm.Select().From(tbSkillRepository).WhereEq("f_skill_id", skillID).First(ctx, skill)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return skill, nil
}

func (s *skillRepositoryDB) SelectSkillListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
	sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (skills []*model.SkillRepositoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillRepository)
	query = s.applyFilterConditions(query, filter).Cursor(cursor).Sort(sort)
	if filter["all"] == nil || filter["all"] == false {
		if limit, ok := filter["limit"].(int); ok {
			query = query.Limit(limit)
		}
		if offset, ok := filter["offset"].(int); ok {
			query = query.Offset(offset)
		}
	}
	skills = []*model.SkillRepositoryDB{}
	err = query.Get(ctx, &skills)
	return skills, err
}

func (s *skillRepositoryDB) SelectSkillBuildPage(ctx context.Context, tx *sql.Tx, cursorUpdateTime int64, cursorSkillID string, limit int) (skills []*model.SkillRepositoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillRepository)
	if cursorUpdateTime > 0 || cursorSkillID != "" {
		query = query.Or(func(w *ormhelper.WhereBuilder) {
			w.Gt("f_update_time", cursorUpdateTime)
			w.And(func(sub *ormhelper.WhereBuilder) {
				sub.Eq("f_update_time", cursorUpdateTime)
				sub.Gt("f_skill_id", cursorSkillID)
			})
		})
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.OrderByAsc("f_update_time").OrderByAsc("f_skill_id")
	skills = []*model.SkillRepositoryDB{}
	err = query.Get(ctx, &skills)
	return skills, err
}

func (s *skillRepositoryDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillRepository)
	query = s.applyFilterConditions(query, filter)
	list := []*model.SkillRepositoryDB{}
	err = query.Get(ctx, &list)
	if err != nil {
		return 0, err
	}
	return int64(len(list)), nil
}

func (s *skillRepositoryDB) DeleteSkillByID(ctx context.Context, tx *sql.Tx, skillID string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillRepository).WhereEq("f_skill_id", skillID).Execute(ctx)
	return err
}

func (s *skillRepositoryDB) applyFilterConditions(query *ormhelper.SelectBuilder, filter map[string]interface{}) *ormhelper.SelectBuilder {
	if name, ok := filter["name"].(string); ok && name != "" {
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if source, ok := filter["source"].(string); ok && source != "" {
		query = query.WhereEq("f_source", source)
	}
	if createUser, ok := filter["create_user"].(string); ok && createUser != "" {
		query = query.WhereEq("f_create_user", createUser)
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query = query.WhereEq("f_status", status)
	}
	if category, ok := filter["category"].(string); ok && category != "" {
		query = query.WhereEq("f_category", category)
	}
	if filter["in"] != nil {
		skillIDs := filter["in"].([]string)
		if len(skillIDs) == 0 {
			return query
		}
		var arr []interface{}
		for _, id := range skillIDs {
			if id != "" {
				arr = append(arr, id)
			}
		}
		if len(arr) > 0 {
			query = query.WhereIn("f_skill_id", arr...)
		}
	}
	// 获取未删除的
	query = query.WhereEq("f_is_deleted", false)
	return query
}

func (s *skillRepositoryDB) SelectSkillByName(ctx context.Context, tx *sql.Tx, name string, status []string) (exists bool, skillDB *model.SkillRepositoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	skillDB = &model.SkillRepositoryDB{}
	args := []interface{}{}
	for _, s := range status {
		args = append(args, s)
	}
	err = orm.Select().From(tbSkillRepository).WhereEq("f_name", name).WhereIn("f_status", args...).First(ctx, skillDB)
	exist, err := checkHasQueryErr(err)
	return exist, skillDB, err
}
