package agentrespvo

import (
	"encoding/json"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/util"
)

// AnswerS 表示回答数据的结构体（外层answer结构体，里面包含具体的final answer等相关变量的值）
/*
示例：
{
  "answer"【对应这块的结构体】: {
    "query": "爱数是一家怎样的公司",
    "previous_status": {
      "tool_time": 0,
      "judge_time": 0,
      "prompt_time": 0,
      "explore_time": 0,
      "assign_time": 0
    },
    "status": {
      "tool_time": 0,
      "judge_time": 0,
      "prompt_time": 1,
      "explore_time": 0,
      "assign_time": 0
    },
    "answer": {
      "answer": "爱数（AnyShare）是一家专注于数据管理与服务的公司，总部位于中国上海。爱数成立于2006年，其主要业务包括数据备份、恢复、归档、分析以及数据治理等，致力于为企业提供全面的数据管理解决方案。爱数的产品和服务覆盖了从数据保护到数据价值挖掘的整个生命周期，帮助企业实现数据的高效利用和安全保护。此外，爱数还提供专业的技术支持和咨询服务，帮助客户优化数据管理流程，提升数据治理能力。爱数在数据管理领域拥有丰富的经验和深厚的技术积累，是该领域的领先企业之一。",
      "think": ""
    },
    "usage": {
      "prompt_tokens": 12,
      "total_tokens": 266,
      "completion_tokens": 254
    }
  },
  "status": "True"
}
*/
// S: struct
type AnswerS struct {
	Interventions Interventions `json:"interventions"` // 存储所有的中断相关信息
	Progress      []*Progress   `json:"_progress"`     // Dolphin中间执行过程展示
	util.DynamicFieldsHolder
}

func NewAnswerS() *AnswerS {
	return &AnswerS{
		Interventions: NewInterventions(),
	}
}

// 如果一个结构体实现了 json.Marshaler 接口（即 MarshalJSON 方法），Sonic 会调用这个方法来进行序列化。这与 encoding/json 的行为一致。
func (p *AnswerS) MarshalJSON() ([]byte, error) {
	baseMap := map[string]interface{}{}

	p.AddDynamicFieldsToMap(baseMap)

	return sonic.Marshal(baseMap)
}

// UnmarshalJSON 自定义 JSON 反序列化
func (p *AnswerS) UnmarshalJSON(data []byte) error {
	type TempProduct AnswerS

	temp := struct {
		*TempProduct
	}{
		TempProduct: (*TempProduct)(p),
	}

	var objMap map[string]json.RawMessage
	if err := sonic.Unmarshal(data, &objMap); err != nil {
		return err
	}

	if err := sonic.Unmarshal(data, &temp); err != nil {
		return err
	}

	if p.DynamicFields == nil {
		p.DynamicFields = make(map[string]interface{})
	}

	knownFields := map[string]bool{
		// "query":         true,
		"interventions":   true,
		"_progress":       true, // Dolphin中间执行过程展示，这个字段不能去掉，否则断言失败；_progress是数组不是map
		"query":           true,
		"history":         true,
		"header":          true,
		"self-config":     true,
		"previous_status": true,
		"status":          true,
	}

	for key, value := range objMap {
		if !knownFields[key] {
			var v interface{}
			if err := sonic.Unmarshal(value, &v); err != nil {
				return err
			}

			p.SetField(key, v)
		}
	}

	return nil
}
