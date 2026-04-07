package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

var (
	mfModelManagerOnce     sync.Once
	mfModelManagerInstance interfaces.MFModelManager
)

var (
	getPromptByPromptIDPath = "/v1/prompt/%s"
)

type mfModelManager struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

func NewMFModelManager() interfaces.MFModelManager {
	mfModelManagerOnce.Do(func() {
		conf := config.NewConfigLoader()
		mfModelManagerInstance = &mfModelManager{
			baseURL: fmt.Sprintf("%s://%s:%d/api/private/mf-model-manager", conf.MFModelManager.PrivateProtocol,
				conf.MFModelManager.PrivateHost, conf.MFModelManager.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return mfModelManagerInstance
}

// GetPromptByPromptID 获取提示词
func (m *mfModelManager) GetPromptByPromptID(ctx context.Context, promptID string) (resp *interfaces.GetPromptResp, err error) {
	src := fmt.Sprintf("%s%s", m.baseURL, fmt.Sprintf(getPromptByPromptIDPath, promptID))
	header := common.GetHeaderFromCtx(ctx)
	_, respData, err := m.httpClient.Get(ctx, src, nil, header)
	if err != nil {
		m.logger.WithContext(ctx).Errorf("failed to get prompt by promptID: %v", err)
		return nil, err
	}
	result := map[string]any{}
	// 转换为map[string]any
	err = utils.AnyToObject(respData, &result)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		m.logger.WithContext(ctx).Errorf("failed to convert respData to map[string]any: %v", err)
		return nil, err
	}
	resp = &interfaces.GetPromptResp{}
	err = utils.AnyToObject(result["res"], resp)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		m.logger.WithContext(ctx).Errorf("failed to convert respData to GetPromptResp: %v", err)
		return nil, err
	}
	return resp, nil
}
