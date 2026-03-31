package agentrespvo

type AnswerPrompt struct {
	Answer string `json:"answer"`
	Think  string `json:"think"`
}

type AnswerPromptText struct {
	Text string `json:"text"`
}
