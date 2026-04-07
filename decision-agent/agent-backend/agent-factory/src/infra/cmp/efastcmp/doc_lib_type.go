package efastcmp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// GetFileDocLibType 获取文件所属文档库类型
func (e *EFast) GetFileDocLibType(ctx context.Context, ids []string) (m map[string]cenum.DocLibType, err error) {
	m = make(map[string]cenum.DocLibType)

	dto := &efastarg.GetFsMetadataArgDto{
		IDs: ids,
		Fields: []efastarg.IbField{
			efastarg.IbFieldDocLibTypes,
		},
	}

	var ret efastret.GetFsMetadataRetDto

	ret, err = e.GetFsMetadata(ctx, dto)
	if err != nil {
		return
	}

	for _, v := range ret {
		m[v.ID] = v.DocLibType
	}

	return
}
