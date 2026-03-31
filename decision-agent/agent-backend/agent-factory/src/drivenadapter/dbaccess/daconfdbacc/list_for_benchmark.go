package daconfdbacc

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *DAConfigRepo) ListForBenchmark2(ctx context.Context, req *agentconfigreq.ListForBenchmarkReq) (pos []*dapo.ListForBenchmarkPo, count int64, err error) {
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	// 1. 构建基础查询条件
	var nameFilter string

	var nameFilter2 string

	if req.Name != "" {
		nameFilter = fmt.Sprintf("AND f_name LIKE '%%%s%%'", req.Name)
		nameFilter2 = fmt.Sprintf("AND r.f_agent_name LIKE '%%%s%%'", req.Name)
	}

	// 2. 先执行count查询
	countSQL := fmt.Sprintf(`
	SELECT COUNT(*) FROM (
		-- 查询未删除的 agent 配置信息，状态为未发布
		SELECT 
			f_id as f_id,
			f_name as f_name
		FROM 
			t_data_agent_config
		WHERE 
			f_deleted_at = 0 %s

		UNION ALL

		-- 查询已发布的 agent 信息，关联配置信息，状态为已发布
		SELECT 
			r.f_agent_id as f_id,
			r.f_agent_name as f_name
		FROM 
			t_data_agent_release r
			INNER JOIN t_data_agent_config c 
				ON r.f_agent_id = c.f_id
		where c.f_deleted_at=0 %s
	) AS total_count
	`, nameFilter, nameFilter2)

	count, err = sr.Raw(countSQL).Count()
	if err != nil {
		err = errors.Wrapf(err, "get agent count")
		return
	}

	if count == 0 {
		return
	}

	// 3. 再执行分页查询
	poList := make([]dapo.ListForBenchmarkPo, 0)

	rawSQL := fmt.Sprintf(`
	(
		-- 查询未删除的 agent 配置信息，状态为未发布
		SELECT 
			f_id as f_id,
			f_key as f_key,
			f_name as f_name,
			'v0' AS f_version,
			'unpublished' AS f_status,
			f_updated_at AS updated_at
		FROM 
			t_data_agent_config
		WHERE 
			f_deleted_at = 0 %s

		UNION ALL

		-- 查询已发布的 agent 信息，关联配置信息，状态为已发布
		SELECT 
			r.f_agent_id as f_id,
			c.f_key as f_key,
			r.f_agent_name as f_name,
			'latest' AS f_version,
			'published' AS f_status,
			r.f_update_time AS updated_at
		FROM 
			t_data_agent_release r
			INNER JOIN t_data_agent_config c 
				ON r.f_agent_id = c.f_id
		where c.f_deleted_at=0 %s
	)
	ORDER BY 
		updated_at DESC
	LIMIT %d OFFSET %d;
	`, nameFilter, nameFilter2, req.GetSize(), req.GetOffset())

	err = sr.Raw(rawSQL).Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get agent list")
		return
	}

	// 4. 转换为指针切片
	pos = cutil.SliceToPtrSlice(poList)

	return
}

func (repo *DAConfigRepo) ListForBenchmark(ctx context.Context, req *agentconfigreq.ListForBenchmarkReq) (pos []*dapo.ListForBenchmarkPo, count int64, err error) {
	srCount := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	pos = make([]*dapo.ListForBenchmarkPo, 0)
	poList := make([]dapo.ListForBenchmarkMerge, 0)

	// 1. 先执行count查询
	// 1.1 构建基础查询条件
	whereArgs := make([]interface{}, 0)

	nameFilter := ""
	if req.Name != "" {
		nameFilter = "AND c.f_name LIKE ?"

		whereArgs = append(whereArgs, "%"+req.Name+"%")
	}

	// 1.1.1 构建业务域过滤条件
	agentIDsFilter := ""

	if len(req.AgentIDsByBizDomain) > 0 {
		placeholders := make([]string, 0, len(req.AgentIDsByBizDomain))
		for _, id := range req.AgentIDsByBizDomain {
			placeholders = append(placeholders, "?")
			whereArgs = append(whereArgs, id)
		}

		agentIDsFilter = fmt.Sprintf("AND c.f_id IN (%s)", strings.Join(placeholders, ","))
	}

	// 1.2 构建count查询SQL
	countSQL := fmt.Sprintf(`
	select count(*) 
from t_data_agent_config c
where c.f_deleted_at = 0 %s %s
	`, nameFilter, agentIDsFilter)

	// 1.3 执行count查询
	count, err = srCount.Raw(countSQL, whereArgs...).Count()
	if err != nil {
		err = errors.Wrapf(err, "get agent count")
		return
	}
	// 1.4 如果count为0，直接返回
	if count == 0 {
		return
	}

	// 2. 再执行分页查询
	rawSQL := fmt.Sprintf(`
	select c.f_id,c.f_key, c.f_name, c.f_updated_at,
       'v0'          AS f_version,
       f_status,

       r.f_agent_id,c.f_key as f_agent_key, r.f_agent_name, r.f_update_time as f_published_at,
       'latest'        AS f_release_version,
       'published'     AS f_release_status
from t_data_agent_config c
         left join t_data_agent_release r on c.f_id = r.f_agent_id
where c.f_deleted_at = 0 %s %s
order by c.f_updated_at desc
limit %d offset %d;
	`, nameFilter, agentIDsFilter, req.GetSize(), req.GetOffset())

	err = sr.Raw(rawSQL, whereArgs...).Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get agent list")
		return
	}

	// 3. 将查询结果转换为ListForBenchmarkPo
	for _, v := range poList {
		// 3.1 处理agent config po
		tmp := &dapo.ListForBenchmarkPo{}

		err = cutil.CopyStructUseJSON(tmp, v.ListForBenchmarkAgentPo)
		if err != nil {
			return
		}

		// 3.2 如果是未发布状态，查看是否有对应的已发布版本(release po)
		// 如果有，则添加已发布版本到pos
		if !tmp.Status.IsPublished() {
			if v.ListForBenchmarkReleasePo.ID.Valid {
				pos = append(pos, v.ListForBenchmarkReleasePo.ToListForBenchmarkPo())
			}
		} else {
			tmp.Version = daconstant.AgentVersionLatest
		}

		// 3.3 添加agent config po（可能是未发布状态，也可能是已发布状态）
		pos = append(pos, tmp)
	}

	return
}
