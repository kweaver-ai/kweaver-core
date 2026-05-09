package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

//go:generate mockgen -source=skill_index_build_task.go -destination=../../mocks/model_skill_index_build_task.go -package=mocks

type SkillIndexBuildTaskDB struct {
	ID               int64  `json:"id" db:"f_id"`
	TaskID           string `json:"task_id" db:"f_task_id"`
	Status           string `json:"status" db:"f_status"`
	ExecuteType      string `json:"execute_type" db:"f_execute_type"`
	TotalCount       int64  `json:"total_count" db:"f_total_count"`
	SuccessCount     int64  `json:"success_count" db:"f_success_count"`
	DeleteCount      int64  `json:"delete_count" db:"f_delete_count"`
	FailedCount      int64  `json:"failed_count" db:"f_failed_count"`
	RetryCount       int64  `json:"retry_count" db:"f_retry_count"`
	MaxRetry         int64  `json:"max_retry" db:"f_max_retry"`
	CursorUpdateTime int64  `json:"cursor_update_time" db:"f_cursor_update_time"`
	CursorSkillID    string `json:"cursor_skill_id" db:"f_cursor_skill_id"`
	ErrorMsg         string `json:"error_msg" db:"f_error_msg"`
	CreateUser       string `json:"create_user" db:"f_create_user"`
	CreateTime       int64  `json:"create_time" db:"f_create_time"`
	UpdateTime       int64  `json:"update_time" db:"f_update_time"`
	LastFinishedTime int64  `json:"last_finished_time" db:"f_last_finished_time"`
}

type ISkillIndexBuildTaskDB interface {
	Insert(ctx context.Context, tx *sql.Tx, task *SkillIndexBuildTaskDB) error
	SelectByTaskID(ctx context.Context, tx *sql.Tx, taskID string) (*SkillIndexBuildTaskDB, error)
	SelectRunningTask(ctx context.Context, tx *sql.Tx) (*SkillIndexBuildTaskDB, error)
	SelectLatestCompletedIncrementalTask(ctx context.Context, tx *sql.Tx) (*SkillIndexBuildTaskDB, error)
	SelectLatestCompletedFullTask(ctx context.Context, tx *sql.Tx) (*SkillIndexBuildTaskDB, error)
	DeleteFinishedTasksBefore(ctx context.Context, tx *sql.Tx, cutoff int64) (int64, error)
	SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*SkillIndexBuildTaskDB, error)
	CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (int64, error)
	UpdateByTaskID(ctx context.Context, tx *sql.Tx, task *SkillIndexBuildTaskDB) error
}
