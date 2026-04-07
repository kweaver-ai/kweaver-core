package productsvc

import (
	"context"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/e2p/producte2p"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (s *productSvc) Update(ctx context.Context, req *productreq.UpdateReq, id int64) (auditloginfo auditlogdto.ProductUpdateAuditLogInfo, err error) {
	auditloginfo = auditlogdto.ProductUpdateAuditLogInfo{}

	// 1. 检查是否存在
	_, err = s.productRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.ProductNotFound, "产品不存在")
			return
		}

		return
	}

	auditloginfo = auditlogdto.ProductUpdateAuditLogInfo{
		ID:   strconv.FormatInt(id, 10),
		Name: req.Name,
	}

	// 2. 检查名称是否重复（如果名称有变化）
	if req.Name != "" {
		var existsByName bool

		existsByName, err = s.productRepo.ExistsByNameExcludeID(ctx, req.Name, id)
		if err != nil {
			return auditloginfo, err
		}

		if existsByName {
			err = capierr.NewCustom409Err(ctx, apierr.ProductNameExists, "产品名称已存在")
			return auditloginfo, err
		}
	}

	// 3. DTO 转 EO
	eo, err := req.D2e()
	if err != nil {
		return auditloginfo, err
	}

	eo.ID = id

	// 4. 设置更新信息
	eo.UpdatedAt = cutil.GetCurrentMSTimestamp()
	eo.UpdatedBy = chelper.GetUserIDFromCtx(ctx)

	// 5. EO转PO
	po, err := producte2p.Product(eo)
	if err != nil {
		return
	}

	// 6. 调用repo层更新数据
	err = s.productRepo.Update(ctx, po)
	if err != nil {
		return
	}

	return
}
