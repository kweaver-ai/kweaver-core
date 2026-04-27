package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type toolboxDB struct {
	dbPool *sqlx.DB
	dbName string
	orm    *ormhelper.DB
}

var (
	bOnce            sync.Once
	toolboxDBService model.IToolboxDB
)

// NewToolboxDB 创建工具箱DB
func NewToolboxDB() model.IToolboxDB {
	bOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		toolboxDBService = &toolboxDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return toolboxDBService
}

const (
	tbToolBox = "t_toolbox"
)

// InsertToolBox 插入工具箱
func (b *toolboxDB) InsertToolBox(ctx context.Context, tx *sql.Tx, toolbox *model.ToolboxDB) (boxID string, err error) {
	if toolbox.BoxID == "" {
		toolbox.BoxID = uuid.NewString()
	}
	boxID = toolbox.BoxID
	orm := b.orm
	if tx != nil {
		orm = b.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbToolBox).Values(map[string]interface{}{
		"f_box_id":        toolbox.BoxID,
		"f_name":          toolbox.Name,
		"f_description":   toolbox.Description,
		"f_status":        toolbox.Status,
		"f_source":        toolbox.Source,
		"f_svc_url":       toolbox.ServerURL,
		"f_category":      toolbox.Category,
		"f_is_internal":   toolbox.IsInternal,
		"f_create_user":   toolbox.CreateUser,
		"f_create_time":   toolbox.CreateTime,
		"f_update_user":   toolbox.UpdateUser,
		"f_update_time":   toolbox.UpdateTime,
		"f_release_time":  toolbox.ReleaseTime,
		"f_release_user":  toolbox.ReleaseUser,
		"f_metadata_type": toolbox.MetadataType,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert toolbox error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert toolbox failed, err: %v", err)
	}
	return
}

// UpdateToolBox 更新工具箱
func (b *toolboxDB) UpdateToolBox(ctx context.Context, tx *sql.Tx, toolbox *model.ToolboxDB) (err error) {
	now := time.Now().UnixNano()
	toolbox.UpdateTime = now
	orm := b.orm
	if tx != nil {
		orm = b.orm.WithTx(tx)
	}
	row, err := orm.Update(tbToolBox).SetData(map[string]interface{}{
		"f_name":         toolbox.Name,
		"f_description":  toolbox.Description,
		"f_svc_url":      toolbox.ServerURL,
		"f_category":     toolbox.Category,
		"f_update_user":  toolbox.UpdateUser,
		"f_update_time":  toolbox.UpdateTime,
		"f_release_user": toolbox.ReleaseUser,
		"f_release_time": toolbox.ReleaseTime,
		"f_status":       toolbox.Status,
	}).WhereEq("f_box_id", toolbox.BoxID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update toolbox error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("update toolbox failed, err: %v", err)
	}
	return
}

// SelectToolBox 查询工具箱
func (b *toolboxDB) SelectToolBox(ctx context.Context, boxID string) (exist bool, toolbox *model.ToolboxDB, err error) {
	toolbox = &model.ToolboxDB{}
	orm := b.orm
	err = orm.Select().From(tbToolBox).WhereEq("f_box_id", boxID).First(ctx, toolbox)
	exist, err = checkHasQueryErr(err)
	return
}

func (b *toolboxDB) buildQueryConditions(query *ormhelper.SelectBuilder, conditions map[string]interface{}) *ormhelper.SelectBuilder {
	if len(conditions) == 0 {
		return query
	}
	if conditions["create_user"] != nil {
		query = query.WhereEq("f_create_user", conditions["create_user"])
	}
	if conditions["release_user"] != nil {
		query = query.WhereEq("f_release_user", conditions["release_user"])
	}
	if conditions["name"] != nil {
		name := conditions["name"].(string)
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if conditions["status"] != nil {
		query = query.WhereEq("f_status", conditions["status"])
	}
	if conditions["category"] != nil {
		query = query.WhereEq("f_category", conditions["category"])
	}
	if conditions["in"] != nil {
		boxIDs := conditions["in"].([]string)
		if len(boxIDs) == 0 {
			return query
		}
		var arr []interface{}
		for _, id := range boxIDs {
			if id != "" {
				arr = append(arr, id)
			}
		}
		if len(arr) > 0 {
			query = query.WhereIn("f_box_id", arr...)
		}
	}
	return query
}

// CountToolBox 查询工具箱数量
func (b *toolboxDB) CountToolBox(ctx context.Context, filter map[string]interface{}) (count int64, err error) {
	orm := b.orm
	queryBuilder := orm.Select().From(tbToolBox)
	queryBuilder = b.buildQueryConditions(queryBuilder, filter)
	count, err = queryBuilder.Count(ctx)
	return
}

// SelectToolBoxList 查询工具列表
func (b *toolboxDB) SelectToolBoxList(ctx context.Context, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (toolboxList []*model.ToolboxDB, err error) {
	orm := b.orm
	queryBuilder := orm.Select().From(tbToolBox)
	queryBuilder = b.buildQueryConditions(queryBuilder, filter)
	queryBuilder.Cursor(cursor)
	queryBuilder.Sort(sort)
	// 处理分页
	if filter["all"] == nil || filter["all"] == false {
		pageSize, ok := filter["limit"].(int)
		if ok {
			queryBuilder.Limit(pageSize)
		}
		offset, ok := filter["offset"].(int)
		if ok {
			queryBuilder.Offset(offset)
		}
	}
	toolboxList = []*model.ToolboxDB{}
	err = queryBuilder.Get(ctx, &toolboxList)
	if err != nil {
		err = errors.Wrapf(err, "select toolbox error")
	}
	return
}

// SelectToolBoxByName 根据名称查询工具箱
func (b *toolboxDB) SelectToolBoxByName(ctx context.Context, name string, status []string) (bool, *model.ToolboxDB, error) {
	toolbox := &model.ToolboxDB{}
	orm := b.orm
	args := []interface{}{}
	for _, s := range status {
		args = append(args, s)
	}
	err := orm.Select().From(tbToolBox).WhereEq("f_name", name).WhereIn("f_status", args...).First(ctx, toolbox)
	exist, err := checkHasQueryErr(err)
	return exist, toolbox, err
}

// DeleteToolBox 删除工具箱
func (b *toolboxDB) DeleteToolBox(ctx context.Context, tx *sql.Tx, boxID string) (err error) {
	orm := b.orm
	if tx != nil {
		orm = b.orm.WithTx(tx)
	}
	row, err := orm.Delete().From(tbToolBox).WhereEq("f_box_id", boxID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete toolbox error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("delete toolbox failed, err: %v", err)
	}
	return
}

// UpdateToolBoxStatus 更新工具箱状态
func (b *toolboxDB) UpdateToolBoxStatus(ctx context.Context, tx *sql.Tx, boxID, status, userID string) (err error) {
	orm := b.orm
	if tx != nil {
		orm = b.orm.WithTx(tx)
	}
	now := time.Now().UnixNano()
	row, err := orm.Update(tbToolBox).SetData(map[string]interface{}{
		"f_status":       status,
		"f_update_user":  userID,
		"f_update_time":  now,
		"f_release_time": now,
		"f_release_user": userID,
	}).WhereEq("f_box_id", boxID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update toolbox status error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("update toolbox status failed, err: %v", err)
	}
	return
}

// SelectListByBoxIDs 获取工具箱列表
func (b *toolboxDB) SelectListByBoxIDs(ctx context.Context, boxIDs []string, status ...string) (toolboxList []*model.ToolboxDB, err error) {
	toolboxList = []*model.ToolboxDB{}
	orm := b.orm
	values := []interface{}{}
	for _, boxID := range boxIDs {
		values = append(values, boxID)
	}
	if len(values) == 0 {
		return
	}
	query := orm.Select().From(tbToolBox).WhereIn("f_box_id", values...)
	if len(status) > 0 {
		args := []interface{}{}
		for _, s := range status {
			args = append(args, s)
		}
		query = query.WhereIn("f_status", args...)
	}
	err = query.Get(ctx, &toolboxList)
	if err != nil {
		err = errors.Wrapf(err, "select toolbox list error")
	}
	return
}

func (b *toolboxDB) SelectListByBoxIDsFilter(ctx context.Context, boxIDs []string, status string, filter map[string]interface{}) (toolboxList []*model.ToolboxDB, err error) {
	toolboxList = []*model.ToolboxDB{}
	orm := b.orm
	values := []interface{}{}
	for _, boxID := range boxIDs {
		values = append(values, boxID)
	}
	var field string
	switch filter["sort_by"] {
	case "name":
		field = "f_name"
	case "create_time":
		field = "f_create_time"
	case "update_time":
		field = "f_update_time"
	default:
		field = "f_update_time"
	}
	var sortOrder string
	switch filter["sort_order"] {
	case "asc":
		sortOrder = "ASC"
	case "desc":
		sortOrder = "DESC"
	default:
		sortOrder = "DESC"
	}
	err = orm.Select().From(tbToolBox).WhereIn("f_box_id", values...).
		WhereEq("f_status", status).Sort(&ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{
				Field: field,
				Order: ormhelper.SortOrder(sortOrder),
			},
		},
	}).Get(ctx, &toolboxList)
	if err != nil {
		err = errors.Wrapf(err, "select toolbox list error")
	}
	return
}

// SelectListByNamesAndStatus 根据名称和状态查询工具箱列表
func (b *toolboxDB) SelectListByNamesAndStatus(ctx context.Context, names []string, status ...string) (toolboxList []*model.ToolboxDB, err error) {
	toolboxList = []*model.ToolboxDB{}
	orm := b.orm
	values := []interface{}{}
	for _, name := range names {
		values = append(values, name)
	}
	query := orm.Select().From(tbToolBox).WhereIn("f_name", values...)
	if len(status) > 0 {
		args := []interface{}{}
		for _, s := range status {
			args = append(args, s)
		}
		query = query.WhereIn("f_status", args...)
	}
	err = query.Get(ctx, &toolboxList)
	if err != nil {
		err = errors.Wrapf(err, "select toolbox list error")
	}
	return toolboxList, err
}
