package drivenadapters

import (
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/authentication.go -destination ../tests/mock_drivenadapters/authentication_mock.go

// Authentication method interface
type Authentication interface {
	// ConfigAuthPerm 配置应用账户获取任意用户访问令牌的权限
	ConfigAuthPerm(appID string) error

	// GetAssertion 获取断言
	GetAssertion(userID, token string) (assertion string, err error)
}

type auth struct {
	publicAddress  string
	privateAddress string
	log            commonLog.Logger
	httpClient     HTTPClient
}

var (
	auOnce sync.Once
	au     Authentication
)

// NewAuthentication 创建获取用户服务
func NewAuthentication() Authentication {
	auOnce.Do(func() {
		config := common.NewConfig()
		if !config.IsAuthEnabled() {
			au = &mockAuthentication{}
		} else {
			au = &auth{
				publicAddress:  fmt.Sprintf("http://%s:%v", config.Authentication.PublicHost, config.Authentication.PublicPort),
				privateAddress: fmt.Sprintf("http://%s:%v", config.Authentication.PrivateHost, config.Authentication.PrivatePort),
				log:            commonLog.NewLogger(),
				httpClient:     NewHTTPClient(),
			}
		}
	})
	return au
}

// ConfigAuthPerm 配置应用账户获取任意用户访问令牌的权限
func (au *auth) ConfigAuthPerm(appID string) error {
	target := fmt.Sprintf("%s/api/authentication/v1/access-token-perm/app/%s",
		au.privateAddress, appID)
	_, _, err := au.httpClient.Put(target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, map[string]string{})
	if err != nil {
		au.log.Errorf("ConfigAuthPerm failed: %v, url: %v", err, target)
		return err
	}

	return nil
}

// GetAssertion 获取断言
func (au *auth) GetAssertion(userID, token string) (assertion string, err error) {
	target := fmt.Sprintf("%s/api/authentication/v1/jwt?user_id=%s",
		au.publicAddress, userID)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8", "Authorization": fmt.Sprintf("Bearer %s", token)}
	resp, err := au.httpClient.Get(target, headers)
	if err != nil {
		au.log.Errorf("GetAssertion failed: %v, url: %v", err, target)
		return
	}

	if res, ok := resp.(map[string]interface{}); ok {
		if _assertion, assOk := res["assertion"].(string); assOk {
			assertion = _assertion
			return
		}

	}

	return
}
