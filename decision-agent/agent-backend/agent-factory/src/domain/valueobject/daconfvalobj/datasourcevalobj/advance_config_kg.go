package datasourcevalobj

import (
	"math"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type KGAdvancedConfig struct {
	TextMatchEntityNums   *int     `json:"text_match_entity_nums" binding:"required"`   // 文本匹配召回实体数量：60
	VectorMatchEntityNums *int     `json:"vector_match_entity_nums" binding:"required"` // 向量匹配召回实体数量：60
	GraphRagTopK          *int     `json:"graph_rag_topk" binding:"required"`           // 重排序后保留参考信息数量：25
	LongTextLength        *int     `json:"long_text_length" binding:"required"`         // 长文本：256
	RerankerSimThreshold  *float64 `json:"reranker_sim_threshold" binding:"required"`   // 图谱rerank bge相似度过滤阈值：-5.5
	// EnableRAG             *bool    `json:"enable_rag" binding:"required"`               // 是否启用RAG
	// EnableNGQL            *bool    `json:"enable_ngql" binding:"required"`              // 是否启用NGQL
	RetrievalMaxLength *int `json:"retrieval_max_length" binding:"required"` // 召回文档最大长度
}

func (c *KGAdvancedConfig) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"TextMatchEntityNums.required":   `"text_match_entity_nums"不能为空`,
		"VectorMatchEntityNums.required": `"vector_match_entity_nums"不能为空`,
		"GraphRagTopK.required":          `"graph_rag_topk"不能为空`,
		"LongTextLength.required":        `"long_text_length"不能为空`,
		"RerankerSimThreshold.required":  `"reranker_sim_threshold"不能为空`,
		//"EnableRAG.required":             `"enable_rag"不能为空`,
		//"EnableNGQL.required":            `"enable_ngql"不能为空`,
		"RetrievalMaxLength.required": `"retrieval_max_length"不能为空`,
	}
}

func (c *KGAdvancedConfig) ValObjCheck() (err error) {
	//if *c.EnableRAG && *c.EnableNGQL {
	//	return errors.New("[KGAdvancedConfig]: enable_rag/enable_ngql cannot both be false")
	//}
	if !cutil.CheckInRange(*c.TextMatchEntityNums, 40, 100) {
		return errors.New("[KGAdvancedConfig]: text_match_entity_nums must between 40 and 100")
	}

	if !cutil.CheckInRange(*c.RerankerSimThreshold, -10, 10) {
		return errors.New("[KGAdvancedConfig]: reranker_sim_threshold must between -10 and 10")
	}

	*c.RerankerSimThreshold = math.Round(*c.RerankerSimThreshold*100) / 100

	if !cutil.CheckInRange(*c.GraphRagTopK, 10, 100) {
		return errors.New("[KGAdvancedConfig]: graph_rag_topk must between 10 and 100")
	}

	if !cutil.CheckMin(*c.LongTextLength, 50) {
		return errors.New("[KGAdvancedConfig]: long_text_length must be greater than 50")
	}

	if !cutil.CheckInRange(*c.VectorMatchEntityNums, 40, 100) {
		return errors.New("[KGAdvancedConfig]: vector_match_entity_nums must between 40 and 100")
	}

	return nil
}
