package model

import (
	"context"
	"database/sql"
)

// OperatorReleaseHistoryDB 算子发布历史表
//
//go:generate mockgen -source=operator_release_history.go -destination=../../mocks/model_operator_release_history.go -package=mocks
type OperatorReleaseHistoryDB struct {
	ID              int64  `json:"id" db:"f_id"`                             // 主键ID
	OpID            string `json:"op_id" db:"f_op_id"`                       // 算子ID
	MetadataVersion string `json:"metadata_version" db:"f_metadata_version"` // 元数据版本
	MetadataType    string `json:"metadata_type" db:"f_metadata_type"`       // 元数据类型
	OpRelease       string `json:"op_release" db:"f_op_release"`             // 算子发布信息
	Tag             int    `json:"tag" db:"f_tag"`                           // 版本
	CreateTime      int64  `json:"create_time" db:"f_create_time"`           // 创建时间
	UpdateTime      int64  `json:"update_time" db:"f_update_time"`           // 更新时间
	CreateUser      string `json:"create_user" db:"f_create_user"`           // 创建用户
	UpdateUser      string `json:"update_user" db:"f_update_user"`           // 更新用户
}

// IOperatorReleaseHistoryDB 算子发布历史表操作接口
type IOperatorReleaseHistoryDB interface {
	Insert(ctx context.Context, tx *sql.Tx, historyDB *OperatorReleaseHistoryDB) (err error)
	DeleteByOpID(ctx context.Context, tx *sql.Tx, opID string) error
	SelectByOpID(ctx context.Context, opID string) (histories []*OperatorReleaseHistoryDB, err error)
	BatchDeleteByID(ctx context.Context, tx *sql.Tx, ids []int64) error
	SelectByOpIDAndMetdata(ctx context.Context, opID, metadataVersion string) (has bool, historyDB *OperatorReleaseHistoryDB, err error)
	SelectByOpIDAndTag(ctx context.Context, opID string, tag int) (has bool, historyDB *OperatorReleaseHistoryDB, err error)
	UpdateReleaseHistoryByID(ctx context.Context, tx *sql.Tx, historyDB *OperatorReleaseHistoryDB) error
}
