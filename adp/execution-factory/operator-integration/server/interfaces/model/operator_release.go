package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// OperatorReleaseDB 算子发布表
//
//go:generate mockgen -source=operator_release.go -destination=../../mocks/model_operator_release.go -package=mocks
type OperatorReleaseDB struct {
	ID              int64  `json:"id" db:"f_id"`                             // 主键
	OpID            string `json:"op_id" db:"f_op_id"`                       // 算子ID
	Name            string `json:"name" db:"f_name"`                         // 算子名称
	MetadataVersion string `json:"metadata_version" db:"f_metadata_version"` // 元数据版本
	MetadataType    string `json:"metadata_type" db:"f_metadata_type"`       // 元数据类型
	Status          string `json:"status" db:"f_status"`                     // 元数据状态
	OperatorType    string `json:"operator_type" db:"f_operator_type"`       // 算子类型
	ExecutionMode   string `json:"execution_mode" db:"f_execution_mode"`     // 执行模式
	ExecuteControl  string `json:"execute_control" db:"f_execute_control"`   // 执行控制
	Category        string `json:"category" db:"f_category"`                 // 算子分类
	Source          string `json:"source" db:"f_source"`                     // 算子来源
	ExtendInfo      string `json:"extend_info" db:"f_extend_info"`           // 扩展信息
	CreateTime      int64  `json:"create_time" db:"f_create_time"`           // 创建时间
	UpdateTime      int64  `json:"update_time" db:"f_update_time"`           // 更新时间
	ReleaseTime     int64  `json:"release_time" db:"f_release_time"`         // 发布时间
	CreateUser      string `json:"create_user" db:"f_create_user"`           // 创建用户
	UpdateUser      string `json:"update_user" db:"f_update_user"`           // 更新用户
	ReleaseUser     string `json:"release_user" db:"f_release_user"`         // 发布用户
	Tag             int    `json:"tag" db:"f_tag"`                           // 版本
	IsInternal      bool   `json:"is_internal" db:"f_is_internal"`           // 是否为内部算子
	IsDataSource    bool   `json:"is_data_source" db:"f_is_data_source"`     // 是否为数据源算子
}

// GetBizID 获取业务ID
func (ore *OperatorReleaseDB) GetBizID() string {
	return ore.OpID
}

// IOperatorReleaseDB 算子发布表操作接口
type IOperatorReleaseDB interface {
	// BatchInsert 批量插入算子发布信息
	BatchInsert(ctx context.Context, tx *sql.Tx, operator []*OperatorReleaseDB) (opIDs []string, err error)
	// Insert 插入算子发布信息
	Insert(ctx context.Context, tx *sql.Tx, operator *OperatorReleaseDB) (err error)
	// UpdateByOpID 更新算子发布信息
	UpdateByOpID(ctx context.Context, tx *sql.Tx, operator *OperatorReleaseDB) error
	// DeleteByOpID 删除算子发布信息
	DeleteByOpID(ctx context.Context, tx *sql.Tx, opID string) error
	// SelectByOpID 根据算子ID查询算子发布信息
	SelectByOpID(ctx context.Context, opID string) (exist bool, releaseDB *OperatorReleaseDB, err error)
	// SelectByName 根据算子名称查询算子发布信息
	SelectByName(ctx context.Context, tx *sql.Tx, name string) (exist bool, releaseDB *OperatorReleaseDB, err error)
	// CountByWhereClause 根据条件查询算子发布信息数量
	CountByWhereClause(ctx context.Context, conditions map[string]interface{}) (count int64, err error)
	// SelectByWhereClause 根据条件查询算子发布信息
	SelectByWhereClause(ctx context.Context, conditions map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (releaseList []*OperatorReleaseDB, err error)
}
