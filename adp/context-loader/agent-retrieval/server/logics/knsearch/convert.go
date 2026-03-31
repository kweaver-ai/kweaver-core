// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch provides business logic for knowledge network search operations.
// file: convert.go
// description: KnSearchReq/KnSearchResp 与本地请求/响应的转换
package knsearch

import (
	"encoding/json"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// KnSearchReqToLocal 将 KnSearchReq 转为 KnSearchLocalRequest
func KnSearchReqToLocal(req *interfaces.KnSearchReq) *interfaces.KnSearchLocalRequest {
	if req == nil {
		return nil
	}
	local := &interfaces.KnSearchLocalRequest{
		AccountID:   req.XAccountID,
		AccountType: req.XAccountType,
		Query:       req.Query,
		KnID:        req.KnID,
	}
	local.OnlySchema = false
	if req.OnlySchema != nil {
		local.OnlySchema = *req.OnlySchema
	}
	local.EnableRerank = true
	if req.EnableRerank != nil {
		local.EnableRerank = *req.EnableRerank
	}
	local.RetrievalConfig = retrievalConfigToLocal(req.RetrievalConfig)
	return local
}

// retrievalConfigToLocal 将 any 形式的 retrieval_config 转为 *KnSearchRetrievalConfig
func retrievalConfigToLocal(cfg any) *interfaces.KnSearchRetrievalConfig {
	if cfg == nil {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	var local interfaces.KnSearchRetrievalConfig
	if err := json.Unmarshal(data, &local); err != nil {
		return nil
	}
	return &local
}

// KnSearchLocalResponseToResp 将 KnSearchLocalResponse 转为 KnSearchResp
func KnSearchLocalResponseToResp(local *interfaces.KnSearchLocalResponse) *interfaces.KnSearchResp {
	if local == nil {
		return nil
	}
	resp := &interfaces.KnSearchResp{
		ObjectTypes:   local.ObjectTypes,
		RelationTypes: local.RelationTypes,
		ActionTypes:   local.ActionTypes,
		Nodes:         local.Nodes,
	}
	if local.Message != "" {
		resp.Message = &local.Message
	}
	return resp
}
