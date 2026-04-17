package model

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=skill_release_history.go -destination=../../mocks/model_skill_release_history.go -package=mocks

// SkillReleaseHistoryDB Skill 发布历史快照
type SkillReleaseHistoryDB struct {
	ID           int64  `json:"id" db:"f_id"`
	SkillID      string `json:"skill_id" db:"f_skill_id"`
	Version      string `json:"version" db:"f_version"`
	SkillRelease string `json:"skill_release" db:"f_skill_release"`
	ReleaseDesc  string `json:"release_desc" db:"f_release_desc"`
	CreateTime   int64  `json:"create_time" db:"f_create_time"`
	CreateUser   string `json:"create_user" db:"f_create_user"`
	UpdateTime   int64  `json:"update_time" db:"f_update_time"`
	UpdateUser   string `json:"update_user" db:"f_update_user"`
}

// ISkillReleaseHistoryDB Skill 发布历史表操作接口
type ISkillReleaseHistoryDB interface {
	Insert(ctx context.Context, tx *sql.Tx, history *SkillReleaseHistoryDB) error
	SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) (histories []*SkillReleaseHistoryDB, err error)
	SelectBySkillIDAndVersion(ctx context.Context, tx *sql.Tx, skillID, version string) (history *SkillReleaseHistoryDB, err error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error
	DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error
}
