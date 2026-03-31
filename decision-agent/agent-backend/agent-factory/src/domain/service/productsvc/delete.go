package productsvc

import (
	"context"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
)

func (s *productSvc) Delete(ctx context.Context, id int64) (auditloginfo auditlogdto.ProductDeleteAuditLogInfo, err error) {
	// 2. 检查是否存在
	exists, err := s.productRepo.ExistsByID(ctx, id)
	if err != nil {
		return
	}

	if !exists {
		err = capierr.NewCustom404Err(ctx, apierr.ProductNotFound, "产品不存在")
		return
	}

	po, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return
	}

	auditloginfo = auditlogdto.ProductDeleteAuditLogInfo{
		ID:   strconv.FormatInt(po.ID, 10),
		Name: po.Name,
	}
	// 注释掉审计日志相关代码
	// 2. 获取原始数据（如果需要发送审计日志）
	// origPo, err := s.productRepo.GetByID2(ctx, id)
	// if err != nil {
	// 	return
	// }
	//
	// 3. PO转EO（如果需要发送审计日志）
	// origEo, err := daconfp2e.DataAgent(origPo)
	// if err != nil {
	// 	return
	// }

	// 4. 调用repo层删除数据（软删除）
	err = s.productRepo.Delete(ctx, id)
	if err != nil {
		return
	}

	// 5. 发送审计日志
	// err = s.sendAuditLog(ctx, origEo, persrecenums.MngLogOpTypeDelete, tx)

	return
}
