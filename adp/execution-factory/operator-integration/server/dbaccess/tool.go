package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/pkg/errors"
)

type toolDB struct {
	dbPool *sqlx.DB
	orm    *ormhelper.DB
	dbName string
}

var (
	tOnce         sync.Once
	toolDBService model.IToolDB
)

// NewToolDB 创建工具DB
func NewToolDB() model.IToolDB {
	tOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		toolDBService = &toolDB{
			dbPool: dbPool,
			dbName: dbName,
			orm:    orm,
		}
	})
	return toolDBService
}

const (
	tbTool = "t_tool"
)

// InsertTool 添加工具
func (t *toolDB) InsertTool(ctx context.Context, tx *sql.Tx, tool *model.ToolDB) (toolID string, err error) {
	now := time.Now().UnixNano()
	if tool.ToolID == "" {
		tool.ToolID = uuid.NewString()
	}
	toolID = tool.ToolID
	tool.CreateTime = now
	tool.UpdateTime = now
	orm := t.orm
	if tx != nil {
		orm = t.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbTool).Values(map[string]interface{}{
		"f_tool_id":     tool.ToolID,
		"f_box_id":      tool.BoxID,
		"f_name":        tool.Name,
		"f_description": tool.Description,
		"f_source_id":   tool.SourceID,
		"f_source_type": tool.SourceType,
		"f_status":      tool.Status,
		"f_use_rule":    tool.UseRule,
		"f_parameters":  tool.Parameters,
		"f_use_count":   tool.UseCount,
		"f_create_user": tool.CreateUser,
		"f_create_time": tool.CreateTime,
		"f_update_user": tool.UpdateUser,
		"f_update_time": tool.UpdateTime,
		"f_extend_info": tool.ExtendInfo,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert tool error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert tool failed, err: %v", err)
	}
	return
}

// InsertTools 批量添加工具
func (t *toolDB) InsertTools(ctx context.Context, tx *sql.Tx, tools []*model.ToolDB) (toolIDs []string, err error) {
	orm := t.orm
	if tx != nil {
		orm = t.orm.WithTx(tx)
	}
	columns := []string{
		"f_tool_id",
		"f_box_id",
		"f_name",
		"f_description",
		"f_source_id",
		"f_source_type",
		"f_status",
		"f_use_rule",
		"f_parameters",
		"f_use_count",
		"f_create_user",
		"f_create_time",
		"f_update_user",
		"f_update_time",
		"f_extend_info",
	}
	now := time.Now().UnixNano()
	values := [][]interface{}{}
	toolIDs = []string{}
	for _, tool := range tools {
		if tool.ToolID == "" {
			tool.ToolID = uuid.NewString()
		}
		toolIDs = append(toolIDs, tool.ToolID)
		values = append(values, []interface{}{
			tool.ToolID,
			tool.BoxID,
			tool.Name,
			tool.Description,
			tool.SourceID,
			tool.SourceType,
			tool.Status,
			tool.UseRule,
			tool.Parameters,
			tool.UseCount,
			tool.CreateUser,
			now,
			tool.UpdateUser,
			now,
			tool.ExtendInfo,
		})
	}
	row, err := orm.Insert().Into(tbTool).BatchValues(columns, values).Execute(ctx)
	if err != nil {
		err = errors.Wrap(err, "insert tools error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("insert tools failed, err: %v", err)
	}
	return
}

// UpdateTool 更新工具
func (t *toolDB) UpdateTool(ctx context.Context, tx *sql.Tx, tool *model.ToolDB) (err error) {
	now := time.Now().UnixNano()
	tool.UpdateTime = now
	orm := t.orm
	if tx != nil {
		orm = t.orm.WithTx(tx)
	}
	row, err := orm.Update(tbTool).SetData(map[string]interface{}{
		"f_name":        tool.Name,
		"f_description": tool.Description,
		"f_source_id":   tool.SourceID,
		"f_source_type": tool.SourceType,
		"f_status":      tool.Status,
		"f_use_rule":    tool.UseRule,
		"f_parameters":  tool.Parameters,
		"f_use_count":   tool.UseCount,
		"f_update_user": tool.UpdateUser,
		"f_update_time": tool.UpdateTime,
		"f_extend_info": tool.ExtendInfo,
	}).WhereEq("f_tool_id", tool.ToolID).Execute(ctx)
	if err != nil {
		err = errors.Wrap(err, "update tool error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("update tool failed, err: %v", err)
	}
	return
}

// SelectTool 查询工具
func (t *toolDB) SelectTool(ctx context.Context, toolID string) (exist bool, tool *model.ToolDB, err error) {
	tool = &model.ToolDB{}
	orm := t.orm
	err = orm.Select().From(tbTool).WhereEq("f_tool_id", toolID).First(ctx, tool)
	exist, err = checkHasQueryErr(err)
	return
}

func buildQueryConditions(query *ormhelper.SelectBuilder, conditions map[string]interface{}) *ormhelper.SelectBuilder {
	if len(conditions) == 0 {
		return query
	}
	if conditions["name"] != nil {
		name := conditions["name"].(string)
		query = query.WhereLike("f_name", "%"+name+"%")
	}
	if conditions["status"] != nil {
		query = query.WhereEq("f_status", conditions["status"])
	}
	return query
}

// CountToolByBoxID 统计工具箱下工具数量
func (t *toolDB) CountToolByBoxID(ctx context.Context, boxID string, filter map[string]interface{}) (count int64, err error) {
	orm := t.orm
	queryBuilder := orm.Select().From(tbTool).WhereEq("f_box_id", boxID)
	queryBuilder = buildQueryConditions(queryBuilder, filter)
	count, err = queryBuilder.Count(ctx)
	return
}

// SelectToolLisByBoxID 查询工具箱下工具列表
func (t *toolDB) SelectToolLisByBoxID(ctx context.Context, boxID string, filter map[string]interface{}) (tools []*model.ToolDB, err error) {
	orm := t.orm
	queryBuilder := orm.Select().From(tbTool).WhereEq("f_box_id", boxID)
	queryBuilder = buildQueryConditions(queryBuilder, filter)
	var field string
	switch filter["sort_by"] {
	case "name":
		field = "f_name"
	case "create_time":
		field = "f_create_time"
	case "update_time":
		field = "f_update_time"
	default:
		field = "f_create_time"
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
	queryBuilder.Sort(&ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{
				Field: field,
				Order: ormhelper.SortOrder(sortOrder),
			},
		},
	})

	// 条件查询
	if filter["all"] == nil || filter["all"] == false {
		pageSize := filter["limit"].(int)
		offset := filter["offset"].(int)
		queryBuilder.Limit(pageSize).Offset(offset)
	}
	tools = []*model.ToolDB{}
	err = queryBuilder.Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool list error")
	}
	return
}

// SelectToolList 查询工具列表
func (t *toolDB) SelectToolList(ctx context.Context, filter map[string]interface{}) (tools []*model.ToolDB, err error) {
	orm := t.orm
	queryBuilder := orm.Select().From(tbTool)
	queryBuilder = buildQueryConditions(queryBuilder, filter)
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
	queryBuilder.Sort(&ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{
				Field: field,
				Order: ormhelper.SortOrder(sortOrder),
			},
		},
	})

	// 条件查询
	if filter["all"] == nil || filter["all"] == false {
		pageSize := filter["limit"].(int)
		offset := filter["offset"].(int)
		queryBuilder.Limit(pageSize).Offset(offset)
	}
	tools = []*model.ToolDB{}
	err = queryBuilder.Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool list error")
	}
	return
}

