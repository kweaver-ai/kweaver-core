package icmp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/eftypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

//go:generate mockgen -package cmpmock -source efast.go -destination ./cmpmock/efast_mock.go
type IEFast interface {
	GetFileDocLibType(ctx context.Context, ids []string) (m map[string]cenum.DocLibType, err error)

	CheckObjExists(ctx context.Context, ids []string) (notExistsIDs []string, err error)

	GetFsMetadata(ctx context.Context, args *efastarg.GetFsMetadataArgDto) (ret efastret.GetFsMetadataRetDto, err error)

	GetInfoByPath(ctx context.Context, path, token string) (isNotExists bool, ret *eftypes.Path2GnsResponse, err error)

	Path2Gns(ctx context.Context, path, token string) (isNotExists bool, gns string, err error)

	CreateMultiLevelDir(ctx context.Context, req *efastarg.CreateMultiLevelDirReq, token string) (ret *efastret.CreateMultiLevelDirRsp, err error)

	GetFsMetadataMap(ctx context.Context, docLibIds []string, fields []efastarg.IbField) (m map[string]*efastret.FsMetadata, err error)

	GetOneFsName(ctx context.Context, docLibId string) (name string, err error)

	GetFsMetadataNameMap(ctx context.Context, docIds []string) (m map[string]string, err error)
}
