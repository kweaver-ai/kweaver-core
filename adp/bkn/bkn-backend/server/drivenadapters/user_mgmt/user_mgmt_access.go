// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	umAccessOnce sync.Once
	umAccess     interfaces.UserMgmtAccess
)

type userMgmtAccess struct {
	appSetting  *common.AppSetting
	httpClient  rest.HTTPClient
	userMgmtUrl string
}

// NewUserMgmtAccess 创建用户管理访问实例
func NewUserMgmtAccess(appSetting *common.AppSetting) interfaces.UserMgmtAccess {
	umAccessOnce.Do(func() {
		umAccess = &userMgmtAccess{
			appSetting:  appSetting,
			httpClient:  common.NewHTTPClient(),
			userMgmtUrl: appSetting.UserMgmtUrl,
		}
	})

	return umAccess
}

func (uma *userMgmtAccess) GetAccountNames(ctx context.Context, accountInfos []*interfaces.AccountInfo) error {
	if len(accountInfos) == 0 {
		return nil
	}

	// 构建请求URL
	httpUrl := fmt.Sprintf("%s/api/user-management/v2/names", uma.userMgmtUrl)

	userIDMap := map[string]string{}
	appIDMap := map[string]string{}
	userIDs := []string{}
	appIDs := []string{}
	for _, accountInfo := range accountInfos {
		switch accountInfo.Type {
		case interfaces.ACCESSOR_TYPE_USER:
			if _, ok := userIDMap[accountInfo.ID]; !ok {
				userIDMap[accountInfo.ID] = "-"
				userIDs = append(userIDs, accountInfo.ID)
			}
		case interfaces.ACCESSOR_TYPE_APP:
			if _, ok := appIDMap[accountInfo.ID]; !ok {
				appIDMap[accountInfo.ID] = "-"
				appIDs = append(appIDs, accountInfo.ID)
			}
		}
	}

	// 构建请求体
	requestBody := map[string]any{
		"method":   http.MethodGet,
		"user_ids": userIDs,
		"app_ids":  appIDs,
		"strict":   false,
	}

	// 设置请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 发送POST请求获取用户信息
	respCode, result, err := uma.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, requestBody)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Get account names request failed: %v", err)
		return fmt.Errorf("get account names request failed: %w", err)
	}

	if respCode != 200 {
		logger.Errorf("Get account names request failed with status code: %d", respCode)
		return fmt.Errorf("get account names request failed with status code: %d", respCode)
	}

	// "{\"app_names\":[{\"id\":\"91efa756-11cc-49d7-ab25-f6e18f9305fe\",\"name\":\"kwww\"}],\"user_names\":[{\"id\":\"f6c6e398-ce82-11f0-888f-3ac1298ec09f\",\"name\":\"kww\"}],\"department_names\":[],\"contactor_names\":[],\"group_names\":[]}"
	// 解析响应数据
	response := struct {
		AppNames []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"app_names"`
		UserNames []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user_names"`
	}{}

	if err := sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("Unmarshal account names response failed: %v", err)
		return fmt.Errorf("unmarshal account names response failed: %w", err)
	}

	for _, user := range response.UserNames {
		userIDMap[user.ID] = user.Name
	}
	for _, app := range response.AppNames {
		appIDMap[app.ID] = app.Name
	}
	for _, accountInfo := range accountInfos {
		switch accountInfo.Type {
		case interfaces.ACCESSOR_TYPE_USER:
			if name, ok := userIDMap[accountInfo.ID]; ok {
				accountInfo.Name = name
			} else {
				accountInfo.Name = "-"
			}
		case interfaces.ACCESSOR_TYPE_APP:
			if name, ok := appIDMap[accountInfo.ID]; ok {
				accountInfo.Name = name
			} else {
				accountInfo.Name = "-"
			}
		}
	}

	return nil
}