// SelectToolBoxIDsByFilter 根据查询条件获取工具箱ID列表
func (t *toolDB) SelectToolBoxIDsByFilter(ctx context.Context, filter map[string]interface{}) (boxIDs []string, err error) {
	orm := t.orm
	queryBuilder := orm.Select("f_box_id", "f_create_time", "f_update_time").From(tbTool)
	queryBuilder = buildQueryConditions(queryBuilder, filter)
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
	queryBuilder.Sort(&ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{
				Field: field,
				Order: ormhelper.SortOrder(sortOrder),
			},
		},
	})
	queryBuilder.GroupBy("f_box_id")
	// 条件查询
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

	toolDBs := []*model.ToolDB{}
	err = queryBuilder.Get(ctx, &toolDBs)
	if err != nil {
		err = errors.Wrapf(err, "select tool box id list error")
	}
	boxIDs = []string{}
	for _, toolDB := range toolDBs {
		boxIDs = append(boxIDs, toolDB.BoxID)
	}
	return
}

// SelectBoxToolByName 根据名称查询工具
func (t *toolDB) SelectBoxToolByName(ctx context.Context, boxID, name string) (exist bool, tool *model.ToolDB, err error) {
	orm := t.orm
	tool = &model.ToolDB{}
	err = orm.Select().From(tbTool).WhereEq("f_box_id", boxID).
		WhereEq("f_name", name).First(ctx, tool)
	exist, err = checkHasQueryErr(err)
	return
}

// SelectToolByBoxID 根据工具箱ID查询工具
func (t *toolDB) SelectToolByBoxID(ctx context.Context, boxID string) (tools []*model.ToolDB, err error) {
	orm := t.orm
	tools = []*model.ToolDB{}
	err = orm.Select().From(tbTool).WhereEq("f_box_id", boxID).Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool list error")
	}
	return
}

