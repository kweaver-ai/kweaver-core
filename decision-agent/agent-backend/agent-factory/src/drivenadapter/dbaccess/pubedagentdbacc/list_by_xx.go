package pubedagentdbacc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/sqlhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *pubedAgentRepo) GetPubedListByXx(ctx context.Context, arg *padbarg.GetPaPoListByXxArg) (ret *padbret.GetPaPoListByXxRet, err error) {
	ret = padbret.NewGetPaPoListByXxRet()

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	pos := make([]dapo.ReleasePartPo, 0)

	if len(arg.AgentKeys) == 0 && len(arg.AgentIDs) == 0 {
		return
	}

	// 1. 构建select字段
	agentCfgPO := &dapo.DataAgentPo{}
	releasePO := &dapo.ReleasePO{}

	sqlFromPart := fmt.Sprintf(
		" FROM %s AS cfg INNER JOIN %s AS r ON cfg.f_id = r.f_agent_id",
		agentCfgPO.TableName(),
		releasePO.TableName(),
	)

	// 2. 构建where条件
	wb := sqlhelper2.NewWhereBuilder()

	wb.WhereEqual("cfg.f_deleted_at", 0)

	if len(arg.AgentKeys) > 0 {
		wb.In("cfg.f_key", arg.AgentKeys)
	}

	if len(arg.AgentIDs) > 0 {
		wb.In("cfg.f_id", arg.AgentIDs)
	}

	if arg.PubToWhereCond != nil {
		if arg.PubToWhereCond.IsToCustomSpace {
			wb.WhereEqual("r.f_is_to_custom_space", 1)
		}

		if arg.PubToWhereCond.IsToSquare {
			wb.WhereEqual("r.f_is_to_square", 1)
		}
	}

	whereSql, whereArgs, err := wb.ToWhereSQL()
	if err != nil {
		err = errors.Wrapf(err, "[GetPubedListByXx] build where sql failed")
		return
	}

	// 3. 构建rawSql
	releasePartPO := &dapo.ReleasePartPo{}

	selectFields := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(releasePartPO), "r")

	rawSql := fmt.Sprintf("SELECT %s %s ", selectFields, sqlFromPart)

	if len(whereSql) > 0 {
		rawSql = fmt.Sprintf("%s WHERE %s", rawSql, whereSql)
	}

	// 4. 执行查询
	err = sr.Raw(rawSql, whereArgs...).Find(&pos)
	if err != nil {
		err = errors.Wrapf(err, "[GetPubedListByXx] find failed")
		return
	}

	// 5. 转换为指针切片

	for _, po := range pos {
		publishedAgentPo := &dapo.PublishedJoinPo{}

		err = publishedAgentPo.LoadFromReleasePartPo(&po)
		if err != nil {
			return
		}

		ret.JoinPos = append(ret.JoinPos, publishedAgentPo)
	}

	return
}
