package agentrespvo

type AnswerExplore struct {
	AgentName   string      `json:"agent_name"`
	Answer      interface{} `json:"answer"`
	Think       string      `json:"think"`
	Status      string      `json:"status"`
	Interrupted bool        `json:"interrupted"`
	// BlockAnswer string `json:"block_answer"`
	InputMessage interface{} `json:"input_message"`
}