// SelectToolNameListByBoxID 根据工具箱ID查询工具名称列表
func (t *toolDB) SelectToolNameListByBoxID(ctx context.Context, boxID []string) (toolNameList map[string][]string, err error) {
	orm := t.orm
	toolDBs := []*model.ToolDB{}
	arrs := []interface{}{}
	for _, id := range boxID {
		arrs = append(arrs, id)
	}
	// 检查工具箱ID是否为空
	if len(arrs) == 0 {
		toolNameList = map[string][]string{}
		return
	}
	err = orm.Select("f_box_id", "f_name").From(tbTool).WhereIn("f_box_id", arrs...).Get(ctx, &toolDBs)
	if err != nil {
		err = errors.Wrapf(err, "select tool name list error")
		return
	}
	toolNameList = map[string][]string{}
	for _, toolDB := range toolDBs {
		if _, ok := toolNameList[toolDB.BoxID]; !ok {
			toolNameList[toolDB.BoxID] = []string{}
		}
		toolNameList[toolDB.BoxID] = append(toolNameList[toolDB.BoxID], toolDB.Name)
	}
	return
}

// DeleteBoxByIDAndTools 删除
func (t *toolDB) DeleteBoxByIDAndTools(ctx context.Context, tx *sql.Tx, boxID string, toolIDs []string) (err error) {
	orm := t.orm
	if tx != nil {
		orm = orm.WithTx(tx)
	}
	args := []interface{}{}
	for _, id := range toolIDs {
		args = append(args, id)
	}
	row, err := orm.Delete().From(tbTool).WhereEq("f_box_id", boxID).WhereIn("f_tool_id", args...).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete tool error")
		return
	}
	_, err = checkAffected(row)
	return
}

// SelectToolBoxByID 获取工具箱内工具信息
func (t *toolDB) SelectToolBoxByID(ctx context.Context, boxID string, toolIDs []string) (tools []*model.ToolDB, err error) {
	orm := t.orm
	tools = []*model.ToolDB{}
	args := []interface{}{}
	for _, id := range toolIDs {
		args = append(args, id)
	}
	if len(args) == 0 {
		return
	}
	err = orm.Select().From(tbTool).WhereEq("f_box_id", boxID).WhereIn("f_tool_id", args...).Get(ctx, &tools)
	return
}

// UpdateToolStatus 更新工具状态
func (t *toolDB) UpdateToolStatus(ctx context.Context, tx *sql.Tx, toolID, status, userID string) (err error) {
	now := time.Now().UnixNano()
	orm := t.orm
	if tx != nil {
		orm = orm.WithTx(tx)
	}
	row, err := orm.Update(tbTool).SetData(map[string]interface{}{
		"f_status":      status,
		"f_update_user": userID,
		"f_update_time": now,
	}).WhereEq("f_tool_id", toolID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update tool status error")
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("update tool status failed, err: %v", err)
	}
	return
}

// SelectToolBoxByIDs 根据工具箱ID查询工具
func (t *toolDB) SelectToolBoxByIDs(ctx context.Context, boxIDs []string) (tools []*model.ToolDB, err error) {
	tools = []*model.ToolDB{}
	args := []interface{}{}
	for _, id := range boxIDs {
		args = append(args, id)
	}
	if len(args) == 0 {
		return
	}
	err = t.orm.Select().From(tbTool).WhereIn("f_box_id", args...).Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool box by ids error")
	}
	return
}

// SelectToolBoxByToolIDs 根据工具ID查询工具箱
func (t *toolDB) SelectToolBoxByToolIDs(ctx context.Context, toolIDs []string) (tools []*model.ToolDB, err error) {
	tools = []*model.ToolDB{}
	args := []interface{}{}
	for _, id := range toolIDs {
		args = append(args, id)
	}
	if len(args) == 0 {
		return
	}
	err = t.orm.Select().From(tbTool).WhereIn("f_tool_id", args...).Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool box by tool ids error")
	}
	return
}

// SelectToolBySource 根据来源类型和来源ID查询工具
func (t *toolDB) SelectToolBySource(ctx context.Context, sourceType model.SourceType, sourceID string) (tools []*model.ToolDB, err error) {
	orm := t.orm
	tools = []*model.ToolDB{}
	err = orm.Select().From(tbTool).WhereEq("f_source_type", sourceType).WhereEq("f_source_id", sourceID).Get(ctx, &tools)
	if err != nil {
		err = errors.Wrapf(err, "select tool by source error")
	}
	return
}
