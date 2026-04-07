// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

var (
	knowledgeRerank = "/tools/knowledge_rerank" // 知识重排接口
	knSearch        = "/tools/kn_search"        // kn_search 接口
)

var (
	drOnce sync.Once
	dr     interfaces.DataRetrieval
)

type dataRetrievalClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

// NewDataRetrievalClient 新建DataRetrievalClient
func NewDataRetrievalClient() interfaces.DataRetrieval {
	drOnce.Do(func() {
		conf := config.NewConfigLoader()
		dr = &dataRetrievalClient{
			baseURL:    conf.DataRetrieval.BuildURL(""),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return dr
}

// KnowledgeRerank 知识重排
func (dr *dataRetrievalClient) KnowledgeRerank(ctx context.Context, req *interfaces.KnowledgeRerankReq) (results []*interfaces.ConceptResult, err error) {
	src := fmt.Sprintf("%s%s", dr.baseURL, knowledgeRerank)
	header := map[string]string{
		rest.ContentTypeKey: rest.ContentTypeJSON,
	}
	_, respData, err := dr.httpClient.Post(ctx, src, header, req)
	if err != nil {
		dr.logger.WithContext(ctx).Errorf("KnowledgeRerank failed, err: %v", err)
		return
	}
	results = []*interfaces.ConceptResult{}
	err = utils.AnyToObject(respData, &results)
	if err != nil {
		dr.logger.WithContext(ctx).Errorf("KnowledgeRerank failed, err: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// KnSearch 知识网络检索
func (dr *dataRetrievalClient) KnSearch(ctx context.Context, req *interfaces.KnSearchReq) (resp *interfaces.KnSearchResp, err error) {
	url := fmt.Sprintf("%s%s", dr.baseURL, knSearch)

	// 构建请求头 - 透传 Header 参数
	header := map[string]string{
		rest.ContentTypeKey: rest.ContentTypeJSON,
	}
	if req.XAccountID != "" {
		header["x-account-id"] = req.XAccountID
	}
	if req.XAccountType != "" {
		header["x-account-type"] = req.XAccountType
	}

	// 构建请求体 - 直接透传所有 Body 参数
	body := map[string]any{
		"query":  req.Query,
		"kn_ids": req.GetKnIDs(),
	}
	if req.RetrievalConfig != nil {
		body["retrieval_config"] = req.RetrievalConfig
	}
	if req.OnlySchema != nil {
		body["only_schema"] = *req.OnlySchema
	}
	if req.EnableRerank != nil {
		body["enable_rerank"] = *req.EnableRerank
	}

	// 记录请求日志
	bodyJSON, _ := json.Marshal(body)
	dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] URL: %s", url)
	dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Request Body: %s", string(bodyJSON))

	// 发送请求
	_, respBody, err := dr.httpClient.Post(ctx, url, header, body)
	if err != nil {
		dr.logger.WithContext(ctx).Errorf("[DataRetrieval#KnSearch] Request failed, err: %v", err)
		return nil, err
	}

	// 解析响应 - 直接解析到 any
	resp = &interfaces.KnSearchResp{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		dr.logger.WithContext(ctx).Errorf("[DataRetrieval#KnSearch] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析 kn_search 响应失败: %v", err))
		return nil, err
	}

	// 记录响应日志
	respJSON, _ := json.Marshal(resp)
	dr.logger.WithContext(ctx).Debugf("[DataRetrieval#KnSearch] Response: %s", string(respJSON))

	return resp, nil
}
