package releaseacc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/sqlhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// List implements release.ReleaseRepo.
func (repo *releaseHistoryRepo) ListByAgentID(ctx context.Context, agentID string) (rt []*dapo.ReleaseHistoryPO, total int64, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	srCount := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	po := &dapo.ReleaseHistoryPO{}
	historyPOList := make([]dapo.ReleaseHistoryPO, 0)

	sr.FromPo(po).WhereEqual("f_agent_id", agentID).Order("f_agent_version DESC")

	err = sr.Find(&historyPOList)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			return nil, 0, nil
		}

		return nil, 0, errors.Wrapf(err, "get release by agent id %s", agentID)
	}

	rt = cutil.SliceToPtrSlice(historyPOList)

	srCount.FromPo(po).WhereEqual("f_agent_id", agentID)

	total, err = srCount.Count()
	if err != nil {
		return nil, 0, errors.Wrapf(err, "get release count")
	}

	return
}

// 根据 publish_to_be 构造 where sql

// 查询最近访问的 Agent 列表
// 由于发布的 agent 需要通过 release 表获取，未发布的 agent 需要通过 config表获取，同时要求如果用户访问了同一个agent的发布和未发布的版本，需要同时返回两个版本在前端显示
// 此处还要支持分页，因此当前的实现方案是，基于传入的最大返回数量，分别查询出未发布和已发布的 agent，然后合并后返回，在 service 层基于访问时间做排序和分页
func (repo *releaseRepo) ListRecentAgentForMarket(ctx context.Context, req squarereq.AgentSquareRecentAgentReq) (rt []*dapo.RecentVisitAgentPO, err error) {
	unpublishedAgentPOList, err := repo.listRecentUnpublishedAgent(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "svc.releaseRepo.listRecentUnpublishedAgent")
	}

	rt = append(rt, unpublishedAgentPOList...)

	publishedAgentPOList, err := repo.listRecentPublishedAgent(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "svc.releaseRepo.listRecentPublishedAgent")
	}

	rt = append(rt, publishedAgentPOList...)

	return
}

// 查询最近访问的未发布 Agent
// 未发布 Agent 为 v0 版本
func (repo *releaseRepo) listRecentUnpublishedAgent(ctx context.Context, req squarereq.AgentSquareRecentAgentReq) (rt []*dapo.RecentVisitAgentPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	unpublishedAgentPOList := make([]dapo.RecentVisitAgentPO, 0)

	agentCfgPO := &dapo.DataAgentPo{}
	visitHistoryPO := &dapo.VisitHistoryPO{}
	sqlselectCfgPart := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(agentCfgPO), "cfg")

	sqlFromPart := fmt.Sprintf(
		" FROM %s AS cfg INNER JOIN %s AS v ON cfg.f_id = v.f_agent_id AND v.f_agent_version='v0'",
		agentCfgPO.TableName(),
		visitHistoryPO.TableName(),
	)
	wb := sqlhelper2.NewWhereBuilder()

	wb.WhereEqual("cfg.f_deleted_at", 0).WhereEqual("v.f_create_by", req.UserID).Where("v.f_update_time", sqlhelper2.OperatorGte, req.StartTime).Where("v.f_update_time", sqlhelper2.OperatorLte, req.EndTime)

	whereSql, whereArgs, err := wb.ToWhereSQL()
	if err != nil {
		return
	}

	toBe := dapo.PublishedToBeStruct{}

	rawSql := fmt.Sprintf("SELECT %s, '' AS f_agent_config, '' AS f_agent_desc, 'v0' AS f_agent_version, 0 AS publish_time, '' AS publish_user_id, v.f_update_time AS last_visit_time, %s %s ", sqlselectCfgPart, toBe.SelectFieldsZero(), sqlFromPart)

	if len(whereSql) > 0 {
		rawSql = fmt.Sprintf("%s WHERE %s", rawSql, whereSql)
	}

	rawSql = fmt.Sprintf("%s ORDER BY v.f_update_time DESC", rawSql)
	rawSql = fmt.Sprintf("%s LIMIT %d OFFSET %d", rawSql, req.Size, 0)

	err = sr.Raw(rawSql, whereArgs...).Find(&unpublishedAgentPOList)
	if err != nil {
		return nil, errors.Wrapf(err, "find release agent")
	}

	rt = cutil.SliceToPtrSlice(unpublishedAgentPOList)

	return
}

