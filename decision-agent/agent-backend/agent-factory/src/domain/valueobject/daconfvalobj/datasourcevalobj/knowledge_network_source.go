package datasourcevalobj

import "github.com/pkg/errors"

type KnowledgeNetworkSource struct {
	KnowledgeNetworkID string        `json:"knowledge_network_id" binding:"required"` // 业务知识网络ID
	ObjectTypes        []*ObjectType `json:"object_types"`                            // 对象类型列表
}

func (p *KnowledgeNetworkSource) ValObjCheck() (err error) {
	// 检查KnowledgeNetworkID是否为空
	if p.KnowledgeNetworkID == "" {
		err = errors.New("[KnowledgeNetworkSource]: knowledge_network_id is required")
		return
	}

	if p.ObjectTypes == nil {
		return
	}

	// 检查ObjectTypes中的每个对象类型是否为空
	for _, objectType := range p.ObjectTypes {
		if err = objectType.ValObjCheck(); err != nil {
			err = errors.Wrap(err, "[KnowledgeNetworkSource]: object_type is invalid")
			return
		}
	}

	return
}

func (p *KnowledgeNetworkSource) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"KnowledgeNetworkID.required": `"knowledge_network_id"不能为空`,
	}
}

type ObjectType struct {
	ObjectTypeID string `json:"object_type_id" binding:"required"` // 对象类型ID
}

func (p *ObjectType) ValObjCheck() (err error) {
	// 检查ObjectTypeID是否为空
	if p.ObjectTypeID == "" {
		err = errors.New("[ObjectType]: object_type_id is required")
		return
	}

	return
}

func (p *ObjectType) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"ObjectTypeID.required": `"object_type_id"不能为空`,
	}
}
