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

// retrievalConfigToLocal 将 any 形式的 retrieval_config 转为 *KnSearchRetrievalConfig。
// 当 cfg 为 *RetrievalConfig 时走显式拷贝：避免 JSON 往返时 bool 字段带 omitempty 导致 false 被省略，
// 进而 Unmarshal 到 *bool 时变成 nil（与显式 false 语义不同）。
func retrievalConfigToLocal(cfg any) *interfaces.KnSearchRetrievalConfig {
	if cfg == nil {
		return nil
	}
	switch v := cfg.(type) {
	case *interfaces.RetrievalConfig:
		return retrievalConfigStructToLocal(v)
	case interfaces.RetrievalConfig:
		return retrievalConfigStructToLocal(&v)
	default:
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
}

func retrievalConfigStructToLocal(rc *interfaces.RetrievalConfig) *interfaces.KnSearchRetrievalConfig {
	if rc == nil {
		return nil
	}
	out := &interfaces.KnSearchRetrievalConfig{}
	if rc.ConceptRetrieval != nil {
		cr := rc.ConceptRetrieval
		out.ConceptRetrieval = &interfaces.KnSearchConceptRetrievalConfig{
			TopK:                   cr.TopK,
			IncludeSampleData:      boolPtr(cr.IncludeSampleData),
			SchemaBrief:            boolPtr(cr.SchemaBrief),
			EnableCoarseRecall:     boolPtr(cr.EnableCoarseRecall),
			CoarseObjectLimit:      cr.CoarseObjectLimit,
			CoarseRelationLimit:    cr.CoarseRelationLimit,
			CoarseMinRelationCount: cr.CoarseMinRelationCount,
			EnablePropertyBrief:    boolPtr(cr.EnablePropertyBrief),
			PerObjectPropertyTopK:  cr.PerObjectPropertyTopK,
			GlobalPropertyTopK:     cr.GlobalPropertyTopK,
		}
	}
	if rc.SemanticInstanceRetrieval != nil {
		s := rc.SemanticInstanceRetrieval
		out.SemanticInstanceRetrieval = &interfaces.KnSearchSemanticInstanceRetrievalConfig{
			InitialCandidateCount:             s.InitialCandidateCount,
			PerTypeInstanceLimit:              s.PerTypeInstanceLimit,
			MaxSemanticSubConditions:          s.MaxSemanticSubConditions,
			SemanticFieldKeepRatio:            s.SemanticFieldKeepRatio,
			SemanticFieldKeepMin:              s.SemanticFieldKeepMin,
			SemanticFieldKeepMax:              s.SemanticFieldKeepMax,
			SemanticFieldRerankBatchSize:      s.SemanticFieldRerankBatchSize,
			MinDirectRelevance:                s.MinDirectRelevance,
			EnableGlobalFinalScoreRatioFilter: boolPtr(s.EnableGlobalFinalScoreRatioFilter),
			GlobalFinalScoreRatio:             s.GlobalFinalScoreRatio,
			ExactNameMatchScore:               s.ExactNameMatchScore,
		}
	}
	if rc.PropertyFilter != nil {
		p := rc.PropertyFilter
		out.PropertyFilter = &interfaces.KnSearchPropertyFilterConfig{
			MaxPropertiesPerInstance: p.MaxPropertiesPerInstance,
			MaxPropertyValueLength:   p.MaxPropertyValueLength,
			EnablePropertyFilter:     boolPtr(p.EnablePropertyFilter),
		}
	}
	return out
}

// KnSearchLocalResponseToResp 将 KnSearchLocalResponse 转为 KnSearchResp
func KnSearchLocalResponseToResp(local *interfaces.KnSearchLocalResponse) *interfaces.KnSearchResp {
	if local == nil {
		return nil
	}
	return &interfaces.KnSearchResp{
		ObjectTypes:   local.ObjectTypes,
		RelationTypes: local.RelationTypes,
		ActionTypes:   local.ActionTypes,
	}
}
