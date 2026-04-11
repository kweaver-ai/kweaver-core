package idocsethttp

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/docsetaccess/docsetdto"
)

type IDocset interface {
	FullText(ctx context.Context, req *docsetdto.FullTextReq) (rsp *docsetdto.FullTextRsp, err error)
}
