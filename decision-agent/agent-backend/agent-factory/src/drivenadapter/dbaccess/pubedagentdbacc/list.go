package pubedagentdbacc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/sqlhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
	// "golang.org/x/sync/singleflight" // reserved for future sfg optimization
)

// var _pubListSfg singleflight.Group // unused, reserved for future sfg optimization

// GetPubedList 获取已发布智能体列表
func (repo *pubedAgentRepo) GetPubedList(ctx context.Context, req *pubedreq.PubedAgentListReq) (rt []*dapo.PublishedJoinPo, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	// srCount := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	rt = make([]*dapo.PublishedJoinPo, 0)

	pos := make([]dapo.ReleasePartPo, 0)

	// 1. 构建SQL
	fromClause, whereSql, whereArgs, err := repo.buildSql(req)
	if err != nil {
		return
	}

	// 2. 构建完整的查询SQL
	releasePartPO := &dapo.ReleasePartPo{}
	otherFields := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(releasePartPO), "r")
	rawSql := fmt.Sprintf("SELECT %s %s", otherFields, fromClause)

	if len(whereSql) > 0 {
		rawSql = fmt.Sprintf("%s WHERE %s", rawSql, whereSql)
	}
	// 3. 添加排序
	rawSql = fmt.Sprintf("%s ORDER BY r.f_update_time DESC,r.f_id DESC", rawSql)

	// 4. 添加分页
	rawSql = fmt.Sprintf("%s LIMIT %d ", rawSql, req.Size)

	//// 5. 查询总数
	// countSql := fmt.Sprintf("SELECT COUNT(*) %s", fromClause)
	//if len(whereSql) > 0 {
	//	countSql = fmt.Sprintf("%s WHERE %s", countSql, whereSql)
	//}
	//
	//total, err = srCount.Raw(countSql, whereArgs...).Count()
	//if err != nil {
	//	err = errors.Wrapf(err, "[GetPubedList] count failed")
	//	return
	//}
	//
	//if total == 0 {
	//	return
	//}

	// 6. 执行查询
	err = sr.Raw(rawSql, whereArgs...).Find(&pos)
	if err != nil {
		err = errors.Wrapf(err, "[GetPubedList] find failed")
		return
	}

	// 7. 转换为PublishedJoinPo
	for _, releasePartPo := range pos {
		publishedAgentPo := &dapo.PublishedJoinPo{}

		err = publishedAgentPo.LoadFromReleasePartPo(&releasePartPo)
		if err != nil {
			return
		}

		rt = append(rt, publishedAgentPo)
	}

	return
}

