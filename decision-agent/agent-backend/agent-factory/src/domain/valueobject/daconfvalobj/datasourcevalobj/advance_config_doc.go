package datasourcevalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type DocAdvancedConfig struct {
	RetrievalSlicesNum *int `json:"retrieval_slices_num" binding:"required"` // 召回切片数：150
	MaxSlicePerCite    *int `json:"max_slice_per_cite" binding:"required"`   // 每篇文档最大保留切片数：16
	// CosSimWeight             float64 `json:"cos_sim_weight"`
	// BM25Weight               float64 `json:"bm_25_weight"`
	// RerankerMethod           string  `json:"reranker_method"`
	// AccRankingScoreThreshold float64 `json:"acc_ranking_score_threshold"`
	// ChooseMethod             string  `json:"choose_method"`
	RerankTopK        *int     `json:"rerank_topk" binding:"required"`        // 重排序后保留切片数 ：15
	SliceHeadNum      *int     `json:"slice_head_num" binding:"required"`     // 获取上文切片数：2
	SliceTailNum      *int     `json:"slice_tail_num" binding:"required"`     // 获取下文切片数：0
	DocumentsNum      *int     `json:"documents_num" binding:"required"`      // 来源文档数量：8
	DocumentThreshold *float64 `json:"document_threshold" binding:"required"` // 文档ranker bge相似度阈值：-5.5

	RetrievalMaxLength *int `json:"retrieval_max_length" binding:"required"` // 召回文档最大长度
}

func (c *DocAdvancedConfig) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"RetrievalSlicesNum.required": `"retrieval_slices_num"不能为空`,
		"MaxSlicePerCite.required":    `"max_slice_per_cite"不能为空`,
		"RerankTopK.required":         `"rerank_topk"不能为空`,
		"SliceHeadNum.required":       `"slice_head_num"不能为空`,
		"SliceTailNum.required":       `"slice_tail_num"不能为空`,
		"DocumentsNum.required":       `"documents_num"不能为空`,
		"DocumentThreshold.required":  `"document_threshold"不能为空`,
		"RetrievalMaxLength.required": `"retrieval_max_length"不能为空`,
	}
}

func (c *DocAdvancedConfig) ValObjCheck() (err error) {
	if !cutil.CheckInRange(*c.RetrievalSlicesNum, 50, 200) {
		return errors.New("[DocAdvancedConfig]: retrieval_slices_num must between 50 and 200")
	}

	if !cutil.CheckInRange(*c.RerankTopK, 10, 30) {
		return errors.New("[DocAdvancedConfig]: rerank_topk must between 10 and 30")
	}

	if !cutil.CheckInRange(*c.SliceHeadNum, 0, 3) {
		return errors.New("[DocAdvancedConfig]: slice_head_num must between 0 and 3")
	}

	if !cutil.CheckInRange(*c.SliceTailNum, 0, 3) {
		return errors.New("[DocAdvancedConfig]: slice_tail_num must between 0 and 3")
	}

	if !cutil.CheckInRange(*c.DocumentsNum, 4, 10) {
		return errors.New("[DocAdvancedConfig]: documents_num must between 4 and 10")
	}

	if !cutil.CheckInRange(*c.MaxSlicePerCite, 5, 20) {
		return errors.New("[DocAdvancedConfig]: max_slice_per_cite must between 5 and 20")
	}

	if !cutil.CheckInRange(*c.DocumentThreshold, -10, 10) {
		return errors.New("[DocAdvancedConfig]: document_threshold must between -10 and 10")
	}

	return nil
}
