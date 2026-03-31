package datasourcevalobj

import "github.com/pkg/errors"

// MetricSource 表示指标类型数据源
type MetricSource struct {
	MetricModelID string `json:"metric_model_id" binding:"required"` // 指标模型ID
}

func (p *MetricSource) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"MetricModelID.required": `"metric_model_id"不能为空`,
	}
}

func (p *MetricSource) ValObjCheck() (err error) {
	// 检查MetricModelID是否为空
	if p.MetricModelID == "" {
		err = errors.New("[MetricSource]: metric_model_id is required")
		return
	}

	return
}
