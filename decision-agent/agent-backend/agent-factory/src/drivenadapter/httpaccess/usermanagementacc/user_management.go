// Package usermanagementacc 身份校验
package usermanagementacc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/pkg/errors"
)

type client struct {
	address    string
	httpClient icmp.IHttpClient
	log        icmp.Logger
}

// NewClient 获取用户管理客户端
func NewClient(conf ...httphelper.Option) iusermanagementacc.UserMgnt {
	return &client{
		address:    cutil.GetHTTPAccess(global.GConfig.Hydra.UserMgnt.Host, global.GConfig.Hydra.UserMgnt.Port, "http"),
		httpClient: httphelper.NewHTTPClient(conf...),
		log:        logger.GetLogger(),
	}
}

var (
	usersInfoURI = "/api/user-management/v1/users/%s/%s"
)

// GetUserInfoByUserID 通过用户id获取用户信息
func (cli *client) GetUserInfoByUserID(ctx context.Context, userIDs []string, fields []string) (usersInfo map[string]*iusermanagementacc.UserInfo, err error) {
	uri := cli.address + fmt.Sprintf(usersInfoURI, strings.Join(userIDs, ","), strings.Join(fields, ","))
	data, err := cli.httpClient.GetExpect2xxByte(ctx, uri, nil)
	if err != nil {
		cli.log.Errorf("[GetUserInfoByUserID] request failed:%v, url:%s", err, uri)
		err = errors.Wrapf(err, "request failed")

		return
	}

	rsp := []*iusermanagementacc.UserInfo{}

	err = json.Unmarshal(data, &rsp)
	if err != nil {
		cli.log.Errorf("[GetUserInfoByUserID] Unmarshal failed:%v, res:%v", err, data)
		err = errors.Wrapf(err, "Unmarshal failed")

		return
	}

	usersInfo = map[string]*iusermanagementacc.UserInfo{}
	for _, info := range rsp {
		usersInfo[info.ID] = info
	}

	return
}
