package agentrespvo

type Interventions []*Intervention

func NewInterventions() Interventions {
	return make(Interventions, 0)
}

func (is Interventions) ToOutputVarMap() (outputVarMap map[string][]*Intervention) {
	outputVarMap = make(map[string][]*Intervention)

	for _, intervention := range is {
		if intervention.ToolCallInfo == nil {
			continue
		}

		outputVar := intervention.ToolCallInfo.OutputVar
		if outputVar == "" {
			continue
		}

		outputVarMap[outputVar] = append(outputVarMap[outputVar], intervention)
	}

	return
}

type Intervention struct {
	ToolName     string        `json:"tool_name"`
	ToolCallInfo *ToolCallInfo `json:"tool_call_info"`
}

type ToolCallInfo struct {
	ToolName   string      `json:"tool_name"`
	Args       interface{} `json:"args"`
	AssignType string      `json:"assign_type"`
	OutputVar  string      `json:"output_var"`
}
