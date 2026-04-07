package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// ToolboxDB 工具箱DB
//
//go:generate mockgen -source=toolbox.go -destination=../../mocks/model_toolbox.go -package=mocks
type ToolboxDB struct {
	ID          int64  `json:"id" db:"f_id"`                     // 主键ID
	BoxID       string `json:"box_id" db:"f_box_id"`             // 工具箱ID
	Name        string `json:"name" db:"f_name"`                 // 工具箱名称
	Description string `json:"description" db:"f_description"`   // 工具箱描述
	Source      string `json:"source" db:"f_source"`             // 工具箱来源
	ServerURL   string `json:"server_url" db:"f_svc_url"`        // 工具箱服务地址
	Category    string `json:"category" db:"f_category"`         // 分类
	Status      string `json:"status" db:"f_status"`             // 状态
	IsInternal  bool   `json:"is_internal" db:"f_is_internal"`   // 是否内置
	CreateUser  string `json:"create_user" db:"f_create_user"`   // 创建人
	CreateTime  int64  `json:"create_time" db:"f_create_time"`   // 创建时间
	UpdateUser  string `json:"update_user" db:"f_update_user"`   // 更新人
	UpdateTime  int64  `json:"update_time" db:"f_update_time"`   // 更新时间
	ReleaseUser string `json:"release_user" db:"f_release_user"` // 发布人
	ReleaseTime int64  `json:"release_time" db:"f_release_time"` // 发布时间
	// 工具箱元数据类型
	MetadataType string `json:"metadata_type" db:"f_metadata_type"` // 工具箱元数据类型
}

// GetBizID 获取业务ID
func (b *ToolboxDB) GetBizID() string {
	return b.BoxID
}

// IToolboxDB 工具箱接口
type IToolboxDB interface {
	InsertToolBox(ctx context.Context, tx *sql.Tx, toolbox *ToolboxDB) (boxID string, err error)
	UpdateToolBox(ctx context.Context, tx *sql.Tx, toolbox *ToolboxDB) error
	SelectToolBox(ctx context.Context, boxID string) (bool, *ToolboxDB, error)
	SelectToolBoxList(ctx context.Context, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*ToolboxDB, error)
	DeleteToolBox(ctx context.Context, tx *sql.Tx, boxID string) error
	CountToolBox(ctx context.Context, filter map[string]interface{}) (int64, error)
	SelectToolBoxByName(ctx context.Context, name string, status []string) (bool, *ToolboxDB, error)
	UpdateToolBoxStatus(ctx context.Context, tx *sql.Tx, boxID, status string, updateUser string) (err error)
	SelectListByBoxIDs(ctx context.Context, boxIDs []string, status ...string) ([]*ToolboxDB, error)
	SelectListByBoxIDsFilter(ctx context.Context, boxIDs []string, status string, filter map[string]interface{}) ([]*ToolboxDB, error)
	SelectListByNamesAndStatus(ctx context.Context, names []string, status ...string) ([]*ToolboxDB, error)
}
