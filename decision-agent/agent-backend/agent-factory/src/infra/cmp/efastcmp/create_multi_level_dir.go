package efastcmp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/efastcmp/dto/efastret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (e *EFast) CreateMultiLevelDir(ctx context.Context, req *efastarg.CreateMultiLevelDirReq, token string) (ret *efastret.CreateMultiLevelDirRsp, err error) {
	url := fmt.Sprintf("%s/v1/dir/createmultileveldir", e.getPublicUrlPrefix())

	// 2、调用接口
	opt := httphelper.WithToken(token)
	c := httphelper.NewHTTPClient(opt)

	resp, err := c.PostJSONExpect2xx(ctx, url, req)
	if err != nil {
		return
	}

	// 3、处理返回结果
	ret = &efastret.CreateMultiLevelDirRsp{}

	err = cutil.JSON().Unmarshal([]byte(resp), &ret)
	if err != nil {
		return
	}

	return
}
