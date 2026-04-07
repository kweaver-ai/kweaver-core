package agentrespvo

import (
	"errors"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/tidwall/gjson"
)

type MiddleOutputVarItem struct {
	VarName  string                    `json:"var_name"`
	Type     chatresenum.OutputVarType `json:"type"`
	Value    interface{}               `json:"value"`
	Thinking string                    `json:"thinking"`

	Interventions []*Intervention `json:"interventions"`
}
type MiddleOutputVarRes struct {
	Vars []*MiddleOutputVarItem `json:"vars"`
}

func NewMiddleOutputVarRes() *MiddleOutputVarRes {
	return &MiddleOutputVarRes{
		Vars: make([]*MiddleOutputVarItem, 0),
	}
}

func (r *MiddleOutputVarRes) LoadFrom(vars []string, valuesMap map[string]interface{}, outputVarInterventionMap map[string][]*Intervention) (err error) {
	for _, varName := range vars {
		if val, ok := valuesMap[varName]; ok {
			// 1. 判断类型
			varType := getVarType(val)

			// 2. 获取prompt类型的val和thinkingVal
			var thinkingVal string
			if varType == chatresenum.OutputVarTypePrompt {
				val, thinkingVal, err = getPromptVal(val)
				if err != nil {
					return
				}
			}

			// 3. 获取中断信息
			interventions := make([]*Intervention, 0)

			if outputVarInterventionMap != nil {
				if v, _ok := outputVarInterventionMap[varName]; _ok {
					interventions = v
				}
			}

			// 4. 添加到r.Vars
			r.Vars = append(r.Vars, &MiddleOutputVarItem{
				VarName:       varName,
				Type:          varType,
				Value:         val,
				Thinking:      thinkingVal,
				Interventions: interventions,
			})
		}
	}

	return
}

func getPromptVal(val interface{}) (promptVal, thinkingVal string, err error) {
	// 1. 转换为 json
	byt, err := sonic.Marshal(val)
	if err != nil {
		return
	}

	// 2. 解析 json
	j := gjson.ParseBytes(byt)

	// 3. 获取 answer
	v := j.Get("answer").Value()

	var ok bool

	promptVal, ok = v.(string)
	if !ok {
		err = errors.New("[getPromptVal]: answer is not string")
		return
	}

	// 4. 获取 think
	thinkingVal, ok = j.Get("think").Value().(string)
	if !ok {
		err = errors.New("[getPromptVal]: think is not string")
		return
	}

	return
}

func getVarType(val interface{}) (varType chatresenum.OutputVarType) {
	// 1. 判断是否为 prompt 类型
	isPromptValid, err := IsPromptTypeInterface(val)
	if err != nil {
		return
	}

	if isPromptValid {
		varType = chatresenum.OutputVarTypePrompt
		return
	}

	// 2. 判断是否为 explore 类型
	isExploreValid, err := IsExploreTypeInterface(val)
	if err != nil {
		return
	}

	if isExploreValid {
		varType = chatresenum.OutputVarTypeExplore
		return
	}

	// 3. 判断是否为其他类型
	varType = chatresenum.OutputVarTypeOther

	return
}

func (r *MiddleOutputVarRes) ToExploreList(val interface{}) (exploreList []*AnswerExplore, err error) {
	exploreList = make([]*AnswerExplore, 0)

	err = cutil.CopyUseJSON(&exploreList, val)
	if err != nil {
		return
	}

	return
}
