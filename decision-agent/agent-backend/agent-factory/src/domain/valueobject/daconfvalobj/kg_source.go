package daconfvalobj

import "github.com/pkg/errors"

// KgSource 表示图谱类型数据源
type KgSource struct {
	KgID            string              `json:"kg_id" binding:"required"`  // 图谱ID
	Fields          []string            `json:"fields" binding:"required"` // 实体类范围
	FieldProperties map[string][]string `json:"field_properties"`          // 实体属性列表
	OutputFields    []string            `json:"output_fields"`             // 结果范围，非必填
}

func (p *KgSource) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"KgID.required":   `"kg_id"不能为空`,
		"Fields.required": `"fields"不能为空`,
	}
}

func (p *KgSource) ValObjCheck() (err error) {
	// 检查KgID是否为空
	if p.KgID == "" {
		err = errors.New("[KgSource]: kg_id is required")
		return
	}

	// 检查Fields是否为空
	if p.Fields == nil {
		err = errors.New("[KgSource]: fields is required")
		return
	}

	return
}
