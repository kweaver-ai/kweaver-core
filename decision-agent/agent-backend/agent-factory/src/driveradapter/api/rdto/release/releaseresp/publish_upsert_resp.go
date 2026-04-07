package releaseresp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
)

// PublishUpsertResp 发布或更新发布信息的响应
type PublishUpsertResp struct {
	ReleaseId       string `json:"release_id"`
	Version         string `json:"version"`
	PublishedAt     int64  `json:"published_at"`
	PublishedBy     string `json:"published_by"`
	PublishedByName string `json:"published_by_name"`
}

func (r *PublishUpsertResp) FillPublishedByName(ctx context.Context, um iumacc.UmHttpAcc) (err error) {
	if cenvhelper.IsLocalDev() {
		r.PublishedByName = r.PublishedBy + "_name"
		return
	}

	name, err := um.GetSingleUserName(ctx, r.PublishedBy)
	if err != nil {
		return
	}

	r.PublishedByName = name

	return
}
