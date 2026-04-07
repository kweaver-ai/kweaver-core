package model

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=skill_file_index.go -destination=../../mocks/model_skill_file_index.go -package=mocks

// SkillFileIndexDB Skill 文件索引表
type SkillFileIndexDB struct {
	ID            int64  `json:"id" db:"f_id"`
	SkillID       string `json:"skill_id" db:"f_skill_id"`
	SkillVersion  string `json:"skill_version" db:"f_skill_version"`
	RelPath       string `json:"rel_path" db:"f_rel_path"`
	PathHash      string `json:"path_hash" db:"f_path_hash"`
	StorageID     string `json:"storage_id" db:"f_storage_id"`
	StorageKey    string `json:"storage_key" db:"f_storage_key"`
	FileType      string `json:"file_type" db:"f_file_type"`
	ContentSHA256 string `json:"content_sha256" db:"f_content_sha256"`
	MimeType      string `json:"mime_type" db:"f_mime_type"`
	Size          int64  `json:"size" db:"f_size"`
	CreateTime    int64  `json:"create_time" db:"f_create_time"`
	UpdateTime    int64  `json:"update_time" db:"f_update_time"`

	// Content 仅在注册或回写对象存储时使用，不入库
	Content []byte `json:"-" db:"-"`
}

// ISkillFileIndex Skill 文件索引接口
type ISkillFileIndex interface {
	InsertSkillFile(ctx context.Context, tx *sql.Tx, file *SkillFileIndexDB) error
	BatchInsertSkillFiles(ctx context.Context, tx *sql.Tx, files []*SkillFileIndexDB) error
	UpdateSkillFile(ctx context.Context, tx *sql.Tx, file *SkillFileIndexDB) error
	SelectSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID, version string) (files []*SkillFileIndexDB, err error)
	SelectSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, version, relPath string) (file *SkillFileIndexDB, err error)
	SelectSkillFileByPathHash(ctx context.Context, tx *sql.Tx, skillID, version, pathHash string) (file *SkillFileIndexDB, err error)
	DeleteSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID, version string) error
	DeleteSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, version, relPath string) error
}
