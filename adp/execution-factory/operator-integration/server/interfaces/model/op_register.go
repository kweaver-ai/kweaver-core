// Package model 定义数据库操作接口
// @file op_registry.go
// @description: 定义t_op_registry表操作接口
package model

//go:generate mockgen -source=op_register.go -destination=../../mocks/model_op_register.go -package=mocks
import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// OperatorRegisterDB 算子注册数据库
type OperatorRegisterDB struct {
	ID              int64  `json:"f_id" db:"f_id"`
	OperatorID      string `json:"f_op_id" db:"f_op_id"`
	Name            string `json:"f_name" db:"f_name"` // 算子名称
	MetadataVersion string `json:"f_metadata_version" db:"f_metadata_version"`
	MetadataType    string `json:"f_metadata_type" db:"f_metadata_type"`
	Status          string `json:"f_status" db:"f_status"`
	OperatorType    string `json:"f_operator_type" db:"f_operator_type"`
	ExecutionMode   string `json:"f_execution_mode" db:"f_execution_mode"`
	Category        string `json:"f_category" db:"f_category"`
	Source          string `json:"f_source" db:"f_source"`
	ExecuteControl  string `json:"f_execute_control" db:"f_execute_control"`
	ExtendInfo      string `json:"f_extend_info" db:"f_extend_info"`
	CreateUser      string `json:"f_create_user" db:"f_create_user"`
	CreateTime      int64  `json:"f_create_time" db:"f_create_time"`
	UpdateUser      string `json:"f_update_user" db:"f_update_user"`
	UpdateTime      int64  `json:"f_update_time" db:"f_update_time"`
	IsInternal      bool   `json:"f_is_internal" db:"f_is_internal"`
	IsDataSource    bool   `json:"f_is_data_source" db:"f_is_data_source"` // 是否为数据源算子
}

// GetBizID 获取业务ID
func (or *OperatorRegisterDB) GetBizID() string {
	return or.OperatorID
}

// IOperatorRegisterDB 算子管理数据库
type IOperatorRegisterDB interface {
	// InsertOperator 插入算子
	// @directUpdate：是否直接更新
	InsertOperator(ctx context.Context, tx *sql.Tx, operator *OperatorRegisterDB) (opID string, err error)
	// SelectByNameAndStatus 根据算子名称获取算子
	SelectByNameAndStatus(ctx context.Context, tx *sql.Tx, name, status string) (has bool, operator *OperatorRegisterDB, err error)
	// SelectByOperatorIDAndVersion 根据算子ID和版本获取算子
	SelectByOperatorIDAndVersion(ctx context.Context, operatorID, version string) (has bool, operator *OperatorRegisterDB, err error)
	// SelectByOperatorID 根据算子ID
	SelectByOperatorID(ctx context.Context, tx *sql.Tx, operatorID string) (has bool, operator *OperatorRegisterDB, err error)
	// CountByWhereClause 统计算子数量
	CountByWhereClause(ctx context.Context, conditions map[string]interface{}) (count int64, err error)
	// SelectListPage 分页查询算子列表
	SelectListPage(ctx context.Context, conditions map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (operatorList []*OperatorRegisterDB, err error)
	// UpdateOperatorStatus 更新算子状态
	UpdateOperatorStatus(ctx context.Context, tx *sql.Tx, operator *OperatorRegisterDB, userID string) error
	// UpdateByOperatorID 根据算子ID和版本更新算子
	UpdateByOperatorID(ctx context.Context, tx *sql.Tx, operator *OperatorRegisterDB) error
	// UpdateNameByOperatorID 根据算子ID更新算子名称
	UpdateNameByOperatorID(ctx context.Context, tx *sql.Tx, operatorID, name string, updateUser string) error
	// DeleteByOperatorID 根据算子ID
	DeleteByOperatorID(ctx context.Context, tx *sql.Tx, operatorID string) error
	SelectByOperatorIDs(ctx context.Context, operatorIDs []string) (operatorList []*OperatorRegisterDB, err error)
	// SelectListByNamesAndStatus 根据算子名称和状态获取算子
	SelectListByNamesAndStatus(ctx context.Context, names []string, status string) (operatorList []*OperatorRegisterDB, err error)
	// // CountByWhereClauseAndIDs 统计算子数量
	// CountByWhereClauseAndIDs(ctx context.Context, conditions map[string]interface{}, operatorIDs []string) (count int64, err error)
	// // SelectListPageByIDs 基于IN分页查询算子列表
	// SelectListPageByIDs(ctx context.Context, pageSize, offset int, conditions map[string]interface{}, operatorIDs []string, orderBy, sortOrder string) (operatorList []*OperatorRegisterDB, err error)
}
