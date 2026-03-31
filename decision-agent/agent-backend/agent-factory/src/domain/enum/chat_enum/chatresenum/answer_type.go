package chatresenum

type AnswerType = OutputVarType

const (
	AnswerTypeExplore AnswerType = OutputVarTypeExplore // explore
	AnswerTypePrompt  AnswerType = OutputVarTypePrompt  // prompt，格式如：{"answer":"xxx","think":"xxx"}
	AnswerTypeOther   AnswerType = OutputVarTypeOther   // other（除了 prompt 和 explore以外的）
)
