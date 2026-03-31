package productsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/productp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productresp"
)

func (s *productSvc) List(ctx context.Context, offset, limit int) (res *productresp.ListRes, err error) {
	// 1. 获取数据
	pos, total, err := s.productRepo.List(ctx, offset, limit)
	if err != nil {
		return
	}

	// 2. 转换为响应DTO
	res = productresp.NewListRes()
	res.Total = total

	if len(pos) == 0 {
		return
	}

	// 3. PO转EO，再转DTO

	eos, err := productp2e.Products(ctx, pos)
	if err != nil {
		return
	}

	err = res.LoadFromEo(eos)
	if err != nil {
		return
	}

	return
}
