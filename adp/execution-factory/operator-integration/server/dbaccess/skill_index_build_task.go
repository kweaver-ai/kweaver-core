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
)

type skillIndexBuildTaskDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	skillIndexBuildTaskOnce sync.Once
	skillIndexBuildTaskInst model.ISkillIndexBuildTaskDB
)

const tbSkillIndexBuildTask = "t_skill_index_build_task"

func NewSkillIndexBuildTaskDB() model.ISkillIndexBuildTaskDB {
	skillIndexBuildTaskOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		skillIndexBuildTaskInst = &skillIndexBuildTaskDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return skillIndexBuildTaskInst
}

func (s *skillIndexBuildTaskDB) Insert(ctx context.Context, tx *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	now := time.Now().UnixNano()
	task.CreateTime = now
	task.UpdateTime = now
	row, err := orm.Insert().Into(tbSkillIndexBuildTask).Values(map[string]interface{}{
		"f_task_id":            task.TaskID,
		"f_status":             task.Status,
		"f_execute_type":       task.ExecuteType,
		"f_total_count":        task.TotalCount,
		"f_success_count":      task.SuccessCount,
		"f_delete_count":       task.DeleteCount,
		"f_failed_count":       task.FailedCount,
		"f_retry_count":        task.RetryCount,
		"f_max_retry":          task.MaxRetry,
		"f_cursor_update_time": task.CursorUpdateTime,
		"f_cursor_skill_id":    task.CursorSkillID,
		"f_error_msg":          task.ErrorMsg,
		"f_create_user":        task.CreateUser,
		"f_create_time":        task.CreateTime,
		"f_update_time":        task.UpdateTime,
		"f_last_finished_time": task.LastFinishedTime,
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

func (s *skillIndexBuildTaskDB) SelectByTaskID(ctx context.Context, tx *sql.Tx, taskID string) (*model.SkillIndexBuildTaskDB, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	task := &model.SkillIndexBuildTaskDB{}
	err := orm.Select().From(tbSkillIndexBuildTask).WhereEq("f_task_id", taskID).First(ctx, task)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *skillIndexBuildTaskDB) SelectRunningTask(ctx context.Context, tx *sql.Tx) (*model.SkillIndexBuildTaskDB, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	task := &model.SkillIndexBuildTaskDB{}
	err := orm.Select().From(tbSkillIndexBuildTask).
		WhereIn("f_status", interfaces.SkillIndexBuildStatusPending.String(), interfaces.SkillIndexBuildStatusRunning.String()).
		OrderByDesc("f_create_time").
		Limit(1).
		First(ctx, task)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *skillIndexBuildTaskDB) SelectLatestCompletedIncrementalTask(ctx context.Context, tx *sql.Tx) (*model.SkillIndexBuildTaskDB, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	task := &model.SkillIndexBuildTaskDB{}
	err := orm.Select().From(tbSkillIndexBuildTask).
		WhereEq("f_execute_type", interfaces.SkillIndexBuildExecuteTypeIncremental.String()).
		WhereEq("f_status", interfaces.SkillIndexBuildStatusCompleted.String()).
		OrderByDesc("f_last_finished_time").
		OrderByDesc("f_update_time").
		Limit(1).
		First(ctx, task)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *skillIndexBuildTaskDB) SelectLatestCompletedFullTask(ctx context.Context, tx *sql.Tx) (*model.SkillIndexBuildTaskDB, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	task := &model.SkillIndexBuildTaskDB{}
	err := orm.Select().From(tbSkillIndexBuildTask).
		WhereEq("f_execute_type", interfaces.SkillIndexBuildExecuteTypeFull.String()).
		WhereEq("f_status", interfaces.SkillIndexBuildStatusCompleted.String()).
		OrderByDesc("f_last_finished_time").
		OrderByDesc("f_update_time").
		Limit(1).
		First(ctx, task)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *skillIndexBuildTaskDB) DeleteFinishedTasksBefore(ctx context.Context, tx *sql.Tx, cutoff int64) (int64, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	row, err := orm.Delete().From(tbSkillIndexBuildTask).
		WhereIn("f_status",
			interfaces.SkillIndexBuildStatusCompleted.String(),
			interfaces.SkillIndexBuildStatusFailed.String(),
			interfaces.SkillIndexBuildStatusCanceled.String(),
		).
		WhereGt("f_last_finished_time", 0).
		WhereLt("f_last_finished_time", cutoff).
		Execute(ctx)
	if err != nil {
		return 0, err
	}
	affected, err := row.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (s *skillIndexBuildTaskDB) SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*model.SkillIndexBuildTaskDB, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillIndexBuildTask)
	query = s.applyFilterConditions(query, filter).Cursor(cursor).Sort(sort)
	if filter["all"] == nil || filter["all"] == false {
		if limit, ok := filter["limit"].(int); ok {
			query = query.Limit(limit)
		}
		if offset, ok := filter["offset"].(int); ok {
			query = query.Offset(offset)
		}
	}
	taskList := []*model.SkillIndexBuildTaskDB{}
	err := query.Get(ctx, &taskList)
	return taskList, err
}

func (s *skillIndexBuildTaskDB) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (int64, error) {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	query := orm.Select().From(tbSkillIndexBuildTask)
	query = s.applyFilterConditions(query, filter)
	list := []*model.SkillIndexBuildTaskDB{}
	if err := query.Get(ctx, &list); err != nil {
		return 0, err
	}
	return int64(len(list)), nil
}

func (s *skillIndexBuildTaskDB) UpdateByTaskID(ctx context.Context, tx *sql.Tx, task *model.SkillIndexBuildTaskDB) error {
	orm := s.orm
	if tx != nil {
		orm = s.orm.WithTx(tx)
	}
	task.UpdateTime = time.Now().UnixNano()
	_, err := orm.Update(tbSkillIndexBuildTask).SetData(map[string]interface{}{
		"f_status":             task.Status,
		"f_total_count":        task.TotalCount,
		"f_success_count":      task.SuccessCount,
		"f_delete_count":       task.DeleteCount,
		"f_failed_count":       task.FailedCount,
		"f_retry_count":        task.RetryCount,
		"f_max_retry":          task.MaxRetry,
		"f_cursor_update_time": task.CursorUpdateTime,
		"f_cursor_skill_id":    task.CursorSkillID,
		"f_error_msg":          task.ErrorMsg,
		"f_update_time":        task.UpdateTime,
		"f_last_finished_time": task.LastFinishedTime,
	}).WhereEq("f_task_id", task.TaskID).Execute(ctx)
	return err
}

func (s *skillIndexBuildTaskDB) applyFilterConditions(query *ormhelper.SelectBuilder, filter map[string]interface{}) *ormhelper.SelectBuilder {
	if filter == nil {
		return query
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query = query.WhereEq("f_status", status)
	}
	if executeType, ok := filter["execute_type"].(string); ok && executeType != "" {
		query = query.WhereEq("f_execute_type", executeType)
	}
	if createUser, ok := filter["create_user"].(string); ok && createUser != "" {
		query = query.WhereEq("f_create_user", createUser)
	}
	return query
}
