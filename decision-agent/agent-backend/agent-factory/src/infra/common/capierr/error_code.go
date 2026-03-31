package capierr

// data agent config error code
const (
	DataAgentConfigLlmRequired                               = "AgentFactory.DataAgentConfig.BadRequest.LlmRequired"                               // llm required
	DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize = "AgentFactory.DataAgentConfig.BadRequest.RetrieverDataSourceKnEntryExceedLimitSize" // kn_entry exceeds limit size
)
