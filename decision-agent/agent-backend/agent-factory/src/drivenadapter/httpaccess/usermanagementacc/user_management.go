// Package usermanagementacc 身份校验
package usermanagementacc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iusermanagementacc"
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
	usersInfoURI       = "/api/user-management/v1/users/%s/%s"
	userIDByAccountURI = "/api/user-management/v1/account-match"
)

// GetUserInfoByUserID 通过用户id获取用户信息
func (cli *client) GetUserInfoByUserID(ctx context.Context, userIDs []string, fields []string) (usersInfo map[string]*iusermanagementacc.UserInfo, err error) {
	if cenvhelper.IsLocalDev() {
		return map[string]*iusermanagementacc.UserInfo{
			"1": {
				Name:  "admin",
				ID:    "1",
				Roles: []string{"super_admin"},
			},
		}, nil
	}

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

type user struct {
	ID            string `json:"id"`
	Account       string `json:"account"`
	DisableStatus bool   `json:"disable_status"`
}

type getUserIDByAccountRsp struct {
	Result bool `json:"result"`
	User   user `json:"user"`
}

func (cli *client) GetUserIDByAccount(ctx context.Context, account string) (invalid bool, userID string, err error) {
	vals := url.Values{}
	vals.Add("account", account)
	uri := fmt.Sprintf("%s%s?%s", cli.address, userIDByAccountURI, vals.Encode())
	data, err := cli.httpClient.GetExpect2xxByte(ctx, uri, nil)
	if err != nil {
		cli.log.Errorf("[GetUserIDByAccount] request failed:%v, url:%s", err, uri)
		err = errors.Wrapf(err, "request failed")

		return
	}

	rsp := &getUserIDByAccountRsp{}

	err = jsoniter.Unmarshal(data, rsp)
	if err != nil {
		cli.log.Errorf("[GetUserIDByAccount] Unmarshal failed:%v, res:%s", err, string(data))
		err = errors.Wrapf(err, "Unmarshal failed")

		return
	}

	invalid = rsp.Result && !rsp.User.DisableStatus
	userID = rsp.User.ID

	return
}

func (cli *client) getUserRoles(ctx context.Context, userID string) (userRoles []string, name string, err error) {
	var usersInfos map[string]*iusermanagementacc.UserInfo

	usersInfos, err = cli.GetUserInfoByUserID(ctx, []string{userID}, []string{"name", "roles"})
	if err != nil {
		return
	}

	userInfo, ok := usersInfos[userID]
	if !ok {
		cli.log.Errorf("user %s not found", userID)
		err = fmt.Errorf("user %s not found", userID)

		return
	}

	userRoles = userInfo.Roles
	name = userInfo.Name

	return
}

func (cli *client) IsSuperAdmin(ctx context.Context, userID string) (isSuperAdmin bool, err error) {
	if cenvhelper.IsLocalDev() {
		return true, nil
	}

	userRoles, _, err := cli.getUserRoles(ctx, userID)
	if err != nil {
		return
	}

	isSuperAdmin = slices.Contains(userRoles, "super_admin")

	return
}
