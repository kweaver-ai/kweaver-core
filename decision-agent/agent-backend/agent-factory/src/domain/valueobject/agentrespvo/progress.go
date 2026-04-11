package agentrespvo

type Progress struct {
	ID        string `json:"id"`
	AgentName string `json:"agent_name"`
	Stage     string `json:"stage"`
	// NOTE: 如果是工具调用结果，则不是string
	Answer    interface{} `json:"answer"`
	Think     interface{} `json:"think"`
	Status    string      `json:"status"`
	SkillInfo *SkillInfo  `json:"skill_info"`
	// NOTE: 如果是工具调用结果，则不是string
	// BlockAnswer  interface{} `json:"block_answer"`
	InputMessage interface{} `json:"input_message"`
	Interrupted  bool        `json:"interrupted"`
	// 自定义标签，用于前端控制是否展示 progress
	Flags                 interface{} `json:"flags"`
	StartTime             float64     `json:"start_time"`
	EndTime               float64     `json:"end_time"`
	EstimatedInputTokens  int64       `json:"estimated_input_tokens"`
	EstimatedOutputTokens int64       `json:"estimated_output_tokens"`
	EstimatedRatioTokens  float64     `json:"estimated_ratio_tokens"`
	TokenUsage            TokenUsage  `json:"token_usage"`
	// 证据注入元数据，标注文本中引用工具结果的位置
	Evidence              map[string]Evidence `json:"_evidence,omitempty"`
}

// Evidence 证据元数据，标注文本中引用工具结果的位置
type Evidence struct {
	ObjectTypeName string      `json:"object_type_name"`
	Positions      [][]int     `json:"positions"`
	ToolName       string      `json:"tool_name,omitempty"`
	Result         interface{} `json:"result,omitempty"`
	ResultType     string      `json:"result_type,omitempty"`
}

type TokenUsage struct {
	PromptTokens       int64              `json:"prompt_tokens"`
	CompletionTokens   int64              `json:"completion_tokens"`
	TotalTokens        int64              `json:"total_tokens"`
	PromptTokenDetails PromptTokenDetails `json:"prompt_tokens_details"`
}
type PromptTokenDetails struct {
	CachedTokens   int64 `json:"cached_tokens"`
	UncachedTokens int64 `json:"uncached_tokens"`
}
type SkillInfo struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Args    []Arg  `json:"args"`
	Checked bool   `json:"checked"`
}

type Arg struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}
