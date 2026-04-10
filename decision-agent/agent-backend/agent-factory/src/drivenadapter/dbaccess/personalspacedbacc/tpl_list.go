package personalspacedbacc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/personalspacedbacc/psdbarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/sqlhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (repo *personalSpaceRepo) ListPersonalSpaceTpl(ctx context.Context, arg *psdbarg.TplListArg) (pos []*dapo.DataAgentTplPo, err error) {
	req := arg.ListReq

	po := &dapo.DataAgentTplPo{}

	// 1. 构建查询条件
	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)
	sr.FromPo(po).
		WhereEqual("f_deleted_at", 0)

	// 1.1 按名称模糊搜索
	if req.Name != "" {
		sr.Like("f_name", req.Name)
	}

	// 1.2 按产品标识过滤
	if req.ProductKey != "" {
		sr.WhereEqual("f_product_key", req.ProductKey)
	}

	// 1.3 按状态过滤
	if req.PublishStatus != "" {
		sr.WhereEqual("f_status", req.PublishStatus)
	}

	// 1.4 按创建人过滤（用于个人空间）
	sr.WhereEqual("f_created_by", arg.CreatedBy)

	// 1.5 按模板创建类型过滤
	if req.AgentTplCreatedType != "" {
		sr.WhereEqual("f_created_type", req.AgentTplCreatedType)
	}

	// 1.6 按业务域ID过滤
	if len(arg.TplIDsByBd) > 0 {
		sr.In("f_id", arg.TplIDsByBd)
	}

	// 1.7 按更新时间过滤
	if req.Marker != nil {
		if err = repo.handleTplMarker(arg, sr); err != nil {
			return
		}
	}

	// 2. 获取列表数据
	poList := make([]dapo.DataAgentTplPo, 0)

	// 2.1 按更新时间倒序排列
	sr.ResetSelect()
	sr.Order("f_updated_at DESC,f_id DESC")
	sr.Limit(req.Size)

	// 2.2 执行查询
	err = sr.Find(&poList)
	if err != nil {
		err = errors.Wrapf(err, "get agent template list")
		return
	}

	// 2.3 转换为指针切片
	pos = cutil.SliceToPtrSlice(poList)

	return
}

// 根据marker过滤
func (repo *personalSpaceRepo) handleTplMarker(req *psdbarg.TplListArg, sr *dbhelper2.SQLRunner) (err error) {
	reqMarker := req.ListReq.Marker
	// 1. 构建wb2
	wb2 := sqlhelper2.NewWhereBuilder()
	wb2.WhereEqual("f_updated_at", reqMarker.UpdatedAt)
	wb2.Where("f_id", sqlhelper2.OperatorLt, reqMarker.LastTplID)

	var (
		wb2Str  string
		wb2Args []interface{}
	)

	wb2Str, wb2Args, err = wb2.ToWhereSQL()
	if err != nil {
		return
	}

	// 2. 构建wb3
	wb3 := sqlhelper2.NewWhereBuilder()

	wb3.Where("f_updated_at", sqlhelper2.OperatorLt, reqMarker.UpdatedAt)
	wb3.WhereOrRaw(wb2Str, wb2Args...)

	var (
		wb3Str  string
		wb3Args []interface{}
	)

	wb3Str, wb3Args, err = wb3.ToWhereSQL()
	if err != nil {
		return
	}

	// 3. 添加到sr
	sr.WhereRaw(wb3Str, wb3Args...)

	return
}
