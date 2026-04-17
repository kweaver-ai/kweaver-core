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
)

type skillReleaseHistoryDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillReleaseHistoryOnce sync.Once
	skillReleaseHistoryInst model.ISkillReleaseHistoryDB
)

const tbSkillReleaseHistory = "t_skill_release_history"

func NewSkillReleaseHistoryDB() model.ISkillReleaseHistoryDB {
	skillReleaseHistoryOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillReleaseHistoryInst = &skillReleaseHistoryDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillReleaseHistoryInst
}

func (s *skillReleaseHistoryDB) Insert(ctx context.Context, tx *sql.Tx, history *model.SkillReleaseHistoryDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbSkillReleaseHistory).Values(map[string]interface{}{
		"f_skill_id":      history.SkillID,
		"f_version":       history.Version,
		"f_skill_release": history.SkillRelease,
		"f_release_desc":  history.ReleaseDesc,
		"f_create_user":   history.CreateUser,
		"f_create_time":   history.CreateTime,
		"f_update_user":   history.UpdateUser,
		"f_update_time":   history.UpdateTime,
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

func (s *skillReleaseHistoryDB) SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) (histories []*model.SkillReleaseHistoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	histories = []*model.SkillReleaseHistoryDB{}
	err = orm.Select().From(tbSkillReleaseHistory).WhereEq("f_skill_id", skillID).OrderByDesc("f_create_time").Get(ctx, &histories)
	return histories, err
}

func (s *skillReleaseHistoryDB) SelectBySkillIDAndVersion(ctx context.Context, tx *sql.Tx, skillID, version string) (history *model.SkillReleaseHistoryDB, err error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	history = &model.SkillReleaseHistoryDB{}
	err = orm.Select().From(tbSkillReleaseHistory).
		WhereEq("f_skill_id", skillID).
		WhereEq("f_version", version).
		OrderByDesc("f_update_time").
		Limit(1).
		First(ctx, history)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (s *skillReleaseHistoryDB) DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillReleaseHistory).WhereEq("f_id", id).Execute(ctx)
	return err
}

func (s *skillReleaseHistoryDB) DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	_, err := orm.Delete().From(tbSkillReleaseHistory).WhereEq("f_skill_id", skillID).Execute(ctx)
	return err
}
