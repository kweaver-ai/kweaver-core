package interfaces

// 管道状态
const (
	PipelineStatus_Error   = "error"
	PipelineStatus_Running = "running"
	PipelineStatus_Close   = "close"
	PipelineStatus_Closing = "closing"
)

const (
	TaskStatus_Error    = "error"
	TaskStatus_Running  = "running"
	TaskStatus_Close    = "close"
	TaskStatus_Closing  = "closing"
	TaskStatus_Stopping = "stopping"
	TaskStatus_Stopped  = "stopped"
)

const (
	TopicOutputName = "%s.mdl.process.%s"
)
