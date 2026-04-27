package agentresperr

type RespErrorType string

const (
	RespErrorTypeAgentFactory  RespErrorType = "agent_factory"  // 来自agent-factory的错误
	RespErrorTypeAgentExecutor RespErrorType = "agent_executor" // 来自agent-executor的错误
)

type RespError struct {
	Type  RespErrorType `json:"type"`
	Error interface{}   `json:"error"`
}

func NewRespError(t RespErrorType, err interface{}) *RespError {
	return &RespError{
		Type:  t,
		Error: err,
	}
}
