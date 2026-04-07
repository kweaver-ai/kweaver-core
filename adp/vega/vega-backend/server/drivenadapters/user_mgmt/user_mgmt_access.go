package user_mgmt

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common"
	"vega-backend/interfaces"
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

func (u *userMgmtAccess) GetAccountNames(ctx context.Context, accountInfos []*interfaces.AccountInfo) error {
	if len(accountInfos) == 0 {
		return nil
	}

	httpUrl := fmt.Sprintf("%s/api/user-management/v2/names", u.userMgmtUrl)

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

	requestBody := map[string]any{
		"method":   http.MethodGet,
		"user_ids": userIDs,
		"app_ids":  appIDs,
		"strict":   false,
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	respCode, result, err := u.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, requestBody)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Get account names request failed: %v", err)
		return fmt.Errorf("get account names request failed: %w", err)
	}

	if respCode != 200 {
		logger.Errorf("Get account names request failed with status code: %d", respCode)
		return fmt.Errorf("get account names request failed with status code: %d", respCode)
	}

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
