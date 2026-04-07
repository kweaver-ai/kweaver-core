package v2agentexecutordto

// InterruptHandle 恢复句柄
type InterruptHandle struct {
	FrameID       string `json:"frame_id"`       // 执行帧ID
	SnapshotID    string `json:"snapshot_id"`    // 快照ID
	ResumeToken   string `json:"resume_token"`   // 恢复令牌
	InterruptType string `json:"interrupt_type"` // 中断类型
	CurrentBlock  int    `json:"current_block"`  // 当前代码块索引
	RestartBlock  bool   `json:"restart_block"`  // 是否重启代码块
}
