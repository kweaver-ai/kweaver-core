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
	mfModelAPIClientOnce     sync.Once
	mfModelAPIClientInstance interfaces.MFModelAPIClient // 模型管理API客户端实例
)

var (
	chatCompletionsPath = "/v1/chat/completions" // 模型调用路径
	embeddingsPath      = "/v1/small-model/embeddings"
)

// mfModelAPIClient 模型管理API客户端
type mfModelAPIClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

// NewMFModelAPIClient 创建模型管理API客户端
func NewMFModelAPIClient() interfaces.MFModelAPIClient {
	mfModelAPIClientOnce.Do(func() {
		conf := config.NewConfigLoader()
		mfModelAPIClientInstance = &mfModelAPIClient{
			baseURL: fmt.Sprintf("%s://%s:%d/api/private/mf-model-api", conf.MFModelAPI.PrivateProtocol,
				conf.MFModelAPI.PrivateHost, conf.MFModelAPI.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return mfModelAPIClientInstance
}

// ChatCompletion 调用模型
func (um *mfModelAPIClient) ChatCompletion(ctx context.Context, req *interfaces.ChatCompletionReq) (resp *interfaces.ChatCompletionResp, err error) {
	src := fmt.Sprintf("%s%s", um.baseURL, chatCompletionsPath)
	req.Stream = false
	header := common.GetHeaderFromCtx(ctx)
	_, result, err := um.httpClient.Post(ctx, src, header, req)
	if err != nil {
		um.logger.WithContext(ctx).Warnf("CallModel failed, err: %v", err)
		return nil, err
	}
	resp = &interfaces.ChatCompletionResp{}
	err = utils.AnyToObject(result, resp)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		um.logger.WithContext(ctx).Warnf("AnyToObject failed, err: %v", err)
		return nil, err
	}
	return resp, nil
}

// StreamChatCompletion 调用模型流式返回
func (um *mfModelAPIClient) StreamChatCompletion(ctx context.Context, req *interfaces.ChatCompletionReq) (chan string, chan error, error) {
	src := fmt.Sprintf("%s%s", um.baseURL, chatCompletionsPath)
	um.logger.WithContext(ctx).Infof("Stream call model: %s", src)
	// 设置流式请求
	req.Stream = true
	// 创建HTTP请求
	header := common.GetHeaderFromCtx(ctx)
	streamCh, errCh, err := um.httpClient.PostStream(ctx, src, header, req)
	if err != nil {
		um.logger.WithContext(ctx).Warnf("StreamChatCompletion failed, err: %v", err)
		return nil, nil, err
	}
	return streamCh, errCh, nil
}

// Embeddings 获取 embedding 向量
func (um *mfModelAPIClient) Embeddings(ctx context.Context, req *interfaces.EmbeddingReq) (resp *interfaces.EmbeddingResp, err error) {
	src := fmt.Sprintf("%s%s", um.baseURL, embeddingsPath)
	header := common.GetHeaderFromCtx(ctx)
	um.logger.WithContext(ctx).Infof("request embeddings, url=%s, inputs=%d", src, len(req.Input))
	_, result, err := um.httpClient.Post(ctx, src, header, req)
	if err != nil {
		um.logger.WithContext(ctx).Warnf("embeddings failed, url=%s, inputs=%d, err=%v", src, len(req.Input), err)
		return nil, err
	}
	resp = &interfaces.EmbeddingResp{}
	err = utils.AnyToObject(result, resp)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		um.logger.WithContext(ctx).Warnf("parse embeddings response failed, url=%s, err=%v", src, err)
		return nil, err
	}
	um.logger.WithContext(ctx).Infof("embeddings success, url=%s, vectors=%d", src, len(resp.Data))
	return resp, nil
}