// 查询最近访问的已发布 Agent
func (repo *releaseRepo) listRecentPublishedAgent(ctx context.Context, req squarereq.AgentSquareRecentAgentReq) (rt []*dapo.RecentVisitAgentPO, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	publishedAgentPOList := make([]dapo.RecentVisitAgentPO, 0)

	agentCfgPO := &dapo.DataAgentPo{}
	releasePO := &dapo.ReleasePO{}
	visitHistoryPO := &dapo.VisitHistoryPO{}

	sqlselectCfgPart := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(agentCfgPO), "cfg")

	pubedToBePo := dapo.PublishedToBeStruct{}
	sqlSelectPubedToBePart := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(pubedToBePo), "r")

	sqlFromPart := fmt.Sprintf(
		" FROM %s AS cfg INNER JOIN %s AS r ON cfg.f_id = r.f_agent_id INNER JOIN %s AS v ON r.f_agent_id = v.f_agent_id AND v.f_agent_version!='v0'",
		agentCfgPO.TableName(),
		releasePO.TableName(),
		visitHistoryPO.TableName(),
	)
	wb := sqlhelper2.NewWhereBuilder()

	wb.WhereEqual("v.f_create_by", req.UserID).Where("v.f_update_time", sqlhelper2.OperatorGte, req.StartTime).Where("v.f_update_time", sqlhelper2.OperatorLte, req.EndTime)

	whereSql, whereArgs, err := wb.ToWhereSQL()
	if err != nil {
		return
	}

	rawSql := fmt.Sprintf("SELECT %s, r.f_agent_config, r.f_agent_desc, r.f_agent_version, r.f_update_time AS publish_time, r.f_update_by AS publish_user_id, v.f_update_time AS last_visit_time,%s %s ", sqlselectCfgPart, sqlSelectPubedToBePart, sqlFromPart)

	if len(whereSql) > 0 {
		rawSql = fmt.Sprintf("%s WHERE %s", rawSql, whereSql)
	}

	rawSql = fmt.Sprintf("%s ORDER BY v.f_update_time DESC", rawSql)
	rawSql = fmt.Sprintf("%s LIMIT %d OFFSET %d", rawSql, req.Size, 0)

	err = sr.Raw(rawSql, whereArgs...).Find(&publishedAgentPOList)
	if err != nil {
		return nil, errors.Wrapf(err, "find release agent")
	}

	rt = cutil.SliceToPtrSlice(publishedAgentPOList)

	return
}

func (repo *releaseRepo) GetMapByAgentIDs(ctx context.Context, agentIDs []string) (m map[string]*dapo.ReleasePO, err error) {
	pos := make([]dapo.ReleasePO, 0)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(&dapo.ReleasePO{})

	err = sr.In("f_agent_id", agentIDs).Find(&pos)
	if err != nil {
		err = errors.Wrap(err, "[releaseRepo][GetMapByAgentIDs] err")
		return
	}

	m = make(map[string]*dapo.ReleasePO)
	for _, po := range pos {
		m[po.ID] = &po
	}

	return
}

func (repo *releaseRepo) GetMapByUniqFlags(ctx context.Context, uniqFlags []*comvalobj.DataAgentUniqFlag) (m map[string]*dapo.ReleasePO, err error) {
	pos := make([]dapo.ReleasePO, 0)

	if len(uniqFlags) == 0 {
		return
	}

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(&dapo.ReleasePO{})

	// 1. 构造where条件
	wb := sqlhelper2.NewWhereBuilder()

	for _, item := range uniqFlags {
		wbTmp := sqlhelper2.NewWhereBuilder()
		wbTmp.WhereEqual("f_agent_id", item.AgentID)
		wbTmp.WhereEqual("f_agent_version", item.AgentVersion)

		var (
			whereSql  string
			whereArgs []interface{}
		)

		whereSql, whereArgs, err = wbTmp.ToWhereSQL()
		if err != nil {
			err = errors.Wrap(err, "[releaseRepo][GetMapByUniqFlags] ToWhereSQL err")
			return
		}

		wb.WhereOrRaw(whereSql, whereArgs...)
	}

	// 2. 执行查询
	err = sr.WhereByWhereBuilder(wb)
	if err != nil {
		err = errors.Wrap(err, "[releaseRepo][GetMapByUniqFlags] WhereByWhereBuilder err")
		return
	}

	err = sr.Find(&pos)
	if err != nil {
		err = errors.Wrap(err, "[releaseRepo][GetMapByUniqFlags] Find err")
		return
	}

	// 3. 构造返回值
	m = make(map[string]*dapo.ReleasePO)
	for _, po := range pos {
		m[po.ID] = &po
	}

	return
}
