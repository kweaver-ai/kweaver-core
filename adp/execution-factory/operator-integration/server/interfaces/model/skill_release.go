package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

//go:generate mockgen -source=skill_release.go -destination=../../mocks/model_skill_release.go -package=mocks

// SkillReleaseDB Skill 当前发布态快照
type SkillReleaseDB struct {
	ID           int64  `json:"id" db:"f_id"`
	SkillID      string `json:"skill_id" db:"f_skill_id"`
	Name         string `json:"name" db:"f_name"`
	Description  string `json:"description" db:"f_description"`
	SkillContent string `json:"skill_content" db:"f_skill_content"`
	Version      string `json:"version" db:"f_version"`
	Category     string `json:"category" db:"f_category"`
	Source       string `json:"source" db:"f_source"`
	ExtendInfo   string `json:"extend_info" db:"f_extend_info"`
	Dependencies string `json:"dependencies" db:"f_dependencies"`
	FileManifest string `json:"file_manifest" db:"f_file_manifest"`
	Status       string `json:"status" db:"f_status"`
	CreateTime   int64  `json:"create_time" db:"f_create_time"`
	CreateUser   string `json:"create_user" db:"f_create_user"`
	UpdateTime   int64  `json:"update_time" db:"f_update_time"`
	UpdateUser   string `json:"update_user" db:"f_update_user"`
	ReleaseTime  int64  `json:"release_time" db:"f_release_time"`
	ReleaseUser  string `json:"release_user" db:"f_release_user"`
	ReleaseDesc  string `json:"release_desc" db:"f_release_desc"`
}

// GetBizID 获取业务 ID
func (s *SkillReleaseDB) GetBizID() string {
	return s.SkillID
}

// ISkillReleaseDB Skill 发布表操作接口
type ISkillReleaseDB interface {
	Insert(ctx context.Context, tx *sql.Tx, release *SkillReleaseDB) error
	UpdateBySkillID(ctx context.Context, tx *sql.Tx, release *SkillReleaseDB) error
	SelectBySkillID(ctx context.Context, tx *sql.Tx, skillID string) (release *SkillReleaseDB, err error)
	SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
		sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (releases []*SkillReleaseDB, err error)
	CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error)
	DeleteBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error
}
