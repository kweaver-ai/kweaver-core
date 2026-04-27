package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	listSmallModelPath      = "/v1/small-model/list"
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
func (m *mfModelManager) buildHeaders(ctx context.Context) map[string]string {
	headers := common.GetHeaderFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	if accountID, ok := headers[string(interfaces.HeaderXAccountID)]; !ok || accountID == "" {
		headers[string(interfaces.HeaderXAccountID)] = interfaces.ADMIN_ACCOUNT_ID
		headers[string(interfaces.HeaderXAccountType)] = interfaces.ADMIN_ACCOUNT_TYPE
	}
	return headers
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

// GetEmbeddingModel 获取 embedding 模型信息
func (m *mfModelManager) GetEmbeddingModel(ctx context.Context, modelName string, modelType string) (resp *interfaces.EmbeddingModel, err error) {
	src := fmt.Sprintf("%s%s", m.baseURL, listSmallModelPath)
	query := url.Values{}
	query.Set("model_name", modelName)
	query.Set("model_type", modelType)
	header := m.buildHeaders(ctx)
	_, respData, err := m.httpClient.Get(ctx, src, query, header)
	if err != nil {
		m.logger.WithContext(ctx).Errorf("failed to get embedding model: %v", err)
		return nil, err
	}

	var payload map[string]any
	if err = utils.AnyToObject(respData, &payload); err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		m.logger.WithContext(ctx).Errorf("failed to convert embedding model response: %v", err)
		return nil, err
	}

	models := make([]*interfaces.EmbeddingModel, 0)
	for _, key := range []string{"res", "data", "list"} {
		if raw, ok := payload[key]; ok {
			if err = utils.AnyToObject(raw, &models); err == nil && len(models) > 0 {
				return models[0], nil
			}
		}
	}

	for _, key := range []string{"res", "data"} {
		if raw, ok := payload[key]; ok {
			var nested map[string]any
			if err = utils.AnyToObject(raw, &nested); err != nil {
				continue
			}
			for _, nestedKey := range []string{"list", "entries", "models"} {
				if nestedRaw, ok := nested[nestedKey]; ok {
					if err = utils.AnyToObject(nestedRaw, &models); err == nil && len(models) > 0 {
						return models[0], nil
					}
				}
			}
		}
	}

	err = errors.DefaultHTTPError(ctx, http.StatusNotFound, "embedding model not found")
	m.logger.WithContext(ctx).Warnf("embedding model not found, model_name=%s, model_type=%s", modelName, modelType)
	return nil, err
}