func (repo *pubedAgentRepo) buildSql(req *pubedreq.PubedAgentListReq) (fromClause string, whereSql string, whereArgs []interface{}, err error) {
	// 1. 构建查询SQL
	agentCfgPO := &dapo.DataAgentPo{}
	releasePO := &dapo.ReleasePO{}

	// 3. 构建FROM和JOIN子句
	fromClause = fmt.Sprintf(
		"FROM %s AS cfg "+
			"INNER JOIN %s AS r ON cfg.f_id = r.f_agent_id ",
		agentCfgPO.TableName(),
		releasePO.TableName(),
	)

	// 4. 构建WHERE条件
	wb := sqlhelper2.NewWhereBuilder()

	// 5. 名称模糊查询
	if req.Name != "" {
		wb.Like("r.f_agent_name", req.Name)
	}

	// 6. 分类ID过滤
	if req.CategoryID != "" {
		categoryPO := &dapo.ReleaseCategoryRelPO{}
		fromClause += fmt.Sprintf(" INNER JOIN %s AS c ON r.f_id = c.f_release_id ", categoryPO.TableName())

		wb.WhereEqual("c.f_category_id", req.CategoryID)
	}

	// 7. 发布标识过滤
	if req.ToBeFlag != "" {
		switch req.ToBeFlag {
		case cdaenum.PublishToBeAPIAgent:
			wb.WhereEqual("r.f_is_api_agent", 1)
		case cdaenum.PublishToBeWebSDKAgent:
			wb.WhereEqual("r.f_is_web_sdk_agent", 1)
		case cdaenum.PublishToBeSkillAgent:
			wb.WhereEqual("r.f_is_skill_agent", 1)
		case cdaenum.PublishToBeDataFlowAgent:
			wb.WhereEqual("r.f_is_data_flow_agent", 1)
		}
	}

	// 8. 自定义空间ID过滤
	//if req.CustomSpaceID != "" {
	//	// 自定义空间资源表（SpaceResourcePo 已清理，直接使用表名）
	//	fromClause += fmt.Sprintf(" INNER JOIN %s AS sr ON cfg.f_id = sr.f_resource_id ", "t_custom_space_resource")
	//
	//	wb.WhereEqual("sr.f_space_id", req.CustomSpaceID)
	//	wb.WhereEqual("sr.f_resource_type", cdaenum.ResourceTypeDataAgent)
	//}

	// 9. ”发布到“过滤
	if req.IsToCustomSpace != 0 {
		wb.WhereEqual("r.f_is_to_custom_space", 1)
	}

	if req.IsToSquare != 0 {
		wb.WhereEqual("r.f_is_to_square", 1)
	}

	// 10. ID过滤
	if len(req.IDs) > 0 {
		wb.In("cfg.f_id", req.IDs)
	}

	// 11. 智能体标识过滤
	if len(req.AgentKeys) > 0 {
		wb.In("cfg.f_key", req.AgentKeys)
	}

	// 12. 排除智能体标识过滤
	if len(req.ExcludeAgentKeys) > 0 {
		wb.NotIn("cfg.f_key", req.ExcludeAgentKeys)
	}

	// 13. 删除标识过滤
	wb.WhereEqual("cfg.f_deleted_at", 0)

	// 14. 根据marker过滤
	if req.Marker != nil {
		err = repo.handleMarker(req, wb)
		if err != nil {
			return
		}
	}

	// 15. 构建WHERE SQL
	whereSql, whereArgs, err = wb.ToWhereSQL()
	if err != nil {
		err = errors.Wrapf(err, "[releaseRepo][buildSql] build where sql failed")
		return
	}

	return
}

// 根据marker过滤
// 参考：
//
//	select * from tb
//	where publish_at < marker.publish_at
//	   or (publish_at = marker.publish_at and id < marker.last_id)
//	order by publish_at desc, id desc
//	limit 2;
//
// 逻辑：
// a. 如果marker为空，表示从头开始
// b. 如果marker不为空，表示从marker开始
// 条件：
//
//	b1. publish_at < marker.publish_at
//	  or
//	b2. publish_at = marker.publish_at and id < marker.last_id (覆盖publish_at相同的情况)
func (repo *pubedAgentRepo) handleMarker(req *pubedreq.PubedAgentListReq, wb *sqlhelper2.WhereBuilder) (err error) {
	// 1 构建wb2
	pubedAt := req.Marker.PublishedAt
	lastID := req.Marker.LastReleaseID

	wb2 := sqlhelper2.NewWhereBuilder()
	wb2.WhereEqual("r.f_update_time", pubedAt)
	wb2.Where("r.f_id", sqlhelper2.OperatorLt, lastID)

	var (
		wb2Str  string
		wb2Args []interface{}
	)

	wb2Str, wb2Args, err = wb2.ToWhereSQL()
	if err != nil {
		return
	}

	// 2 构建wb3
	wb3 := sqlhelper2.NewWhereBuilder()

	wb3.Where("r.f_update_time", sqlhelper2.OperatorLt, pubedAt)
	wb3.WhereOrRaw(wb2Str, wb2Args...)

	var (
		wb3Str  string
		wb3Args []interface{}
	)

	wb3Str, wb3Args, err = wb3.ToWhereSQL()
	if err != nil {
		return
	}

	// 3. 添加到wb
	wb.WhereRaw(wb3Str, wb3Args...)

	return
}
