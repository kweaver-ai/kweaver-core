package datasourcevalobj

import "github.com/pkg/errors"

// KnEntrySource 表示知识条目类型数据源
type KnEntrySource struct {
	KnEntryID string `json:"kn_entry_id" binding:"required"` // 知识条目ID
}

func (p *KnEntrySource) ValObjCheck() (err error) {
	// 检查KnEntryID是否为空
	if p.KnEntryID == "" {
		err = errors.New("[KnEntrySource]: kn_entry_id is required")
		return
	}

	return
}

func (p *KnEntrySource) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"KnEntryID.required": `"kn_entry_id"不能为空`,
	}
}
