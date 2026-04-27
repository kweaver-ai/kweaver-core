package publishedtpldbacc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/dbhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/sqlhelper2"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// GetPubTplList 获取已发布模板列表
func (repo *PubedTplRepo) GetPubTplList(ctx context.Context, req *pubedreq.PubedTplListReq) (rt []*dapo.PublishedTplPo, err error) {
	// 1. 初始化一些变量等
	agentTplPo := &dapo.PublishedTplPo{}
	categoryAccPo := &dapo.PubTplCatAssocPo{}

	selectFieldsStr := sqlhelper2.GenSQLSelectFieldsStr(sqlhelper2.AllFieldsByStruct(agentTplPo), "pt")

	fromClause := fmt.Sprintf(
		"FROM %s AS pt ",
		agentTplPo.TableName(),
	)

	sr := dbhelper2.NewSQLRunner(repo.db, repo.logger)

	rt = make([]*dapo.PublishedTplPo, 0)

	publishedTplList := make([]dapo.PublishedTplPo, 0)
	po := &dapo.PublishedTplPo{}

	sr.FromPo(po)

	// 2. 构建WHERE条件
	wb := sqlhelper2.NewWhereBuilder()

	// 2.2 名称模糊查询
	if req.Name != "" {
		wb.Like("pt.f_name", req.Name)
	}

	// 2.3 分类ID过滤
	if req.CategoryID != "" {
		fromClause += fmt.Sprintf("INNER JOIN %s AS rel ON pt.f_id = rel.f_published_tpl_id", categoryAccPo.TableName())

		wb.WhereEqual("rel.f_category_id", req.CategoryID)
	}

	// 2.4 按业务域进行过滤
	if len(req.TplIDsByBd) > 0 {
		wb.In("pt.f_tpl_id", req.TplIDsByBd)
	}

	// 2.5 根据marker过滤
	if req.Marker != nil && req.Marker.LastPubedTplID > 0 {
		wb.Where("pt.f_id", sqlhelper2.OperatorLt, req.Marker.LastPubedTplID)
	}

	// 2.6 构建WHERE子句
	whereSql, whereArgs, err := wb.ToWhereSQL()
	if err != nil {
		err = errors.Wrapf(err, "[PubedTplRepo][GetPubTplList] build where sql failed")
		return
	}

	var whereClause string
	if len(whereSql) > 0 {
		whereClause = fmt.Sprintf(" WHERE %s", whereSql)
	}

	// 3. 查询列表
	// 3.1 构建SELECT SQL
	rawSql := fmt.Sprintf("SELECT %s %s", selectFieldsStr, fromClause)

	if len(whereClause) > 0 {
		rawSql = fmt.Sprintf("%s %s", rawSql, whereClause)
	}

	// 3.2 添加排序
	rawSql = fmt.Sprintf("%s ORDER BY pt.f_id DESC", rawSql)

	// 3.3 添加分页
	rawSql = fmt.Sprintf("%s LIMIT %d ", rawSql, req.Size)

	// 3.4 执行查询
	err = sr.Raw(rawSql, whereArgs...).Find(&publishedTplList)
	if err != nil {
		err = errors.Wrapf(err, "[PubedTplRepo][GetPubTplList] find failed")
		return
	}

	// 4. 转换为指针切片
	rt = cutil.SliceToPtrSlice(publishedTplList)

	return
}
